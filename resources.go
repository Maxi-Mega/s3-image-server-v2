package main

import (
	_ "embed"
	"io"
)

//go:embed resources/sample-config.yml
var sampleConfig []byte

func printSampleConfig(w io.Writer) error {
	_, err := w.Write(sampleConfig)
	return err //nolint:wrapcheck
}
