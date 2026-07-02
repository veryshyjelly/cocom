package app

import (
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/fsnotify/fsnotify"
)

type FileChanged struct {
	filename string
}

func FileLoop(w *fsnotify.Watcher, p *tea.Program, root string, files chan string) {
	var (
		filename string
		timer    *time.Timer
	)

	for {
		select {
		case file := <-files:
			filename = filepath.Base(file)
			err := w.Add(filepath.Dir(filepath.Join(root, file)))
			unwrap("couldn't watch parent directory", err)
		case e := <-w.Events:
			if e.Op&fsnotify.Write == 0 {
				continue
			}
			if filename == "" || filename != filepath.Base(e.Name) {
				continue
			}
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(200*time.Millisecond, func() {
				log.Debug("File changed detected", "filename", e.Name)
				p.Send(FileChanged{filename})
			})
		case err := <-w.Errors:
			log.Error(err)
		}
	}
}
