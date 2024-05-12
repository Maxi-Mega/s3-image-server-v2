package web

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxBase64FileSize = 10 << 20 // 10MiB

var (
	errBase64IsEmpty  = errors.New("file is empty")
	errBase64TooLarge = errors.New("file is too large - should be less than 10MB")
)

type Error struct {
	Err error `json:"error"`
}

func (err Error) MarshalJSON() ([]byte, error) {
	if err.Err != nil {
		return []byte(fmt.Sprintf(`{"error":"%s"}`, err.Err.Error())), nil //nolint:nilerr
	}

	return []byte(`{"error":null}`), nil
}

func getBase64Content(base64Path string) (string, error) {
	if base64Path == "" {
		return "", nil
	}

	logoBase64, err := os.ReadFile(base64Path)
	if err != nil {
		return "", fmt.Errorf("can't read base64 file: %w", err)
	}

	if len(logoBase64) == 0 {
		return "", errBase64IsEmpty
	}

	if len(logoBase64) > maxBase64FileSize {
		return "", errBase64TooLarge
	}

	return string(logoBase64), nil
}

func detectContentType(filename string, data []byte) string {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".json":
		return "application/json"
	default:
		return http.DetectContentType(data)
	}
}
