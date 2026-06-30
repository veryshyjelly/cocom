package main

import (
	"os"

	"charm.land/log/v2"
)

func Unwrap(message string, err error) {
	if err != nil {
		log.Error(message, "err", err)
		os.Exit(1)
	}
}
