package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/log/v2"
	"github.com/alecthomas/kong"
	"github.com/fsnotify/fsnotify"
	"github.com/veryshyjelly/cocom/app"
	"github.com/veryshyjelly/cocom/templates"
)

// CLI represents the command-line interface arguments and flags for the application.
type CLI struct {
	Config  string `help:"Path to the configuration file." default:"./cocom.yml" short:"c"`
	Root    string `help:"Project root directory." default:"." short:"C" type:"existingdir"`
	Version bool   `help:"Show version information." short:"v"`
	Debug   bool   `help:"Show debug information." short:"d"`
}

func main() {
	var cli CLI
	kong.Parse(&cli,
		kong.Name("cocom"),
		kong.Description("Competitive programming companion."),
	)

	if cli.Version {
		fmt.Println("cocom version:0.0.1")
		return
	}

	// Initialize comprehensive logging
	logPath := filepath.Join(os.TempDir(), "cocom.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	if cli.Debug {
		log.SetLevel(log.DebugLevel)
		log.Debug("Debug logging enabled", "path", logPath)
	} else {
		log.SetLevel(log.InfoLevel)
		log.Info("Info logging enabled", "path", logPath)
	}

	config, err := app.ReadConfig(cli.Config)
	if os.IsNotExist(err) {
		log.Info("config file not found, prompting user for template")
		var language string
		err := huh.NewSelect[string]().
			Title("Pick a language (doesn't matter just pick one).").
			Options(
				huh.NewOption[string]("Rust", "rust"),
				huh.NewOption[string]("Python", "python"),
				huh.NewOption[string]("Ocaml", "ocaml"),
			).Value(&language).
			Run()
		unwrap("can't get language of choice", err)

		data, err := templates.FS.ReadFile(language + ".yml")
		unwrap("failed to read config template", err)

		err = os.WriteFile(cli.Config, data, 0644)
		unwrap("failed to write config template", err)

		log.Info("config template copied, please edit it before running the program again", "path", cli.Config)
		os.Exit(0)
	} else if err != nil {
		unwrap("failed to decode config file", err)
	}
	log.Debug("successfully loaded config", "config", config)

	log.Debug("creating watcher for the directory")
	w, err := fsnotify.NewWatcher()
	unwrap("creating a new watcher", err)
	defer w.Close()

	// initiate model and tea program, then start the fileloop
	fileChan := make(chan string, 10)
	model := app.NewModel(cli.Root, config, fileChan)
	p := tea.NewProgram(app.NewSplash(model))
	go fileLoop(w, p, cli.Root, fileChan)

	http.HandleFunc("/", app.HandleData(p))
	go func() {
		log.Info("Starting HTTP server", "addr", "127.0.0.1:27121")
		err := http.ListenAndServe("127.0.0.1:27121", nil)
		log.Fatal("http server crashed", "err", err)
	}()

	log.Info("Starting TUI")
	if _, err := p.Run(); err != nil {
		log.Fatal("TUI crashed", "err", err)
	}
}
