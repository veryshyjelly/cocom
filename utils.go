package main

import (
	"charm.land/log/v2"
)

// Unwrap is the root-level equivalent of app.unwrap. It logs a fatal error
// and exits the program if an error occurs during early startup phases,
// such as CLI parsing or initial config generation.
func Unwrap(message string, err error) {
	if err != nil {
		log.Fatal(message, "err", err)
	}
}
