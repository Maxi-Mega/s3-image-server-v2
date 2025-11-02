package web

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

type Error struct {
	Err error `json:"error"`
}

func (err Error) MarshalJSON() ([]byte, error) {
	if err.Err != nil {
		return fmt.Appendf(nil, `{"error": %q}`, err.Err), nil //nolint:nilerr
	}

	return []byte(`{"error": null}`), nil
}

func detectContentType(filename string, data []byte) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".json":
		return "application/json"
	default:
		return http.DetectContentType(data)
	}
}
