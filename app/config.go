package app

type Config struct {
	Author   string `yaml:"author"`
	Editor   string `yaml:"editor"`
	Filename `yaml:"filename"`
	Template `yaml:"template"`
	Code     `yaml:"code"`
	Lib      `yaml:"lib"`
	Compiler `yaml:"compiler"`
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
	Compile string   `yaml:"compile"`
	Args    []string `yaml:"args"`
	Run     string   `yaml:"run"`
}
