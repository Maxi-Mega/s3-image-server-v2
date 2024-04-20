package web

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const maxBase64FileSize = 10 << 20 // 10MiB

var (
	errBase64IsEmpty  = errors.New("file is empty")
	errBase64TooLarge = errors.New("file is too large - should be less than 10MB")
)

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

func formatRoutes(routes gin.RoutesInfo) string {
	strRoutes := make([]string, len(routes))

	for i, route := range routes {
		lastSlash := strings.LastIndex(route.Handler, "/")
		handler := route.Handler[lastSlash+1:]
		strRoutes[i] = fmt.Sprintf("{%s %s -> %s}", route.Method, route.Path, handler)
	}

	return strings.Join(strRoutes, " ")
}
