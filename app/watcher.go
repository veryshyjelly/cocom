package app

import (
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/fsnotify/fsnotify"
)

type FileChanged struct {
	filename string
}

func FileLoop(w *fsnotify.Watcher, p *tea.Program, files chan string) {
	filename := ""
	for {
		select {
		case err, ok := <-w.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			log.Errorf("ERROR: %s", err)
		case file := <-files:
			// new file has come
			log.Debug("New file registered", "filename", file)
			filename = file
		// Read from Events.
		case e := <-w.Events:
			log.Debug("File changed", "filename", e.Name)
			// Ignore files we're not interested in.
			if filename == "" || filepath.Clean(filename) != filepath.Clean(e.Name) {
				continue
			}
			p.Send(FileChanged{filename})
		}
	}
}
