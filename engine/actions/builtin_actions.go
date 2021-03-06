package actions

import (
	"fmt"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"github.com/ghthor/gowol"
	"github.com/scorredoira/email"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func RegisterStandardActions(am *Manager) {
	am.RegisterAction("set", actionSet)
	am.RegisterAction("wol", actionWol)
	am.RegisterAction("ping", actionPing)
	am.RegisterAction("fetch", actionFetch)
	am.RegisterAction("email", actionEmail)
}

func errorHandler(e error) {
	if e != nil {
		log.Println("Background action error: ", e)
	}
}

// TODO, perhaps non-existent components should error out. But that gets
// weird with wildcards.

// Implement the "set" action.
func actionSet(s *status.Status, action *status.Status) (e error) {
	componentUrl, _, e := action.GetString("status://component")
	if e != nil {
		return e
	}

	dest, _, e := action.GetString("status://dest")
	if e != nil {
		return e
	}

	value, _, e := action.Get("status://value")
	if e != nil {
		return e
	}

	// Dest isn't allow to be a sub-url.
	if strings.Contains(dest, "/") {
		return fmt.Errorf("Action has dest %s, but / not allowed.", dest)
	}

	// Lookup affected components, expanding wildcards.
	componentMatches, e := s.GetMatchingUrls(componentUrl)
	if e != nil {
		return e
	}

	var final error

	for cUrl := range componentMatches {
		destUrl := cUrl + "/" + dest
		if e = s.Set(destUrl, value, status.UNCHECKED_REVISION); e != nil {
			final = e
		}
	}

	return final
}

// Send a Wake On Lan request to a component. The component must have a "mac"
// value defined with is the components network mac address.
func actionWol(s *status.Status, action *status.Status) (e error) {
	componentUrl, _, e := action.GetString("status://component")
	if e != nil {
		return e
	}

	// Lookup affected components, expanding wildcards.
	componentMatches, e := s.GetMatchingUrls(componentUrl)
	if e != nil {
		return e
	}

	for _, cMatch := range componentMatches {
		componentStatus := status.Status{}
		componentStatus.Set("status://", cMatch.Value, 0)

		// If the host doesn't have a Mac, that's okay. Just quietly skip it.
		mac, _, e := componentStatus.GetString("status://mac")
		if e != nil {
			continue
		}

		// Send the WOL Packet out.
		go func() { errorHandler(wol.MagicWake(mac, "255.255.255.255")) }()
	}

	return nil
}

// Ping a component, and set the "up" value on component to true or false. The
// name of the component is the name to ping. The "up" value is updated in the
// background after an arbitrary delay, not right away.
func actionPing(s *status.Status, action *status.Status) (e error) {
	componentUrl, _, e := action.GetString("status://component")
	if e != nil {
		return e
	}

	// Lookup affected components, expanding wildcards.
	componentMatches, e := s.GetMatchingUrls(componentUrl)
	if e != nil {
		return e
	}

	for cUrl := range componentMatches {
		resultUrl := cUrl + "/up"

		// Extract the hostname from component URL.
		url_parts := strings.Split(cUrl, "/")
		hostname := url_parts[len(url_parts)-1]

		go func() { errorHandler(performPing(s, hostname, resultUrl)) }()
	}

	return nil
}

func performPing(s *status.Status, hostname, resultUrl string) error {

	// Shell out to perform the ping. This avoids needing root permissions.
	cmd := exec.Command("/bin/ping", "-q", "-c", "3", hostname)
	_, e := cmd.CombinedOutput()

	// If there was no error, the host is up.
	result := e == nil

	if e = s.Set(resultUrl, result, status.UNCHECKED_REVISION); e != nil {
		return e
	}

	return nil
}

// Fetch a URL. If DownloadName is not present, the result is thrown away.
// otherwise,
// Happens ansynronously.
// Does not require a component to fire.
func actionFetch(s *status.Status, action *status.Status) (e error) {
	// "url"
	// "download_name"

	url, _, e := action.GetString("status://url")
	if e != nil {
		return e
	}

	fileName := action.GetStringWithDefault("status://download_name", "")
	fileName = expandFileName(s, fileName)

	// Fetch the file, and download to fileName if fileName != ""
	_, e = performFetch(url, fileName)
	return e
}

func performFetch(url, fileName string) (contentsBuffer []byte, e error) {
	res, e := http.Get(url)
	if e != nil {
		return nil, e
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received StatusCode: %d", res.StatusCode)
	}

	contentsBuffer, e = ioutil.ReadAll(res.Body)
	if e != nil {
		return nil, e
	}

	if fileName != "" {
		e = ioutil.WriteFile(fileName, contentsBuffer, 0666)
		if e != nil {
			return nil, e
		}
	}

	log.Println("Downloaded: ", url, " to: ", fileName)
	return contentsBuffer, nil
}

type attachemetDesc struct {
	url      string
	filename string
}

func actionEmail(s *status.Status, action *status.Status) (e error) {
	//   * to - Address to send email too.
	//   * subject - Optional subject string.
	//   * body - Optional body string.
	//   * attachments - List of attachments.
	//     * url - URL to fetch and attach to email.
	//     * download_name - Name to download and attach as. Follows same rules as fetch_url:download_name.

	values, e := action.GetStrings([]string{
		"status://to", "status://subject", "status://body"})
	if e != nil {
		return e
	}

	to := values[0]
	subject := values[1]
	body := values[2]

	attachments := []attachemetDesc{}
	if attachmentsRaw, _, e := action.Get("status://attachments"); e == nil {

		// If there are attachments, convert them to a friendly format.
		attachArray, ok := attachmentsRaw.([]interface{})
		if !ok {
			return fmt.Errorf("Bad attachment syntax.")
		}

		for _, attachRaw := range attachArray {
			attachMap, ok := attachRaw.(map[string]interface{})
			if !ok {
				return fmt.Errorf("Bad attachment syntax.")
			}

			url, ok := attachMap["url"].(string)
			if !ok {
				return fmt.Errorf("Bad attachment syntax.")
			}

			filename, ok := attachMap["download_name"].(string)
			if !ok {
				return fmt.Errorf("Bad attachment syntax.")
			}

			filename = expandFileName(s, filename)
			attachments = append(attachments, attachemetDesc{url, filename})
		}
	}

	//
	// We also depend on a number of server configuration values to actually
	// send the email.
	//

	values, e = s.GetStrings([]string{
		"status://server/email_address",
		"status://server/relay_server",
		"status://server/relay_user",
		"status://server/relay_password",
		"status://server/relay_id_server"})
	if e != nil {
		return e
	}

	from := values[0]
	relayServer := values[1]
	relayUser := values[2]
	relayPassword := values[3]
	relayIdServer := values[4]

	attachementFiles, e := fetchAttachments(attachments)
	if e != nil {
		// If we got an error, log and add it to the message body.
		log.Println("Error fetching attachements: ", e)
		body = fmt.Sprintf("%s\n\nError fetching: %s", body, e.Error())
	}

	e = sendEmail(
		to, from, subject, body, attachementFiles,
		relayServer, relayUser, relayPassword, relayIdServer)

	return e
}

// Fetch all of the attachments requested for a given email. It will always
// download as many as possible. The list of files will contain all successful
// downloads, even if there were some errors. The error (if any) will describe
// all failures.
func fetchAttachments(attachments []attachemetDesc) (files []string, e error) {

	files = []string{}
	var collectedErrors []string

	for _, attach := range attachments {
		_, e = performFetch(attach.url, attach.filename)
		if e == nil {
			files = append(files, attach.filename)
		} else {
			collectedErrors = append(collectedErrors, fmt.Sprintf(
				"%s -> %s: %s",
				attach.url, attach.filename, e.Error()))
		}
	}

	e = nil
	if collectedErrors != nil {
		e = fmt.Errorf("%s", strings.Join(collectedErrors, "\n"))
	}

	return files, e
}

// Handle actually sending out an email.
func sendEmail(
	to, from, subject, body string, attachments []string,
	relayServer, relayUser, relayPassword, relayIdServer string) error {

	m := email.NewMessage(subject, body)
	m.From = from
	m.To = strings.Split(to, ",")

	for i := range m.To {
		m.To[i] = strings.TrimSpace(m.To[i])
	}

	for _, attachment := range attachments {
		if e := m.Attach(attachment); e != nil {
			return e
		}
	}

	log.Printf("Sending Email '%s' to '%s'", subject, to)
	return email.Send(relayServer,
		smtp.PlainAuth("", relayUser, relayPassword, relayIdServer),
		m)
}

// This is used by both the fetch and email actions to handle filenames for
// downloaded content.
func expandFileName(s *status.Status, fileName string) string {
	if fileName == "" {
		return fileName
	}

	// Expand {time placeholder}
	now := time.Now()
	nowUnix := fmt.Sprintf("%d", now.Unix())
	fileName = strings.Replace(fileName, "{time}", nowUnix, -1)

	// Append downloads directory.
	if !filepath.IsAbs(fileName) {
		downloadsDir := s.GetStringWithDefault(options.DOWNLOADS_DIR, "")
		fileName = filepath.Join(downloadsDir, fileName)
	}

	return fileName
}
