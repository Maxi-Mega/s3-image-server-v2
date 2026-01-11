package main

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"path/filepath"
)

//go:embed resources/sample-config.yml
var sampleConfig []byte

//go:embed resources/*_doc.md
var docs embed.FS

//go:embed all:frontend/dist
var frontend embed.FS

func printSampleConfig(w io.Writer) error {
	_, err := fmt.Fprintln(w, string(sampleConfig))
	return err //nolint:wrapcheck
}

func printDocumentation(w io.Writer) error {
	const docDir = "resources"

	docFiles, err := docs.ReadDir(docDir)
	if err != nil {
		return fmt.Errorf("bad documentation dir: %w", err)
	}

	for _, docFile := range docFiles {
		if docFile.IsDir() {
			continue
		}

		doc, err := docs.ReadFile(filepath.Join(docDir, docFile.Name()))
		if err != nil {
			return fmt.Errorf("can't read doc file: %w", err)
		}

		var docStart int

		if bytes.HasPrefix(doc, []byte("[//]: # Code generated")) {
			docStart = bytes.IndexByte(doc, '\n') + 2 // skip the first two lines (comment)
		}

		_, _ = fmt.Fprintln(w, string(doc[docStart:]))
		_, _ = fmt.Fprintln(w)
	}

	return nil
}
