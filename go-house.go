package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	// "github.com/cpucycle/astrotime"
	"log"
	"net/http"
	// "time"
	"github.com/DonGar/go-house/status"
)

func loadServerConfig(options Options, s *status.Status) (e error) {
	configFile := filepath.Join(options.ConfigDir, "server.json")

	rawJson, e := ioutil.ReadFile(configFile)
	if e != nil {
		return e
	}

	e = s.SetJson("status://server", rawJson, status.UNCHECKED_REVISION)
	if e != nil {
		return e
	}

	// Success!
	return nil
}

type Options struct {
	ConfigDir string
	StaticDir string
}

func findOptions() (options Options, e error) {
	execName, e := filepath.Abs(os.Args[0])
	if e != nil {
		return
	}

	// TODO: parse command args and make this configurable.
	options.ConfigDir = filepath.Dir(execName)
	options.StaticDir = "/home/dgarrett/Development/go-house/static"

	return
}

func main() {
	log.Println("Starting up.")

	options, e := findOptions()
	if e != nil {
		return
	}

	status := status.Status{}

	// Load the initial config.
	e = loadServerConfig(options, &status)
	if e != nil {
		return
	}

	// TODO: Start up the rules engine.
	// TODO: Start up the adapters

	handler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
	}

	log.Println("Starting web server.")
	http.HandleFunc("/", handler)
	http.Handle("/static/", http.StripPrefix("/static/",
		http.FileServer(http.Dir(options.StaticDir))))
	http.ListenAndServe(":8082", nil)
}
