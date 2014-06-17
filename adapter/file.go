package adapter

import (
	"github.com/DonGar/go-house/status"
	"io/ioutil"
	"path/filepath"
)

type FileAdapter struct {
	base
	filename string
}

func NewFileAdapter(m *AdapterManager, base base) (a Adapter, e error) {

	filename, e := base.config.GetString("status://filename")
	// Todo: if an error is present, verify it's for filename not existing.

	// The default file name is based on the name of the adapter.
	if filename == "" {
		filename = filepath.Base(base.adapterUrl) + ".json"
	}

	relative_name := filepath.Join(base.options.ConfigDir, filename)
	abs_name, e := filepath.Abs(relative_name)
	if e != nil {
		return nil, e
	}

	fa := &FileAdapter{base, abs_name}

	e = fa.loadFile()
	if e != nil {
		return nil, e
	}

	// TODO: Setup watch on file so we can reload if it's updated.

	return fa, nil
}

func (fa *FileAdapter) loadFile() (e error) {
	rawJson, e := ioutil.ReadFile(fa.filename)
	if e != nil {
		return e
	}

	e = fa.status.SetJson(fa.adapterUrl, rawJson, status.UNCHECKED_REVISION)
	if e != nil {
		return e
	}

	return
}
