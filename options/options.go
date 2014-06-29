package options

import (
	"github.com/DonGar/go-house/status"
	"os"
	"path/filepath"
)

type Options struct {
	status *status.Status
}

func NewOptions(s *status.Status) (options *Options, e error) {
	// TODO: parse command args and such.
	return &Options{status: s}, nil
}

func (o *Options) ConfigDir() (v string, e error) {
	// Look it up from the config. This is useful in test code
	// that manually setup status, and need a known config dir.
	v, e = o.status.GetString("status://server/config")

	if e != nil {
		execName, e := filepath.Abs(os.Args[0])
		if e != nil {
			return "", e
		}

		return filepath.Dir(execName), nil
	}

	return v, e
}

func (o *Options) StaticDir() (string, error) {
	return "/home/dgarrett/Development/go-house/static", nil
}

func (o *Options) Latitude() (float64, error) {
	return o.status.GetFloat("status://server/latitude")
}

func (o *Options) Longitude() (float64, error) {
	return o.status.GetFloat("status://server/longitude")
}

func (o *Options) DownloadDir() (string, error) {
	return o.status.GetString("status://server/downloads")
}
