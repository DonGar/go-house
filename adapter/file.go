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
	stop     chan bool
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

	fa := &fileAdapter{base, abs_name, watcher, make(chan bool)}

	e = fa.loadFile()
	if e != nil {
		return nil, e
	}

	// Setup watch on file so we can reload if it's updated.
	go fa.watchForUpdates()
	e = fa.watcher.Watch(fa.filename)
	if e != nil {
		return nil, e
	}

	return fa, nil
}

func (fa *fileAdapter) loadFile() (e error) {
	rawJson, e := ioutil.ReadFile(fa.filename)
	if e != nil {
		// If we can't read the file, nil the contents.
		fa.status.Set(fa.adapterUrl, nil, status.UNCHECKED_REVISION)
		return e
	}

	return fa.status.SetJson(fa.adapterUrl, rawJson, status.UNCHECKED_REVISION)
}

// Remove this adapter from the web URLs section, the default Stop.
func (fa *fileAdapter) Stop() (e error) {

	fa.stop <- true
	<-fa.stop
	return fa.base.Stop()
}

func (fa *fileAdapter) watchForUpdates() {
	for {
		select {
		case ev := <-fa.watcher.Event:
			log.Println("event:", ev)
			e := fa.loadFile()
			if e != nil {
				log.Printf("File Adapter (%s) Load Error: %s", fa.adapterUrl, e)
			}

		case err := <-fa.watcher.Error:
			log.Printf("File Adapter (%s) Watcher Error: %s", fa.adapterUrl, err)

		case <-fa.stop:
			fa.watcher.Close()
			fa.stop <- true
			return
		}
	}
}
