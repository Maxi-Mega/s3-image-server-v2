package utils //nolint: revive,nolintlint

import (
	"fmt"
	"image"
	_ "image/jpeg" // Registering JPEG image format.
	_ "image/png"  // Register PNG image format.
	"os"
	"path/filepath"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
)

func GetImageSize(imagePath, cacheDir string) (types.ImageSize, error) {
	file, err := os.Open(filepath.Join(cacheDir, imagePath))
	if err != nil {
		return types.ImageSize{}, fmt.Errorf("opening image: %w", err)
	}

	defer file.Close()

	cfg, _, err := image.DecodeConfig(file)
	if err != nil {
		return types.ImageSize{}, fmt.Errorf("reading image metadata: %w", err)
	}

	return types.ImageSize{
		Width:  cfg.Width,
		Height: cfg.Height,
	}, nil
}
