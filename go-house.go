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
)

func mainWork() error {
	log.Println("Starting up.")

	status := &status.Status{}

	// Load the initial config.
	e := options.LoadServerConfig(status)
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

	e = server.RunHttpServerForever(status, adapterMgr)
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
