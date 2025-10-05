package config

import (
	"os"
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
			MaxImagesDisplayCount:  100,
			Map: UIMap{
				TileServerURL: "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
			},
		},
		Products: Products{
			DefaultPreviewSuffix: "preview.jpg",
			GeonamesFilename:     "geonames.json",
			LocalizationFilename: "localization.json",
		},
		Cache: Cache{
			CacheDir:        os.TempDir(),
			RetentionPeriod: 7 * 24 * time.Hour,
		},
		Log: Log{
			LogLevel:      zerolog.LevelInfoValue,
			ColorLogs:     true,
			JSONLogFormat: false,
			HTTPTrace:     false,
		},
		Monitoring: Monitoring{
			PrometheusInstanceLabel: "s3_image_server",
		},
	}
}
