package app

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
	"github.com/google/shlex"
)

// createFile generates a boilerplate source code file based on the current problem's
// URL and the configured filename rules and templates.
//
// If configured, it also asynchronously launches an external editor pointing to the
// newly created file. This function is designed to be executed as a Bubble Tea command
// and always returns nil.
func (m Model) createFile() tea.Msg {
	filename := filepath.Join(m.Root, m.getFileName())
	log.Info("Attempting to create file", "filename", filename)

	defer func() {
		if m.Config.Editor != "" {
			var editor bytes.Buffer
			err := template.Must(template.New("editor").
				Funcs(funcMap).
				Parse(m.Config.Editor)).Execute(&editor, map[string]interface{}{
				"Filename": filename,
			})
			unwrap("couldn't render editor template", err)

			args, err := shlex.Split(editor.String())
			unwrap("couldn't parse editor template", err)

			log.Debug("Executing editor command", "args", args)
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = m.Root
			err = cmd.Run()
			if err != nil {
				log.Error("Failed to open editor", "err", err, "args", args)
			} else {
				log.Info("Editor closed")
			}
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
	unwrap("couldn't create file", err)
	defer file.Close()

	// if template is provided then fill the new file with template
	if m.Template.Source != "" {
		templPath := filepath.Join(m.Root, m.Template.Source)
		log.Debug("Reading template file", "path", templPath)
		templ, err := os.ReadFile(templPath)
		unwrap("couldn't open template", err)

		// render the template modifier and write directly in the file
		modifier := m.Template.Modifier
		log.Debug("Rendering template modifier")
		err = template.Must(template.New("template").
			Funcs(funcMap).
			Parse(modifier)).Execute(file, map[string]interface{}{
			"Author": m.Author,
			"Time":   time.Now().Format("2006/01/02 15:04"),
			"Url":    m.Url,
			"Code":   string(templ),
		})
		unwrap("couldn't write template", err)
	}

	log.Info("Successfully created file", "filename", filename)
	return nil
}

// getFileName determines the local filename for the current problem by matching
// the problem's URL against configured site-specific regex rules. It captures
// URL substrings and injects them into a Go template to generate the final name.
//
// Note: Calls os.Exit(1) if no matching rule is found for the current URL host.
func (m Model) getFileName() string {
	u, err := url.Parse(m.Url)
	unwrap("couldn't parse url", err)
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	log.Debug("Parsing filename", "url", m.Url, "host", host)

	// find the rule satisfying this url
	index := slices.IndexFunc(m.Rules, func(r Rule) bool { return r.Site == host })
	if index == -1 {
		log.Error("Could not find a rule matching site", "site", host, "url", m.Url)
		os.Exit(1)
	}
	log.Debug("Matched rule", "site", m.Rules[index].Site, "regex", m.Rules[index].Regex)

	// parse template for this rule
	nameTemplate := template.Must(template.New("filename").
		Funcs(funcMap).
		Parse(m.Rules[index].Template))

	// capture url parts using provided regex
	regex, err := regexp.Compile(m.Rules[index].Regex)
	unwrap("invalid regex for url parsing", err)

	captures := regex.FindStringSubmatch(m.Url)
	log.Debug("Regex captures", "captures", captures)

	var buffer bytes.Buffer
	err = nameTemplate.Execute(&buffer, map[string]interface{}{
		"Captures": captures,
		"Title":    m.Title,
	})
	unwrap("template error", err)

	generatedName := buffer.String()
	log.Debug("Generated filename", "name", generatedName)
	return generatedName
}

// getLibFiles scans configured library paths (both individual files and directories)
// and returns a map of library filenames to their raw source code contents.
//
// The scan is filtered to only include files that match the extension of the
// current problem's target filename.
func (m Model) getLibFiles() map[string]string {
	extension := filepath.Ext(m.getFileName())
	log.Debug("Fetching library files", "extension", extension)

	libFiles := make(map[string]string)
	for name, location := range m.Lib.Include {
		location = filepath.Join(m.Root, location)
		stat, err := os.Stat(location)
		unwrap("couldn't state lib file", err)

		if stat.IsDir() {
			log.Debug("Scanning library directory", "dir", location)
			dir, err := os.ReadDir(location)
			unwrap("couldn't read lib directory", err)
			for _, d := range dir {
				if !d.IsDir() && filepath.Ext(d.Name()) == extension {
					libFile, err := os.ReadFile(filepath.Join(location, d.Name()))
					unwrap("couldn't read lib file", err)
					libFiles[filepath.Base(d.Name())] = string(libFile)
					log.Debug("Loaded library file from directory", "file", d.Name())
				}
			}
		} else {
			libFile, err := os.ReadFile(location)
			unwrap("couldn't read lib file", err)
			libFiles[name] = string(libFile)
			log.Debug("Loaded library file", "name", name, "path", location)
		}
	}

	log.Info("Finished fetching library files", "count", len(libFiles))
	return libFiles
}
