package main

import (
	// "github.com/cpucycle/astrotime"
	"log"
	// "time"
	"github.com/DonGar/go-house/http-server"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"io/ioutil"
	"path/filepath"
)

func loadServerConfig(options *options.Options, s *status.Status) (e error) {
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

func main() {
	log.Println("Starting up.")

	options, e := options.FindOptions()
	if e != nil {
		panic(e)
	}

	status := status.Status{}

	// Load the initial config.
	e = loadServerConfig(options, &status)
	if e != nil {
		panic(e)
	}

	// TODO: Start up the rules engine.
	// TODO: Start up the adapters

	server.RunHttpServerForever(options, &status)
}
