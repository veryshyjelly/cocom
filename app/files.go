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
	"github.com/google/shlex"
)

func (m Model) createFile() tea.Msg {
	filename := filepath.Join(m.Root, m.getFileName())

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
			logger.Debug("executing editor", "args", args)
			unwrap("couldn't parse editor template", err)
			cmd := exec.Command(args[0], args[1:]...)
			cmd.Dir = m.Root
			err = cmd.Run()
			unwrap("couldn't open editor", err)
		}
	}()

	// if the file already exists then we ought not do anything
	_, err := os.ReadFile(filename)
	if err == nil {
		return nil
	}

	// now create the file
	file, err := os.Create(filename)
	unwrap("couldn't create file", err)
	defer file.Close()
	// if template is provided then fill the new file with template
	if m.Template.Source != "" {
		templ, err := os.ReadFile(filepath.Join(m.Root, m.Template.Source))
		unwrap("couldn't open template", err)
		// render the template modifier and write directly in the file
		modifier := m.Template.Modifier
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

	logger.Info("created file", "filename", filename)

	return nil
}

func (m Model) getFileName() string {
	u, err := url.Parse(m.Url)
	unwrap("couldn't parse url", err)
	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
	// find the rule satisfying this url
	index := slices.IndexFunc(m.Rules, func(r Rule) bool { return r.Site == host })
	if index == -1 {
		logger.Error("could not find a rule matching site", "site", m.Url)
		os.Exit(1)
	}
	// parse template for this rule
	nameTemplate := template.Must(template.New("filename").
		Funcs(funcMap).
		Parse(m.Rules[index].Template))
	// capture url parts using provided regex
	regex, err := regexp.Compile(m.Rules[index].Regex)
	unwrap("invalid regex for url parsing", err)
	captures := regex.FindStringSubmatch(m.Url)

	var buffer bytes.Buffer
	err = nameTemplate.Execute(&buffer, map[string]interface{}{
		"Captures": captures,
		"Title":    m.Title,
	})
	unwrap("template error", err)

	return buffer.String()
}

func (m Model) getLibFiles() map[string]string {
	extension := filepath.Ext(m.getFileName())
	libFiles := make(map[string]string)
	for name, location := range m.Lib.Include {
		location = filepath.Join(m.Root, location)
		stat, err := os.Stat(location)
		unwrap("couldn't state lib file", err)
		if stat.IsDir() {
			dir, err := os.ReadDir(location)
			unwrap("couldn't read lib directory", err)
			for _, d := range dir {
				if !d.IsDir() && filepath.Ext(d.Name()) == extension {
					libFile, err := os.ReadFile(filepath.Join(location, d.Name()))
					unwrap("couldn't read lib file", err)
					libFiles[filepath.Base(d.Name())] = string(libFile)
				}
			}
		} else {
			libFile, err := os.ReadFile(location)
			unwrap("couldn't read lib file", err)
			libFiles[name] = string(libFile)
		}
	}
	return libFiles
}
