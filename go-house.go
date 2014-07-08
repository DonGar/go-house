package main

import (
	// "github.com/cpucycle/astrotime"
	"log"
	// "time"
	"github.com/DonGar/go-house/adapter"
	"github.com/DonGar/go-house/http-server"
	"github.com/DonGar/go-house/logging"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/rules"
	"github.com/DonGar/go-house/status"
	"io"
	"os"
)

func mainWork() error {

	// Setup a cache of all recent logs, and capture logging output in it.
	cachedLogging := &logging.CachedLogging{}
	log.SetOutput(io.MultiWriter(os.Stderr, cachedLogging))

	// Create the system wide status variable.
	status := &status.Status{}

	// Load the initial config, and parse command line arguments.
	// All settings are stored in status://server.
	e := options.IntializeServerConfig(status, os.Args)
	if e != nil {
		return e
	}

	// Start the rules manager
	rulesMgr, e := rules.NewManager(status)
	if e != nil {
		return e
	}
	defer rulesMgr.Stop()

	// Start the AdapterManager.
	adapterMgr, e := adapter.NewManager(status)
	if e != nil {
		return e
	}
	defer adapterMgr.Stop()

	// Run the web server. This normally never returns.
	return server.RunHttpServerForever(status, adapterMgr, cachedLogging)
}

func main() {
	if e := mainWork(); e != nil {
		panic(e)
	}
}
