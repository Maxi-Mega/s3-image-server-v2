package web

import (
	"errors"
	"fmt"
	"os"
)

const maxLogoFileSize = 10 << 20 // 10MiB

var (
	errLogoIsEmpty  = errors.New("logo file is empty")
	errLogoTooLarge = errors.New("logo file is too large")
)

func getLogoBase64(logoBase64Path string) (string, error) {
	if logoBase64Path == "" {
		return "", nil
	}

	logoBase64, err := os.ReadFile(logoBase64Path)
	if err != nil {
		return "", fmt.Errorf("can't read logo base64 file: %w", err)
	}

	if len(logoBase64) == 0 {
		return "", errLogoIsEmpty
	}

	if len(logoBase64) > maxLogoFileSize {
		return "", errLogoTooLarge
	}

	return string(logoBase64), nil
}
