package options

import (
	"os"
	"path/filepath"
)

type Options struct {
	ConfigDir string
	StaticDir string
}

func FindOptions() (options Options, e error) {
	execName, e := filepath.Abs(os.Args[0])
	if e != nil {
		return
	}

	// TODO: parse command args and make this configurable.
	options.ConfigDir = filepath.Dir(execName)
	options.StaticDir = "/home/dgarrett/Development/go-house/static"

	return
}
