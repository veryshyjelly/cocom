package app

import (
	"slices"
	"strings"
	"text/template"

	"charm.land/log/v2"
	"github.com/ettle/strcase"
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

// stripPrefix removes a prefix from a string up to the first occurrence of a
// specified delimiter, returning the trimmed remainder. Primarily used for
// cleaning up problem titles fetched from competitive programming platforms.
func stripPrefix(delim, title string) string {
	parts := strings.SplitN(title, delim, 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[1])
	}
	return title
}

// extractBlock parses a source code string to extract a specific block of text
// enclosed between `@tag begin` and `@tag end` markers.
//
// Returns the extracted block, or the default value if the tag is missing or malformed.
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

// extractHeaderBlock is a convenience wrapper around extractBlock specifically
// targeting the `@head` tag to extract include/import headers from library files.
func extractHeaderBlock(source string) string {
	return extractBlock(source, "head", "")
}

// extractCodeBlock is a convenience wrapper around extractBlock specifically
// targeting the `@code` tag to extract the main logic body of a library or solution file.
func extractCodeBlock(source string) string {
	return extractBlock(source, "code", source)
}

// unwrap is a fatal error handling utility. If the provided error is non-nil,
// it logs the error message and immediately terminates the program with exit code 1.
func unwrap(message string, err error) {
	if err != nil {
		log.Fatal(message, "err", err)
	}
}
