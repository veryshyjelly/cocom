package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validConfig() Config {
	return Config{
		Author: "tester",
		Editor: "code {{ .Filename }}",
		Filename: Filename{
			Rules: []Rule{
				{
					Site:     "codeforces.com",
					Regex:    "(?:problemset/problem|contest)/(\\d+)(?:/problem)?/([A-Za-z0-9]+)",
					Template: "./src/bin/{{ index .Captures 1 }}-{{ index .Captures 2 }}.rs",
				},
			},
		},
		Template: Template{
			Source:   "./src/main.rs",
			Modifier: "{{ .Code }}",
		},
		Code: Code{
			Modifier: "{{ .Code }}",
		},
		Lib: Lib{
			Regex: "use\\s*{{.Name}}",
			Include: map[string]string{
				"_": "./src/lib/",
			},
		},
		Compiler: Compiler{
			Name:    "Rust",
			Source:  "main.rs",
			Compile: "rustc",
			Args:    []string{"--edition=2024", "-O", "main.rs"},
			Run:     "./main",
		},
		CreateFile: true,
		RunOnSave:  true,
	}
}

func setupValidRoot(t *testing.T) string {
	root := t.TempDir()
	mustMkdirAll(t, filepath.Join(root, "src", "bin"))
	mustMkdirAll(t, filepath.Join(root, "src", "lib"))
	mustWriteFile(t, filepath.Join(root, "src", "main.rs"), "// template\n")
	return root
}

func mustMkdirAll(t *testing.T, path string) {
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestValidateValidConfig(t *testing.T) {
	root := setupValidRoot(t)
	cfg := validConfig()

	if err := Validate(cfg, root, "./cocom.yml"); err != nil {
		t.Fatalf("expected valid config, got: %v", err)
	}
}

func TestValidateMissingTemplateOnRule(t *testing.T) {
	root := setupValidRoot(t)
	cfg := validConfig()
	cfg.Rules = append(cfg.Rules, Rule{
		Site:  "codeforces.com",
		Regex: "problem/([^/?]+)?",
	})

	err := Validate(cfg, root, "./cocom.yml")
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected ValidationError, got %T", err)
	}

	found := false
	for _, issue := range validationErr.Issues {
		if strings.Contains(issue.Path, "codeforces.com") &&
			strings.Contains(issue.Message, "missing required field 'template'") {
			found = true
			if !strings.Contains(issue.Hint, "template:") {
				t.Fatalf("expected hint with template example, got: %q", issue.Hint)
			}
		}
	}
	if !found {
		t.Fatalf("expected missing template issue for codeforces.com, got: %+v", validationErr.Issues)
	}
}

func TestValidateInvalidRegex(t *testing.T) {
	root := setupValidRoot(t)
	cfg := validConfig()
	cfg.Rules[0].Regex = "(unclosed"

	err := Validate(cfg, root, "./cocom.yml")
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(*ValidationError)
	found := false
	for _, issue := range validationErr.Issues {
		if strings.Contains(issue.Path, "codeforces.com") &&
			strings.Contains(issue.Message, "invalid regex") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected invalid regex issue, got: %+v", validationErr.Issues)
	}
}

func TestValidateTemplateEndingWithSeparator(t *testing.T) {
	root := setupValidRoot(t)
	cfg := validConfig()
	cfg.Rules[0].Template = "./src/bin/"

	err := Validate(cfg, root, "./cocom.yml")
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(*ValidationError)
	found := false
	for _, issue := range validationErr.Issues {
		if strings.Contains(issue.Message, "directory path") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected directory path issue, got: %+v", validationErr.Issues)
	}
}

func TestValidateTemplateRendersToEmptyString(t *testing.T) {
	root := setupValidRoot(t)
	cfg := validConfig()
	cfg.Rules[0].Template = "{{ if false }}file.rs{{ end }}"

	err := Validate(cfg, root, "./cocom.yml")
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(*ValidationError)
	found := false
	for _, issue := range validationErr.Issues {
		if strings.Contains(issue.Message, "empty path") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected empty path issue, got: %+v", validationErr.Issues)
	}
}

func TestValidateMissingCompilerRun(t *testing.T) {
	root := setupValidRoot(t)
	cfg := validConfig()
	cfg.Compiler.Run = ""

	err := Validate(cfg, root, "./cocom.yml")
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(*ValidationError)
	found := false
	for _, issue := range validationErr.Issues {
		if issue.Path == "compiler.run" &&
			strings.Contains(issue.Message, "missing required field 'run'") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected compiler.run issue, got: %+v", validationErr.Issues)
	}
}

func TestValidateMissingTemplateSourceFile(t *testing.T) {
	root := setupValidRoot(t)
	cfg := validConfig()
	cfg.Template.Source = "./src/missing.rs"

	err := Validate(cfg, root, "./cocom.yml")
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(*ValidationError)
	found := false
	for _, issue := range validationErr.Issues {
		if issue.Path == "template.source" &&
			strings.Contains(issue.Message, "does not exist") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected template.source issue, got: %+v", validationErr.Issues)
	}
}

func TestValidateMissingLibIncludePath(t *testing.T) {
	root := setupValidRoot(t)
	cfg := validConfig()
	cfg.Lib.Include["missing"] = "./src/lib/missing.rs"

	err := Validate(cfg, root, "./cocom.yml")
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(*ValidationError)
	found := false
	for _, issue := range validationErr.Issues {
		if issue.Path == "lib.include.missing" &&
			strings.Contains(issue.Message, "does not exist") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected lib.include issue, got: %+v", validationErr.Issues)
	}
}

func TestValidateTemplateRendersToExistingDirectory(t *testing.T) {
	root := setupValidRoot(t)
	mustMkdirAll(t, filepath.Join(root, "problem.rs"))
	cfg := validConfig()
	cfg.Rules[0].Template = "./problem.rs"

	err := Validate(cfg, root, "./cocom.yml")
	if err == nil {
		t.Fatal("expected validation error")
	}

	validationErr := err.(*ValidationError)
	found := false
	for _, issue := range validationErr.Issues {
		if strings.Contains(issue.Message, "existing directory") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected existing directory issue, got: %+v", validationErr.Issues)
	}
}

func TestValidationErrorFormatting(t *testing.T) {
	err := &ValidationError{
		ConfigPath: "./cocom.yml",
		Issues: []Issue{
			{
				Path:    "filename.rules[5] (site: codeforces.com)",
				Message: "missing required field 'template'",
				Hint:    "template: './src/bin/{{ index .Captures 1 }}.rs'",
			},
		},
	}

	output := err.Error()
	if !strings.Contains(output, "Error: Invalid configuration file (./cocom.yml)") {
		t.Fatalf("unexpected header: %s", output)
	}
	if !strings.Contains(output, "filename.rules[5] (site: codeforces.com)") {
		t.Fatalf("missing rule path: %s", output)
	}
	if !strings.Contains(output, configDocsURL) {
		t.Fatalf("missing docs link: %s", output)
	}
}
