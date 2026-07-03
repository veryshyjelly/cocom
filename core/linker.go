package core

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"time"

	"charm.land/log/v2"
	"github.com/samber/lo"
)

type Library struct {
	Name    string
	Content string
}

type node struct {
	name    string
	content string
	deps    []string
}

// getSolution orchestrates the final code generation process for submission.
// It reads the main source file, resolves and topologically sorts library dependencies,
// extracts and deduplicates header blocks (e.g., #includes), and merges everything
// into a single, deployable source code string using a configured template modifier.
func (app App) getSolution() string {
	log.Debug("Generating final solution string")
	code, err := os.ReadFile(filepath.Join(app.Root, app.GetFileName()))
	Unwrap("couldn't read file in getSolution", err)

	// get the lib files in topo sorted order
	libFiles := app.linkFiles()
	log.Debug("Extracting code blocks from libraries", "count", len(libFiles))
	libCode := lo.Map(libFiles, // extract out the code blocks
		func(item Library, _ int) Library {
			return Library{
				Name:    item.Name,
				Content: extractCodeBlock(item.Content),
			}
		})

	// extract out headers from each lib file
	log.Debug("Extracting header blocks from libraries")
	libHeader := lo.Map(libFiles, func(item Library, _ int) string {
		return extractHeaderBlock(item.Content)
	})

	codeHeader := slices.Collect(strings.Lines(extractHeaderBlock(string(code))))

	// merge and dedup the headers
	headers := strings.Join(lo.Uniq(slices.Concat(libHeader, codeHeader)), "\n")
	log.Debug("Merged and deduplicated headers", "length", len(headers))

	var solution bytes.Buffer
	log.Debug("Executing solution template modifier")
	err = template.Must(template.New("template").
		Funcs(funcMap).Parse(app.Code.Modifier)).
		Execute(&solution, map[string]interface{}{
			"Author":   app.Author,
			"Url":      app.Url,
			"Time":     time.Now().Format("2006/01/02 15:04"),
			"LibFiles": libCode,
			"Header":   headers,
			"Code":     extractCodeBlock(string(code)),
		})
	Unwrap("couldn't execute template on solution", err)

	log.Info("Successfully generated solution string", "size_bytes", solution.Len())
	return solution.String()
}

// linkFiles performs a depth-first topological sort on the project's library files
// to resolve inclusion dependencies. It builds a dependency graph by matching
// configured regex patterns against file contents.
//
// Detects cyclic dependencies and fatally exits if one is found. Returns an ordered
// slice of libraries ensuring that dependencies are always declared before their dependents.
func (app App) linkFiles() []Library {
	log.Debug("Starting library dependency linking")
	rootFile, err := os.ReadFile(filepath.Join(app.Root, app.GetFileName()))
	Unwrap("couldn't read file in linkFiles", err)

	nodes := lo.MapEntries(app.getLibFiles(),
		func(name, content string) (string, *node) {
			return name, &node{
				name:    name,
				content: content,
			}
		})

	regTemplate := template.Must(template.New("libReg").Funcs(funcMap).Parse(app.Lib.Regex))
	var thisRegex bytes.Buffer

	log.Debug("Building dependency graph")
	for _, n := range nodes {
		for dep := range nodes {
			thisRegex.Reset()
			err = regTemplate.Execute(&thisRegex, map[string]interface{}{"Name": dep})
			Unwrap("couldn't generate libcheck regex from template", err)
			if re := regexp.MustCompile(thisRegex.String()); re.MatchString(n.content) {
				n.deps = append(n.deps, dep)
				log.Debug("Found dependency", "node", n.name, "depends_on", dep)
			}
		}
	}

	var roots []string
	for name := range nodes {
		thisRegex.Reset()
		err = regTemplate.Execute(&thisRegex, map[string]interface{}{"Name": name})
		log.Debug("Regex for ", "node", name, "regex", thisRegex.String())
		Unwrap("couldn't generate libcheck regex from template", err)
		if re := regexp.MustCompile(thisRegex.String()); re.Match(rootFile) {
			roots = append(roots, name)
			log.Debug("Identified root dependency", "name", name)
		}
	}
	log.Info("Found library roots", "roots", roots)

	const (
		white = iota
		gray
		black
	)
	state := map[string]int{}
	var order []Library

	// topological sort
	var dfs func(string)
	dfs = func(name string) {
		if state[name] == gray {
			log.Error("Cyclic library dependency detected", "node", name)
			os.Exit(1)
		} else if state[name] == black {
			return
		}
		state[name] = gray
		log.Debug("DFS traversing", "node", name)
		for _, dep := range nodes[name].deps {
			dfs(dep)
		}
		state[name] = black
		order = append(order, Library{
			Name:    name,
			Content: nodes[name].content,
		})
	}

	for _, root := range roots {
		dfs(root)
	}

	log.Info("Successfully linked libraries", "order", lo.Map(order, func(l Library, _ int) string { return l.Name }))
	return order
}
