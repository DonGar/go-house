package server

import (
	"fmt"
	"github.com/DonGar/go-house/options"
	"log"
	"net/http"
)

func monkeyHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func RunHttpServerForever(options options.Options) {
	log.Println("Starting web server.")
	http.HandleFunc("/", monkeyHandler)
	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(options.StaticDir))))
	http.ListenAndServe(":8082", nil)
}
