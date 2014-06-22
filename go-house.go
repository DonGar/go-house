package main

import (
	// "github.com/cpucycle/astrotime"
	"log"
	// "time"
	"github.com/DonGar/go-house/adapter"
	"github.com/DonGar/go-house/http-server"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"io/ioutil"
	"path/filepath"
)

// Load the initial server config into our status struct.
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

	status := &status.Status{}

	// Load the initial config.
	e = loadServerConfig(options, status)
	if e != nil {
		panic(e)
	}

	// Start the ActionManager
	//rulesMgr, e := rules.NewManager(options, status)

	// Start the AdapterManager.
	adapterMgr, e := adapter.NewManager(options, status)

	// TODO: Start up the rules engine.

	server.RunHttpServerForever(options, status, adapterMgr)
}
