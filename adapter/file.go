package adapter

import (
	"code.google.com/p/go.exp/fsnotify"
	"github.com/DonGar/go-house/options"
	"github.com/DonGar/go-house/status"
	"io/ioutil"
	"log"
	"path/filepath"
)

type fileAdapter struct {
	base
	filename string
	watcher  *fsnotify.Watcher
}

func newFileAdapter(m *Manager, base base) (a adapter, e error) {

	filename, e := base.config.GetString("status://filename")
	// Todo: if an error is present, verify it's for filename not existing.

	// The default file name is based on the name of the adapter.
	if filename == "" {
		filename = filepath.Base(base.adapterUrl) + ".json"
	}

	configDir, e := base.status.GetString(options.CONFIG_DIR)
	if e != nil {
		return nil, e
	}

	relative_name := filepath.Join(configDir, filename)
	abs_name, e := filepath.Abs(relative_name)
	if e != nil {
		return nil, e
	}

	watcher, e := fsnotify.NewWatcher()
	if e != nil {
		return nil, e
	}

	fa := &fileAdapter{base, abs_name, watcher}

	e = fa.loadFile()
	if e != nil {
		return nil, e
	}

	// Setup watch on file so we can reload if it's updated.
	go fa.Handler()
	e = fa.watcher.Watch(fa.filename)
	if e != nil {
		return nil, e
	}

	return fa, nil
}

func (a *fileAdapter) loadFile() (e error) {
	rawJson, e := ioutil.ReadFile(a.filename)
	if e != nil {
		// If we can't read the file, nil the contents.
		a.status.Set(a.adapterUrl, nil, status.UNCHECKED_REVISION)
		return e
	}

	return a.status.SetJson(a.adapterUrl, rawJson, status.UNCHECKED_REVISION)
}

func (a *fileAdapter) Handler() {
	for {
		select {
		case ev := <-a.watcher.Event:
			log.Println("event:", ev)
			e := a.loadFile()
			if e != nil {
				log.Printf("File Adapter (%s) Load Error: %v", a.adapterUrl, e.Error())
			}

		case err := <-a.watcher.Error:
			log.Printf("File Adapter (%s) Watcher Error: %v", a.adapterUrl, err)

		case <-a.StopChan:
			a.watcher.Close()
			a.StopChan <- true
			return
		}
	}
}
