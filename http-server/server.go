package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/DonGar/go-house/adapter"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

//
// Define the type used to handle status requests.
//
type StatusHandler struct {
	status     *status.Status
	adapterMgr *adapter.Manager
}

// Handle Get/Post Status requests.
func (s *StatusHandler) HandleGet(
	w http.ResponseWriter, r *http.Request,
	statusUrl string, revision int) {

	// Start watching for changes to the requested URL.
	wc, e := s.status.WatchForUpdate(statusUrl)
	if e != nil {
		http.Error(w, e.Error(), http.StatusNotFound)
		return
	}
	defer s.status.ReleaseWatch(wc)

	// Fetch a close channel from http.ResponseWriter, if supported.
	var closeChannel <-chan bool
	if closeNotifier, ok := w.(http.CloseNotifier); ok {
		closeChannel = closeNotifier.CloseNotify()
	}

	for {
		select {
		case matches := <-wc:
			match, ok := matches[statusUrl]
			if !ok {
				// If our URL isn't in the matches, it doesn't exist.
				http.Error(w, "Status url not found: "+statusUrl, http.StatusNotFound)
				return
			}

			if match.Revision == revision {
				// The client already has our current revision. Block until a new
				// revision is available.
				continue
			}

			// We've found our result, convert to json.
			valueJson, e := json.Marshal(match.Value)
			if e != nil {
				// TODO: Find better result code.
				http.Error(w, e.Error(), http.StatusNotFound)
				return
			}

			// Send final result.
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"revision":%d,"status":%s}`, match.Revision, valueJson)
			return

		case <-closeChannel:
			log.Println("Connection was closed.")
			return
		}
	}
}

// Find out if the requested URL can be legally written too in this request.
func (s *StatusHandler) VerfiyStatusUrlAgainstAdapters(statusUrl string) bool {
	for _, u := range s.adapterMgr.WebAdapterStatusUrls() {
		if statusUrl == u || strings.HasPrefix(statusUrl, u+"/") {
			return true
		}
	}

	return false
}

func (s *StatusHandler) HandlePut(
	w http.ResponseWriter, r *http.Request,
	statusUrl string, revision int) {

	// TODO: Verify URL against web adapters.
	if !s.VerfiyStatusUrlAgainstAdapters(statusUrl) {
		http.Error(w, fmt.Sprintf("No adapter for %s.", statusUrl), http.StatusBadRequest)
		return
	}

	// Read the body into memory.
	body := bytes.NewBuffer(nil)
	_, e := io.CopyN(body, r.Body, 1*1024*1024) // Limit read size to 1M
	if e != io.EOF {
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	// Put it into the status tree.
	e = s.status.SetJson(statusUrl, body.Bytes(), revision)
	if e != nil {
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}
}

// Handle a Status request. This parses arguments, then hands off to Method
// specific handlers.
func (s *StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	statusUrl := "status://" + r.URL.Path[len("/status/"):]

	// Find the revision associated with the request.
	revision := status.UNCHECKED_REVISION
	revision_str := r.FormValue("revision")
	if revision_str != "" {
		var e error
		revision, e = strconv.Atoi(revision_str)
		if e != nil {
			// TODO: Produce other error codes as needed.
			http.Error(w, e.Error(), http.StatusBadRequest)
			return
		}
	}

	// Ensure the remote user only uses queries with a simple URL.
	if e := status.CheckForWildcard(statusUrl); e != nil {
		http.Error(w, e.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Handling Status %s Rev: %d - Url: '%s'\n",
		r.Method, revision, statusUrl)

	// Dispatch the request, based on the type of request.
	switch r.Method {
	case "GET", "POST":
		s.HandleGet(w, r, statusUrl, revision)
	case "PUT":
		s.HandlePut(w, r, statusUrl, revision)
	default:
		http.Error(w, fmt.Sprintf("Method %s no supported", r.Method),
			http.StatusMethodNotAllowed)
	}
}

// This method configures our HTTP Handlers, and runs the web server forever. It
// does not return.
func RunHttpServerForever(options *options.Options, status *status.Status, adapterMgr *adapter.Manager) {
	http.Handle("/", http.FileServer(http.Dir(options.StaticDir)))
	http.Handle("/status/", &StatusHandler{status: status, adapterMgr: adapterMgr})

	log.Println("Starting web server.")
	http.ListenAndServe(":8082", nil)
}
