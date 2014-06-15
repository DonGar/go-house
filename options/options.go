package options

import (
	"os"
	"path/filepath"
)

type Options struct {
	ConfigDir string
	StaticDir string
}

func FindOptions() (options *Options, e error) {
	execName, e := filepath.Abs(os.Args[0])
	if e != nil {
		return nil, e
	}

	// TODO: parse command args and make this configurable.
	options = &Options{
		ConfigDir: filepath.Dir(execName),
		StaticDir: "/home/dgarrett/Development/go-house/static",
	}

	return options, nil
}
