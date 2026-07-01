package app

import (
	"os"
	"slices"
	"strings"
	"text/template"

	"charm.land/log/v2"
	"github.com/ettle/strcase"
)

var (
	fff, _ = os.Create("/tmp/cocom.log")
	logger = log.New(fff)
)

var funcMap = template.FuncMap{
	"toKebabCase":  strcase.ToKebab,
	"toCamelCase":  strcase.ToCamel,
	"toSnakeCase":  strcase.ToSnake,
	"toPascalCase": strcase.ToPascal,
	"toKEBABCase":  strcase.ToKEBAB,
	"toSNAKECase":  strcase.ToSNAKE,
	"toLowerCase":  strings.ToLower,
	"toUpperCase":  strings.ToUpper,
	"stripPrefix":  stripPrefix,
}

func stripPrefix(delim, title string) string {
	parts := strings.SplitN(title, delim, 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return title
}

func extractBlock(source, tag string, defaultValue string) string {
	if !strings.Contains(source, "@"+tag) {
		return defaultValue
	}

	lines := strings.Split(source, "\n")
	var start int
	if start = slices.IndexFunc(lines,
		func(line string) bool {
			return strings.Contains(line, "@"+tag) &&
				strings.Contains(line, "begin")
		}); start == -1 {
		return source
	}

	end := slices.IndexFunc(lines[start+1:],
		func(line string) bool {
			return strings.Contains(line, "@"+tag) &&
				strings.Contains(line, "end")
		})

	var block []string
	if end == -1 {
		block = lines[start+1:]
	} else {
		block = lines[start+1 : start+1+end]
	}

	if len(block) == 0 {
		return source
	}

	return strings.TrimSpace(strings.Join(block, "\n"))
}

func extractHeaderBlock(source string) string {
	return extractBlock(source, "head", "")
}

func extractCodeBlock(source string) string {
	return extractBlock(source, "code", source)
}

func unwrap(message string, err error) {
	if err != nil {
		logger.Error(message, "err", err)
		fff.Close()
		os.Exit(1)
	}
}
