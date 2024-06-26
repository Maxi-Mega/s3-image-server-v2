package main

import (
	"embed"
	"fmt"
	"io"
)

//go:embed resources/sample-config.yml
var sampleConfig []byte

func printSampleConfig(w io.Writer) error {
	_, err := fmt.Fprintln(w, string(sampleConfig))
	return err //nolint:wrapcheck
}

//go:embed all:frontend/dist
var frontend embed.FS
