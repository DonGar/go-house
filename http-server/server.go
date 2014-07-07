package server

import (
	"fmt"
	"github.com/DonGar/go-house/adapter"
	"github.com/DonGar/go-house/logging"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"log"
	"net/http"
)

// Log each incoming request before handling it.
func Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

// Call this instead of http.Error if there is a problem.
func logAndHttpError(w http.ResponseWriter, error string, code int) {
	log.Printf("%d: %s\n", code, error)
	http.Error(w, error, code)
}

// This method configures our HTTP Handlers, and runs the web server forever. It
// does not return.
func RunHttpServerForever(
	status *status.Status,
	adapterMgr *adapter.Manager,
	cachedLogging *logging.CachedLogging) error {

	staticDir, e := status.GetString(options.STATIC_DIR)
	if e != nil {
		return e
	}

	port := status.GetIntWithDefault(options.PORT, 80)

	http.Handle("/", http.FileServer(http.Dir(staticDir)))
	http.Handle("/status/", &StatusHandler{status: status, adapterMgr: adapterMgr})
	http.Handle("/log/", &LogHandler{cachedLogging})

	log.Printf("Starting web server on %d.", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), Log(http.DefaultServeMux))
	return nil
}
