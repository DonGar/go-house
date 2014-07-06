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
	"os"
)

func mainWork() error {
	log.Println("Starting up.")

	status := &status.Status{}

	// Load the initial config. All settings are stored in status://server.
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
	adapterMgr.Stop()

	// Run the web server. This normally never returns.
	return server.RunHttpServerForever(status, adapterMgr)
}

func main() {
	e := mainWork()
	if e != nil {
		panic(e)
	}
}
