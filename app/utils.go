package app

import (
	"bytes"
	"net/url"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/log/v2"
	"github.com/ettle/strcase"

	"regexp"
	"text/template"
)

var funcMap = template.FuncMap{
	"toKebabCase":  strcase.ToKebab,
	"toCamelCase":  strcase.ToCamel,
	"toSnakeCase":  strcase.ToSnake,
	"toPascalCase": strcase.ToPascal,
	"toKEBABCase":  strcase.ToKEBAB,
	"toSNAKECase":  strcase.ToSNAKE,
	"stripPrefix":  stripPrefix,
}

func (m Model) CreateFile() tea.Msg {
	u, err := url.Parse(m.Url)
	if err != nil {
		return err
	}

	host := strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")

	index := -1
	for i, rule := range m.Rules {
		if rule.Site == host {
			index = i
			break
		}
	}

	if index == -1 {
		log.Error("Could not find a rule matching site", "site", m.Url)
		return nil
	}

	nameTemplate := template.Must(template.New("filename").
		Funcs(funcMap).
		Parse(m.Rules[index].Regex))

	regex, err := regexp.Compile(m.Regex)
	unwrap("invalid regex for url parsing", err)
	captures := regex.FindStringSubmatch(m.Url)

	var buffer bytes.Buffer
	err = nameTemplate.Execute(&buffer, map[string]interface{}{
		"Captures": captures,
		"Title":    m.Title,
	})
	unwrap("template error", err)

	return nil
}

func stripPrefix(title, delim string) string {
	parts := strings.SplitN(title, delim, 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return title
}

func unwrap(message string, err error) {
	if err != nil {
		log.Error(message, "err", err)
		os.Exit(1)
	}
}
