package main

import (
	// "github.com/cpucycle/astrotime"
	"log"
	// "time"
	"github.com/DonGar/go-house/adapter"
	"github.com/DonGar/go-house/http-server"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/rules"
	"github.com/DonGar/go-house/status"
	"io/ioutil"
	"path/filepath"
)

// Load the initial server config into our status struct.
func loadServerConfig(options *options.Options, s *status.Status) (e error) {
	cd, e := options.ConfigDir()
	if e != nil {
		return e
	}

	configFile := filepath.Join(cd, "server.json")

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

func mainWork() error {
	log.Println("Starting up.")

	status := &status.Status{}

	options, e := options.NewOptions(status)
	if e != nil {
		return e
	}

	// Load the initial config.
	e = loadServerConfig(options, status)
	if e != nil {
		return e
	}

	// Start the rules manager
	rulesMgr, e := rules.NewManager(options, status)
	if e != nil {
		return e
	}
	defer rulesMgr.Stop()

	// Start the AdapterManager.
	adapterMgr, e := adapter.NewManager(options, status)
	if e != nil {
		return e
	}
	adapterMgr.Stop()

	e = server.RunHttpServerForever(options, status, adapterMgr)
	if e != nil {
		return e
	}

	return nil
}

func main() {
	e := mainWork()
	if e != nil {
		panic(e)
	}
}
