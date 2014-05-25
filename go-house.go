package main

import (
	"fmt"
	// "github.com/cpucycle/astrotime"
	"log"
	"net/http"
	// "time"
	"github.com/DonGar/go-house/status"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func main() {

	status := status.Status{}

	log.Info("%s", status)

	http.HandleFunc("/", handler)
	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir("/home/dgarrett/Development/go-house/static"))))
	http.ListenAndServe(":8082", nil)
}
