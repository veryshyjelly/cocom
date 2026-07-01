package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"time"

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

// getSolution processes the main code, linked libraries,
// and headers to generate a final solution string.
func (m Model) getSolution() string {
	code, err := os.ReadFile(filepath.Join(m.Root, m.getFileName()))
	unwrap("couldn't read file in getSolution", err)
	// get the lib files in topo sorted order
	libFiles := m.linkFiles()
	libCode := lo.Map(libFiles, // extract out the code blocks
		func(item Library, _ int) Library {
			return Library{
				Name:    item.Name,
				Content: extractCodeBlock(item.Content),
			}
		})
	// extract out headers from each lib file
	libHeader := lo.Map(libFiles, func(item Library, _ int) string {
		return extractHeaderBlock(item.Content)
	})
	codeHeader := slices.Collect(strings.Lines(extractHeaderBlock(string(code))))
	// merge and dedup the headers
	headers := strings.Join(lo.Uniq(slices.Concat(libHeader, codeHeader)), "\n")

	var solution bytes.Buffer
	err = template.Must(template.New("template").
		Funcs(funcMap).Parse(m.Code.Modifier)).
		Execute(&solution, map[string]interface{}{
			"Author":   m.Author,
			"Url":      m.Url,
			"Time":     time.Now().Format("2006/01/02 15:04"),
			"LibFiles": libCode,
			"Header":   headers,
			"Code":     extractCodeBlock(string(code)),
		})
	unwrap("couldn't execute template on solution", err)

	return solution.String()
}

// linkFiles performs a topological sort on library dependencies to determine the correct build order.
// It reads a root file and identifies libraries referenced within it as starting points.
// Libraries are scanned to find their direct dependencies on other libraries.
// The method constructs a dependency graph and then performs a depth-first search (DFS)
// to order the libraries. It detects and reports cyclic dependencies, exiting on discovery.
// The result is a slice of Library structs, ordered such that dependencies appear before their dependents.
func (m Model) linkFiles() []Library {
	rootFile, err := os.ReadFile(filepath.Join(m.Root, m.getFileName()))
	unwrap("couldn't read file in linkFiles", err)

	nodes := lo.MapEntries(m.getLibFiles(),
		func(name, content string) (string, *node) {
			return name, &node{
				name:    name,
				content: content,
			}
		})

	for _, n := range nodes {
		for dep := range nodes {
			re := regexp.MustCompile(fmt.Sprintf(m.Lib.Regex, regexp.QuoteMeta(dep)))
			if re.MatchString(n.content) {
				n.deps = append(n.deps, dep)
			}
		}
	}

	var roots []string
	for name := range nodes {
		re := regexp.MustCompile(fmt.Sprintf(m.Lib.Regex, regexp.QuoteMeta(name)))
		if re.Match(rootFile) {
			roots = append(roots, name)
		}
	}

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
			logger.Error("cyclic library dependency")
			os.Exit(1)
		} else if state[name] == black {
			return
		}

		state[name] = gray
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

	return order
}
