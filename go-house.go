package main

import (
	"github.com/DonGar/go-house/adapter"
	"github.com/DonGar/go-house/engine"
	"github.com/DonGar/go-house/engine/actions"
	"github.com/DonGar/go-house/http-server"
	"github.com/DonGar/go-house/logging"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"io"
	"log"
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
	err := options.IntializeServerConfig(status, os.Args)
	if err != nil {
		return err
	}

	// Redirect logging to include a log file, if listed in options.
	logfileName, err := status.GetString(options.LOG_FILE)
	if err == nil {
		logfile, err := os.Create(logfileName)
		if err != nil {
			return err
		}

		// Write to the log file, along with other targets.
		log.SetOutput(io.MultiWriter(os.Stderr, cachedLogging, logfile))
	}

	// Setup syslog writting.
	//syslogWriter, err := syslog.New(syslog.LOG_NOTICE, "go-house")

	// Create the action registrar
	actionsMgr := actions.NewManager()
	actions.RegisterStandardActions(actionsMgr)

	// Start the engine (rules, properties, active reactions)
	engine, err := engine.NewEngine(status, actionsMgr)
	if err != nil {
		return err
	}
	defer engine.Stop()

	// Start the AdapterManager.
	adapterMgr, err := adapter.NewManager(status)
	if err != nil {
		return err
	}
	defer adapterMgr.Stop()

	// Run the web server. This normally never returns.
	return server.RunHttpServerForever(status, adapterMgr, cachedLogging)
}

func main() {
	if err := mainWork(); err != nil {
		log.Panic(err)
	}
}
