package core

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"charm.land/log/v2"
)

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

// Unwrap is a fatal error handling utility. If the provided error is non-nil,
// it logs the error message and immediately terminates the program with exit code 1.
func Unwrap(message string, err error) {
	if err != nil {
		_, _ = fmt.Fprint(os.Stderr,
			"\x1b[0m"+ // Reset SGR (colors/styles)
				"\x1b[?25h"+ // Show cursor
				"\x1b[?1049l"+ // Leave alternate screen
				"\x1b[?1000l"+ // Disable X10 mouse
				"\x1b[?1002l"+ // Disable button-event mouse
				"\x1b[?1003l"+ // Disable all-motion mouse
				"\x1b[?1005l"+ // Disable UTF-8 mouse
				"\x1b[?1006l"+ // Disable SGR mouse
				"\x1b[?1015l"+ // Disable urxvt mouse
				"\x1b[?2004l"+ // Disable bracketed paste
				"\x1b[?1l"+ // Normal cursor keys
				"\x1b>"+ // Normal keypad mode
				"Please see $TMPDIR/cocom.log\r\n",
		)
		log.Fatal(message, "err", err)
	}
}
