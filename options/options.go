package options

import (
	"github.com/DonGar/go-house/status"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	CONFIG_DIR    = "status://server/config"
	STATIC_DIR    = "status://server/static"
	DOWNLOADS_DIR = "status://server/downloads"
	LATITUDE      = "status://server/latitude"
	LONGITUDE     = "status://server/longitude"
	ADAPTERS      = "status://server/adapters"
)

const fixedStaticDir = "/home/dgarrett/Development/go-house/static"

// Load the initial server config into our status struct.
func LoadServerConfig(s *status.Status) (e error) {

	// Figure out where the server config files are.
	configDir, e := s.GetString(CONFIG_DIR)
	if e != nil {
		configDir, e = defaultConfigDir()
		if e != nil {
			return e
		}
	}

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

	e = setDefaults(s, configDir)
	if e != nil {
		return e
	}

	// Success!
	return nil
}

func defaultConfigDir() (v string, e error) {
	execName, e := filepath.Abs(os.Args[0])
	if e != nil {
		return "", e
	}

	return filepath.Dir(execName), nil
}

func setDefaults(s *status.Status, configDir string) error {
	defaults := map[string]interface{}{
		CONFIG_DIR:    configDir,
		STATIC_DIR:    fixedStaticDir,
		DOWNLOADS_DIR: "/tmp/Downloads",
		LATITUDE:      0.0,
		LONGITUDE:     0.0,
	}

	for u, v := range defaults {
		_, _, e := s.Get(u)

		if e != nil {
			e = s.Set(u, v, status.UNCHECKED_REVISION)
			if e != nil {
				return e
			}
		}
	}

	return nil
}
