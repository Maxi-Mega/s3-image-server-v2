package main

import (
	"embed"
	"fmt"
	"io"
)

//go:embed resources/sample-config.yml
var sampleConfig []byte

//go:embed resources/expr-doc/expr.md
var exprDoc []byte

//go:embed all:frontend/dist
var frontend embed.FS

func printSampleConfig(w io.Writer) error {
	_, err := fmt.Fprintln(w, string(sampleConfig))
	return err //nolint:wrapcheck
}

func printDocumentation(w io.Writer) error {
	_, err := fmt.Fprintln(w, string(exprDoc))
	return err //nolint:wrapcheck
}
