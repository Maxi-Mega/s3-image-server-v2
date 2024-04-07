package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

func defaultConfig() Config {
	return Config{
		S3: S3{
			PollingMode:   true,
			PollingPeriod: 30 * time.Second,
		},
		UI: UI{
			BaseURL:                "/",
			WindowTitle:            "S3 Image Viewer",
			ScaleInitialPercentage: 50,
			MaxImagesDisplayCount:  10,
			Map: UIMap{
				TileServerURL: "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
			},
		},
		Products: Products{
			PreviewFilename:      "preview.jpg",
			GeonamesFilename:     "geonames.json",
			LocalizationFilename: "localization.json",
		},
		Cache: Cache{
			CacheDir: filepath.Join(os.TempDir(), defaultCacheDirName),
		},
		Log: Log{
			LogLevel:      zerolog.LevelInfoValue,
			ColorLogs:     true,
			JSONLogFormat: false,
			HTTPTrace:     false,
		},
	}
}
