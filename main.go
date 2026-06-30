package main

import (
	"fmt"
	"net/http"
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/log/v2"
	"github.com/alecthomas/kong"
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

	if cli.Debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	config, err := app.ReadConfig(cli.Config)
	if os.IsNotExist(err) {
		log.Info("config file not found, choose a template")
		var language string
		err := huh.NewSelect[string]().
			Title("Pick a language (doesn't matter just pick one).").
			Options(
				huh.NewOption[string]("Rust", "rust"),
				huh.NewOption[string]("Python", "python"),
				huh.NewOption[string]("Ocaml", "ocaml"),
			).Value(&language).
			Run()
		Unwrap("can't get language of choice", err)
		data, err := templates.FS.ReadFile(language + ".yml")
		Unwrap("failed to read config template", err)

		err = os.WriteFile(cli.Config, data, 0644)
		Unwrap("failed to write config template", err)

		log.Info("config template copied, please edit it before running the program again")

		os.Exit(0)
	} else if err != nil {
		Unwrap("failed to decode config file", err)
	}

	log.Debug("got config", "config", config)

	model := app.NewModel(cli.Root, config)
	p := tea.NewProgram(model)

	http.HandleFunc("/", app.HandleData(p))
	go func() {
		log.Fatal("http server crashed", "err",
			http.ListenAndServe("127.0.0.1:27121", nil))
	}()

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
