package server

import (
	"encoding/json"
	"fmt"
	"github.com/DonGar/go-house/logging"
	"github.com/DonGar/go-house/status"
	"log"
	"net/http"
	"strconv"
)

//
// Define the type used to handle status requests.
//
type LogHandler struct {
	logs *logging.CachedLogging
}

// Handle Get/Post Status requests.
func (l *LogHandler) HandleGet(w http.ResponseWriter, r *http.Request, revision int) {

	// Start watching for changes to the requested URL.
	watch := l.logs.WatchForUpdate()
	defer l.logs.ReleaseWatch(watch)

	// Fetch a close channel from http.ResponseWriter, if supported.
	var closeChannel <-chan bool
	if closeNotifier, ok := w.(http.CloseNotifier); ok {
		closeChannel = closeNotifier.CloseNotify()
	}

	for {
		select {
		case logUpdate := <-watch:

			if revision >= logUpdate.Revision-1 && revision <= logUpdate.Revision {
				// If the requestor is 'almost current', don't update. Otherwise, we get
				// into an update loop because of the web request log for fetching logs.
				continue
			}

			// Create Json mapping.
			jsonLogs := map[string]interface{}{
				"revision": logUpdate.Revision,
				"logs":     logUpdate.Logs,
			}

			// Convert to Json string.
			valueJson, e := json.MarshalIndent(jsonLogs, "", "  ")
			if e != nil {
				logAndHttpError(w, e.Error(), http.StatusInternalServerError)
				return
			}

			// Send final result.
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, string(valueJson))
			return

		case <-closeChannel:
			log.Println("Connection was closed.")
			return
		}
	}
}

// Handle a Status request. This parses arguments, then hands off to Method
// specific handlers.
func (l *LogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Find the revision associated with the request.
	revision := status.UNCHECKED_REVISION
	revision_str := r.FormValue("revision")
	if revision_str != "" {
		var e error
		revision, e = strconv.Atoi(revision_str)
		if e != nil {
			logAndHttpError(w, e.Error(), http.StatusBadRequest)
			return
		}
	}

	// Dispatch the request, based on the type of request.
	switch r.Method {
	case "GET", "POST":
		l.HandleGet(w, r, revision)
	default:
		logAndHttpError(w, fmt.Sprintf("Method %s not supported", r.Method),
			http.StatusMethodNotAllowed)
	}
}
