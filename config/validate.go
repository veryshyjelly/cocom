package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/google/shlex"
	"github.com/veryshyjelly/cocom/tmpl"
)

const configDocsURL = "https://github.com/veryshyjelly/cocom#-configuration-cocomyml"

// Issue describes a single configuration problem.
type Issue struct {
	Path    string
	Message string
	Hint    string
}

// ValidationError collects configuration issues found during startup validation.
type ValidationError struct {
	ConfigPath string
	Issues     []Issue
}

func (e *ValidationError) Error() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Error: Invalid configuration file (%s)\n", e.ConfigPath))
	for _, issue := range e.Issues {
		b.WriteString("\n  ")
		b.WriteString(issue.Path)
		b.WriteString(": ")
		b.WriteString(issue.Message)
		if issue.Hint != "" {
			b.WriteString("\n    hint: ")
			b.WriteString(issue.Hint)
		}
	}
	b.WriteString("\n\nSee: ")
	b.WriteString(configDocsURL)
	b.WriteString("\n")
	return b.String()
}

// Validate checks that the configuration is complete and usable before startup.
func Validate(cfg Config, root string, configPath string) error {
	var issues []Issue

	issues = append(issues, validateFilenameRules(cfg, root)...)
	issues = append(issues, validateTemplateSection(cfg, root)...)
	issues = append(issues, validateCompilerSection(cfg)...)
	issues = append(issues, validateLibSection(cfg, root)...)
	issues = append(issues, validateOptionalTemplates(cfg)...)

	if len(issues) == 0 {
		return nil
	}
	return &ValidationError{
		ConfigPath: configPath,
		Issues:     issues,
	}
}

func validateFilenameRules(cfg Config, root string) []Issue {
	var issues []Issue

	if len(cfg.Rules) == 0 {
		issues = append(issues, Issue{
			Path:    "filename.rules",
			Message: "must contain at least one rule",
			Hint:    "add a rule with site, regex, and template fields",
		})
		return issues
	}

	for i, rule := range cfg.Rules {
		path := rulePath(i, rule.Site)
		if strings.TrimSpace(rule.Site) == "" {
			issues = append(issues, Issue{
				Path:    path,
				Message: "missing required field 'site'",
				Hint:    "site: codeforces.com",
			})
		}

		if strings.TrimSpace(rule.Regex) == "" {
			issues = append(issues, Issue{
				Path:    path,
				Message: "missing required field 'regex'",
				Hint:    "regex: \"problem/([^/?]+)\"",
			})
			continue
		}

		compiled, err := regexp.Compile(rule.Regex)
		if err != nil {
			issues = append(issues, Issue{
				Path:    path,
				Message: fmt.Sprintf("invalid regex: %v", err),
			})
			continue
		}

		if strings.TrimSpace(rule.Template) == "" {
			issues = append(issues, Issue{
				Path:    path,
				Message: "missing required field 'template'",
				Hint:    "template: './src/bin/{{ index .Captures 1 }}.rs'",
			})
			continue
		}

		issues = append(issues, validateFilenameTemplate(path, rule.Template, compiled, root)...)
	}

	return issues
}

func validateFilenameTemplate(path string, templateStr string, regex *regexp.Regexp, root string) []Issue {
	var issues []Issue

	tmpl, err := template.New("filename").Funcs(tmpl.FuncMap).Parse(templateStr)
	if err != nil {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("invalid template: %v", err),
		})
		return issues
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, map[string]any{
		"Captures": dummyCapturesForRegex(regex),
		"Title":    "Example Problem",
	})
	if err != nil {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("template execution failed: %v", err),
		})
		return issues
	}

	rendered := strings.TrimSpace(buffer.String())
	if rendered == "" {
		issues = append(issues, Issue{
			Path:    path,
			Message: "template renders to an empty path",
			Hint:    "include a filename with extension, e.g. './src/bin/{{ index .Captures 1 }}.rs'",
		})
		return issues
	}

	if strings.HasSuffix(rendered, "/") || strings.HasSuffix(rendered, "\\") {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("template renders to a directory path '%s'", rendered),
			Hint:    "remove trailing path separator and include a filename with extension",
		})
		return issues
	}

	base := filepath.Base(rendered)
	if base == "." || base == ".." {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("template renders to invalid path '%s'", rendered),
			Hint:    "include a filename with extension",
		})
		return issues
	}

	if filepath.Ext(base) == "" {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("template renders to path without file extension '%s'", rendered),
			Hint:    "add a file extension, e.g. '.rs' or '.cpp'",
		})
	}

	fullPath := filepath.Join(root, rendered)
	stat, err := os.Stat(fullPath)
	if err == nil && stat.IsDir() {
		issues = append(issues, Issue{
			Path:    path,
			Message: fmt.Sprintf("template renders to existing directory '%s'", rendered),
			Hint:    "use a file path instead of an existing directory",
		})
	}

	return issues
}

func validateTemplateSection(cfg Config, root string) []Issue {
	var issues []Issue

	if strings.TrimSpace(cfg.Template.Source) == "" {
		issues = append(issues, Issue{
			Path:    "template.source",
			Message: "missing required field 'source'",
			Hint:    "source: ./src/main.rs",
		})
		return issues
	}

	sourcePath := filepath.Join(root, cfg.Template.Source)
	stat, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			issues = append(issues, Issue{
				Path:    "template.source",
				Message: fmt.Sprintf("file does not exist: %s", cfg.Template.Source),
				Hint:    "create the template file or update the path",
			})
		} else {
			issues = append(issues, Issue{
				Path:    "template.source",
				Message: fmt.Sprintf("cannot access file: %v", err),
			})
		}
	} else if stat.IsDir() {
		issues = append(issues, Issue{
			Path:    "template.source",
			Message: fmt.Sprintf("path is a directory, not a file: %s", cfg.Template.Source),
			Hint:    "point to a template source file",
		})
	}

	if strings.TrimSpace(cfg.Template.Modifier) != "" {
		issues = append(issues, validateTemplateString("template.modifier", cfg.Template.Modifier)...)
	}

	return issues
}

func validateCompilerSection(cfg Config) []Issue {
	var issues []Issue

	if strings.TrimSpace(cfg.Compiler.Source) == "" {
		issues = append(issues, Issue{
			Path:    "compiler.source",
			Message: "missing required field 'source'",
			Hint:    "source: main.rs",
		})
	}

	if strings.TrimSpace(cfg.Compiler.Run) == "" {
		issues = append(issues, Issue{
			Path:    "compiler.run",
			Message: "missing required field 'run'",
			Hint:    "run: ./main",
		})
	} else if _, err := shlex.Split(cfg.Compiler.Run); err != nil {
		issues = append(issues, Issue{
			Path:    "compiler.run",
			Message: fmt.Sprintf("invalid run command: %v", err),
		})
	}

	if strings.TrimSpace(cfg.Compiler.Compile) != "" && len(cfg.Compiler.Args) == 0 {
		issues = append(issues, Issue{
			Path:    "compiler.args",
			Message: "must be non-empty when compiler.compile is set",
			Hint:    "args:\n  - --edition=2024\n  - -O\n  - main.rs",
		})
	}

	return issues
}

func validateLibSection(cfg Config, root string) []Issue {
	var issues []Issue

	if strings.TrimSpace(cfg.Lib.Regex) != "" {
		issues = append(issues, validateTemplateString("lib.regex", cfg.Lib.Regex)...)
	}

	for name, location := range cfg.Lib.Include {
		includePath := filepath.Join(root, location)
		stat, err := os.Stat(includePath)
		if err != nil {
			if os.IsNotExist(err) {
				issues = append(issues, Issue{
					Path:    fmt.Sprintf("lib.include.%s", name),
					Message: fmt.Sprintf("path does not exist: %s", location),
					Hint:    "create the library file or directory, or update the path",
				})
			} else {
				issues = append(issues, Issue{
					Path:    fmt.Sprintf("lib.include.%s", name),
					Message: fmt.Sprintf("cannot access path: %v", err),
				})
			}
			continue
		}

		if !stat.IsDir() && !stat.Mode().IsRegular() {
			issues = append(issues, Issue{
				Path:    fmt.Sprintf("lib.include.%s", name),
				Message: fmt.Sprintf("path is not a regular file or directory: %s", location),
			})
		}
	}

	return issues
}

func validateOptionalTemplates(cfg Config) []Issue {
	var issues []Issue

	if strings.TrimSpace(cfg.Editor) != "" {
		issues = append(issues, validateTemplateString("editor", cfg.Editor)...)
	}

	if strings.TrimSpace(cfg.Code.Modifier) != "" {
		issues = append(issues, validateTemplateString("code.modifier", cfg.Code.Modifier)...)
	}

	return issues
}

func validateTemplateString(path string, value string) []Issue {
	_, err := template.New(path).Funcs(tmpl.FuncMap).Parse(value)
	if err != nil {
		return []Issue{{
			Path:    path,
			Message: fmt.Sprintf("invalid template: %v", err),
		}}
	}
	return nil
}

func rulePath(index int, site string) string {
	if strings.TrimSpace(site) == "" {
		return fmt.Sprintf("filename.rules[%d]", index)
	}
	return fmt.Sprintf("filename.rules[%d] (site: %s)", index, site)
}

func dummyCapturesForRegex(regex *regexp.Regexp) []string {
	names := regex.SubexpNames()
	captures := make([]string, len(names))
	for i := range captures {
		if i == 0 {
			captures[i] = "https://example.com/problem/123"
			continue
		}
		captures[i] = fmt.Sprintf("capture%d", i)
	}
	return captures
}
