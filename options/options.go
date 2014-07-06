package options

import (
	"flag"
	"github.com/DonGar/go-house/status"
	"io/ioutil"
	"path/filepath"
)

const (
	PORT          = "status://server/port"
	CONFIG_DIR    = "status://server/config"
	STATIC_DIR    = "status://server/static"
	DOWNLOADS_DIR = "status://server/downloads"
	ADAPTERS      = "status://server/adapters"
)

// Load the initial server config into our status struct.
func IntializeServerConfig(s *status.Status, arguments []string) (e error) {

	configDir, e := parseFlags(s, arguments)
	if e != nil {
		return e
	}

	if e = loadServerConfig(s, configDir); e != nil {
		return e
	}

	if _, e := parseFlags(s, arguments); e != nil {
		return e
	}

	// Success!
	return nil
}

func defaultConfigDir(arguments []string) (v string, e error) {
	execName, e := filepath.Abs(arguments[0])
	if e != nil {
		return "", e
	}

	return filepath.Join(filepath.Dir(execName), "config"), nil
}

func loadServerConfig(s *status.Status, configDir string) error {
	// Load our main config file into status://server.
	configFile := filepath.Join(configDir, "server.json")
	rawJson, e := ioutil.ReadFile(configFile)
	if e != nil {
		return e
	}

	e = s.SetJson("status://server", rawJson, status.UNCHECKED_REVISION)
	if e != nil {
		return e
	}

	return nil
}

func parseFlags(s *status.Status, arguments []string) (configDir string, e error) {
	// Figure out where the server config files are.
	configDir, e = defaultConfigDir(arguments)
	if e != nil {
		return "", e
	}

	flagSet := flag.NewFlagSet("go-house flags", flag.ExitOnError)

	// Set our default values, and let command line arguments override them as
	// needed.

	port := s.GetIntWithDefault(PORT, 80)
	configDir = s.GetStringWithDefault(CONFIG_DIR, configDir)
	staticDir := s.GetStringWithDefault(STATIC_DIR, filepath.Join(filepath.Dir(configDir), "static"))
	downloadsDir := s.GetStringWithDefault(DOWNLOADS_DIR, "/tmp/Downloads")

	flagSet.IntVar(&port, "port", port,
		"Port number for the go-house webserver.")

	flagSet.StringVar(&configDir, "config_dir", configDir,
		"Directory that holds configuration files, especially server.json.")

	flagSet.StringVar(&staticDir, "static_dir", staticDir,
		"Directory that holds static website contents.")

	flagSet.StringVar(&downloadsDir, "downloads_dir", downloadsDir,
		"Directory in which fetched files are stored.")

	if e = flagSet.Parse(arguments[1:]); e != nil {
		return "", e
	}

	resultsSet := map[string]interface{}{
		PORT:          port,
		CONFIG_DIR:    configDir,
		STATIC_DIR:    staticDir,
		DOWNLOADS_DIR: downloadsDir,
	}

	for url, value := range resultsSet {
		if e = s.Set(url, value, status.UNCHECKED_REVISION); e != nil {
			return "", e
		}
	}

	return configDir, nil
}
