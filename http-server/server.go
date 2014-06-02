package server

import (
	"fmt"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"log"
	"net/http"
	"strconv"
)

//
// Define the type used to handle status requests.
//
type StatusHandler struct {
	status *status.Status
}

// Handle Get/Post Status requests.
func (s *StatusHandler) HandleGet(
	w http.ResponseWriter, r *http.Request,
	status_url string, revision int) {

	json_value, revision, e := s.status.GetJson(status_url)
	if e != nil {
		// TODO: Produce other error codes as needed.
		http.Error(w, e.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"revision":%d,"status":%s}`, revision, json_value)
}

// Handle a Status request. This parses arguments, then hands off to Method
// specific handlers.
func (s *StatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	status_url := "status://" + r.URL.Path[len("/status/"):]

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

	log.Printf("Handling Status %s Rev: %d - Url: '%s'\n",
		r.Method, revision, status_url)

	// Dispatch the request, based on the type of request.
	switch r.Method {
	case "GET", "POST":
		s.HandleGet(w, r, status_url, revision)
	default:
		http.Error(w, fmt.Sprintf("Method %s no supported", r.Method),
			http.StatusMethodNotAllowed)
	}
}

// This method configures our HTTP Handlers, and runs the web server forever. It
// does not return.
func RunHttpServerForever(options options.Options, status *status.Status) {
	http.Handle("/", http.FileServer(http.Dir(options.StaticDir)))
	http.Handle("/status/", &StatusHandler{status: status})

	log.Println("Starting web server.")
	http.ListenAndServe(":8082", nil)
}
