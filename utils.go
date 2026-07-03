package main

import (
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/fsnotify/fsnotify"
)

// fileLoop watches the file in the root directory, when a new file comes
// with relative path it computes the parent directory and starts watching it
func fileLoop(w *fsnotify.Watcher, p *tea.Program, root string, files chan string) {
	var (
		filename string
		timer    *time.Timer
	)

	for {
		select {
		case file := <-files:
			_ = w.Remove(filepath.Dir(filepath.Join(root, filename)))
			filename = file
			err := w.Add(filepath.Dir(filepath.Join(root, filename)))
			unwrap("couldn't watch parent directory", err)
		case e := <-w.Events:
			if e.Op&fsnotify.Write == 0 {
				continue
			}
			if filename == "" || filepath.Base(filename) != filepath.Base(e.Name) {
				continue
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(200*time.Millisecond, func() {
				log.Debug("File changed detected", "filename", e.Name)
				p.Send(e)
			})
		case err := <-w.Errors:
			log.Error(err)
		}
	}
}

// Unwrap is the root-level equivalent of app.unwrap. It logs a fatal error
// and exits the program if an error occurs during early startup phases,
// such as CLI parsing or initial config generation.
func unwrap(message string, err error) {
	if err != nil {
		log.Fatal(message, "err", err)
	}
}
