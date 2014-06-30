package actions

import (
	"fmt"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"github.com/ghthor/gowol"
	"io/ioutil"
	"log"
	"net/http"
	//"os"
	"path/filepath"
	"strings"
	"time"
)

// TODO, perhaps non-existent components should error out. But that gets
// weird with wildcards.

// Implement the "set" action.
func ActionSet(r ActionRegistrar, s *status.Status, action *status.Status) (e error) {
	componentUrl, e := action.GetString("status://component")
	if e != nil {
		return e
	}

	dest, e := action.GetString("status://dest")
	if e != nil {
		return e
	}

	value, e := action.GetString("status://value")
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

	for cUrl, _ := range componentMatches {
		destUrl := cUrl + "/" + dest
		if e = s.Set(destUrl, value, status.UNCHECKED_REVISION); e != nil {
			final = e
		}
	}

	return final
}

// Send a Wake On Lan request to a component. The component must have a "mac"
// value defined with is the components network mac address.
func ActionWol(r ActionRegistrar, s *status.Status, action *status.Status) (e error) {
	componentUrl, e := action.GetString("status://component")
	if e != nil {
		return e
	}

	// Lookup affected components, expanding wildcards.
	componentMatches, e := s.GetMatchingUrls(componentUrl)
	if e != nil {
		return e
	}

	var final error

	for _, cMatch := range componentMatches {
		componentStatus := status.Status{}
		componentStatus.Set("status://", cMatch.Value, 0)

		// If the host doesn't have a Mac, that's okay. Just quietly skip it.
		mac, e := componentStatus.GetString("status://mac")
		if e != nil {
			continue
		}

		// Send the WOL Packet out.
		e = wol.MagicWake(mac, "255.255.255.255")
		if e != nil {
			final = e
		}
	}

	return final
}

// Ping a component, and set the "up" value on component to true or false. The
// name of the component is the name to ping. The "up" value is updated in the
// background after an arbitrary delay, not right away.
func ActionPing(r ActionRegistrar, s *status.Status, action *status.Status) (e error) {
	componentUrl, e := action.GetString("status://component")
	if e != nil {
		return e
	}

	// Lookup affected components, expanding wildcards.
	componentMatches, e := s.GetMatchingUrls(componentUrl)
	if e != nil {
		return e
	}

	for cUrl, _ := range componentMatches {
		pingResultUrl := cUrl + "/up"

		// Extract the hostname from component URL.
		url_parts := strings.Split(cUrl, "/")
		hostname := url_parts[len(url_parts)-1]

		go func() {
			// TODO: Really implement this.
			// Example: https://github.com/atomaths/gtug8/blob/master/ping/ping.go
			_ = hostname
			result := false
			s.Set(pingResultUrl, result, status.UNCHECKED_REVISION)
		}()
	}

	return nil
}

// Fetch a URL. If DownloadName is not present, the result is thrown away.
// otherwise,
// Happens ansynronously.
// Does not require a component to fire.
func ActionFetch(r ActionRegistrar, s *status.Status, action *status.Status) (e error) {
	// "url"
	// "download_name"

	url, e := action.GetString("status://url")
	if e != nil {
		return e
	}

	fileName, _ := action.GetString("status://download_name")
	fileName = actionFetchExpandFileName(s, fileName)

	// Fetch the file, and download to fileName if fileName != ""
	go func() {
		res, e := http.Get(url)
		if e != nil {
			log.Println("Fire Error: ", e)
			return
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			log.Println("Fire Error: ", fmt.Errorf("Received StatusCode: %d", res.StatusCode))
			return
		}

		contentsBuffer, e := ioutil.ReadAll(res.Body)
		if e != nil {
			log.Println("Fire Error: ", e)
			return
		}

		if fileName != "" {
			e = ioutil.WriteFile(fileName, contentsBuffer, 0666)
			if e != nil {
				log.Println("Fire Error: ", e)
				return
			}
		}

		log.Println("Downloaded: ", url, " to: ", fileName)
	}()

	return nil
}

func actionFetchExpandFileName(s *status.Status, fileName string) string {
	if fileName == "" {
		return fileName
	}

	// Expand {time placeholder}
	now := time.Now()
	nowUnix := fmt.Sprintf("%d", now.Unix())
	fileName = strings.Replace(fileName, "{time}", nowUnix, -1)

	// Append downloads directory.
	if !filepath.IsAbs(fileName) {
		downloadsDir, _ := s.GetString(options.DOWNLOADS_DIR)
		fileName = filepath.Join(downloadsDir, fileName)
	}

	return fileName
}

func ActionEmail(r ActionRegistrar, s *status.Status, action *status.Status) (e error) {
	//   * to - Address to send email too.
	//   * subject - Optional subject string.
	//   * body - Optional body string.
	//   * attachments - List of attachments.
	//     * url - URL to fetch and attach to email.
	//     * download_name - Name to download and attach as. Follows same rules as fetch_url:download_name.
	//     * preserve - optional flag to keep in downloads directory.

	to, e := action.GetString("status://to")
	if e != nil {
		return e
	}

	subject, e := action.GetString("status://subject")
	if e != nil {
		return e
	}

	body, e := action.GetString("status://body")
	if e != nil {
		return e
	}

	_ = to
	_ = subject
	_ = body

	return nil
}
