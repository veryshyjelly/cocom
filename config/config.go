package config

import (
	"os"

	"github.com/goccy/go-yaml"
)

// Config represents the complete configuration for the application.
// It bundles various settings related to authors, editors, file naming, templates, code generation,
// external libraries, and compilation processes.
type Config struct {
	Author     string `yaml:"author"`
	Editor     string `yaml:"editor"`
	Filename   `yaml:"filename"`
	Template   `yaml:"template"`
	Code       `yaml:"code"`
	Lib        `yaml:"lib"`
	Compiler   `yaml:"compiler"`
	CreateFile bool `yaml:"create_file"`
	RunOnSave  bool `yaml:"run_on_save"`
}

type Filename struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	Site     string `yaml:"site"`
	Regex    string `yaml:"regex"`
	Template string `yaml:"template"`
}

type Template struct {
	Source   string `yaml:"source"`
	Modifier string `yaml:"modifier"`
}

type Code struct {
	Modifier string `yaml:"modifier"`
}

type Lib struct {
	Regex   string            `yaml:"regex"`
	Include map[string]string `yaml:"include"`
}

type Compiler struct {
	Name    string   `yaml:"name"`
	Source  string   `yaml:"source"`
	Compile string   `yaml:"compile"`
	Args    []string `yaml:"args"`
	Run     string   `yaml:"run"`
}

// ReadConfig reads and parses a YAML configuration file from the specified path.
// It decodes the file's content into a Config struct, validating the structure
// of the competitive programming environment settings.
//
// Returns the populated Config struct and an error if the file cannot be opened
// or if the YAML decoding fails.
func ReadConfig(path string) (config Config, err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return
	}

	return
}
