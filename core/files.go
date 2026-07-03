package core

import (
	"bytes"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/atotto/clipboard"
	"github.com/google/shlex"
	"github.com/veryshyjelly/cocom/config"
)

// CreateFile generates a boilerplate source code file based on the current problem's
// URL and the configured filename rules and templates.
//
// This function is designed to be executed as a Bubble Tea command and always returns nil.
func (app App) CreateFile() tea.Msg {
	filename := filepath.Join(app.Root, app.GetFileName())
	log.Info("Attempting to create file", "filename", filename)

	defer func() {
		if app.Config.Editor != "" {
			app.openEditor()
		}
	}()

	// if the file already exists then we ought not do anything
	_, err := os.ReadFile(filename)
	if err == nil {
		log.Debug("File already exists, skipping creation", "filename", filename)
		return nil
	}

	// now create the file
	file, err := os.Create(filename)
	Unwrap("couldn't create file", err)
	defer file.Close()

	// if template is provided then fill the new file with template
	if app.Template.Source != "" {
		templPath := filepath.Join(app.Root, app.Template.Source)
		log.Debug("Reading template file", "path", templPath)
		templ, err := os.ReadFile(templPath)
		Unwrap("couldn't open template", err)

		// render the template modifier and write directly in the file
		modifier := app.Template.Modifier
		log.Debug("Rendering template modifier")
		err = template.Must(template.New("template").
			Funcs(funcMap).
			Parse(modifier)).Execute(file, map[string]interface{}{
			"Author": app.Author,
			"Time":   time.Now().Format("2006/01/02 15:04"),
			"Url":    app.Url,
			"Code":   string(templ),
		})
		Unwrap("couldn't write template", err)
	}

	log.Info("Successfully created file", "filename", filename)
	return nil
}

// openEditor constructs and executes an external editor command to open the current problem's file.
// It renders the editor command from a template defined in the model's configuration.
// The template receives the full path to the file as "Filename".
// The command is then split into arguments and executed in the `m.Root` directory.
// Errors during template rendering, command parsing, or execution are logged.
func (app App) openEditor() {
	filename := filepath.Join(app.Root, app.GetFileName())
	var editor bytes.Buffer
	err := template.Must(template.New("editor").
		Funcs(funcMap).
		Parse(app.Config.Editor)).Execute(&editor, map[string]interface{}{
		"Filename": filename,
	})
	Unwrap("couldn't render editor template", err)

	args, err := shlex.Split(editor.String())
	Unwrap("couldn't parse editor template", err)

	log.Debug("Executing editor command", "args", args)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = app.Root
	err = cmd.Run()
	if err != nil {
		log.Error("Failed to open editor", "err", err, "args", args)
	} else {
		log.Info("Editor closed")
	}
}

// CopyFile copies the final solution to clipboard
func (app App) CopyFile() tea.Msg {
	err := clipboard.WriteAll(app.getSolution())
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// GetFileName determines the local filename for the current problem by matching
// the problem's URL against configured site-specific regex rules. It captures
// URL substrings and injects them into a Go template to generate the final name.
//
// Note: Calls os.Exit(1) if no matching rule is found for the current URL host.
func (app App) GetFileName() string {
	u, err := url.Parse(app.Url)
	Unwrap("couldn't parse url", err)
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	log.Debug("Parsing filename", "url", app.Url, "host", host)

	// find the rule satisfying this url
	index := slices.IndexFunc(app.Rules, func(r config.Rule) bool { return r.Site == host })
	if index == -1 {
		log.Error("Could not find a rule matching site", "site", host, "url", app.Url)
		os.Exit(1)
	}
	log.Debug("Matched rule", "site", app.Rules[index].Site, "regex", app.Rules[index].Regex)

	// parse template for this rule
	nameTemplate := template.Must(template.New("filename").
		Funcs(funcMap).
		Parse(app.Rules[index].Template))

	// capture url parts using provided regex
	regex, err := regexp.Compile(app.Rules[index].Regex)
	Unwrap("invalid regex for url parsing", err)

	captures := regex.FindStringSubmatch(app.Url)
	log.Debug("Regex captures", "captures", captures)

	var buffer bytes.Buffer

	cleanedTitle := strings.NewReplacer("'", "", "\"", "").Replace(app.Title)
	err = nameTemplate.Execute(&buffer, map[string]interface{}{
		"Captures": captures,
		"Title":    cleanedTitle,
	})
	Unwrap("template error", err)

	generatedName := buffer.String()
	log.Debug("Generated filename", "name", generatedName)
	return generatedName
}

// getLibFiles scans configured library paths (both individual files and directories)
// and returns a map of library filenames to their raw source code contents.
//
// The scan is filtered to only include files that match the extension of the
// current problem's target filename.
func (app App) getLibFiles() map[string]string {
	extension := filepath.Ext(app.GetFileName())
	log.Debug("Fetching library files", "extension", extension)

	libFiles := make(map[string]string)
	for name, location := range app.Lib.Include {
		location = filepath.Join(app.Root, location)
		stat, err := os.Stat(location)
		Unwrap("couldn't state lib file", err)

		if stat.IsDir() {
			log.Debug("Scanning library directory", "dir", location)
			dir, err := os.ReadDir(location)
			Unwrap("couldn't read lib directory", err)
			for _, d := range dir {
				if !d.IsDir() && filepath.Ext(d.Name()) == extension {
					libFile, err := os.ReadFile(filepath.Join(location, d.Name()))
					Unwrap("couldn't read lib file", err)
					libFiles[strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))] = string(libFile)
					log.Debug("Loaded library file from directory", "file", d.Name())
				}
			}
		} else {
			libFile, err := os.ReadFile(location)
			Unwrap("couldn't read lib file", err)
			libFiles[name] = string(libFile)
			log.Debug("Loaded library file", "name", name, "path", location)
		}
	}

	log.Info("Finished fetching library files", "count", len(libFiles))
	return libFiles
}
