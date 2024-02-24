package config

import "time"

type (
	Config struct {
		S3       S3       `yaml:"s3"`
		UI       UI       `yaml:"ui"`
		Products Products `yaml:"products"`
		Cache    Cache    `yaml:"cache"`
		Log      Log      `yaml:"log"`
	}

	S3 struct {
		PollingMode   bool          `yaml:"pollingMode"`
		PollingPeriod time.Duration `yaml:"pollingPeriod"`
		ExitOnS3Error bool          `yaml:"exitOnS3Error"`
		EndPoint      string        `yaml:"endPoint"`
		AccessID      string        `yaml:"accessID"`
		AccessSecret  string        `yaml:"accessSecret"`
		UseSSL        bool          `yaml:"useSSL"`
	}

	UI struct {
		WebServerPort          uint16 `yaml:"webServerPort"`
		BasePath               string `yaml:"basePath"`
		WindowTitle            string `yaml:"windowTitle"`
		ApplicationTitle       string `yaml:"applicationTitle"`
		LogoBase64Path         string `yaml:"logoBase64Path"`
		ScaleInitialPercentage uint   `yaml:"scaleInitialPercentage"`
		MaxImagesDisplayCount  uint   `yaml:"maxImagesDisplayCount"`
		Map                    UIMap  `yaml:"map"`
	}

	Products struct {
		PreviewFilename              string       `yaml:"previewFilename"`
		Geonames                     string       `yaml:"geonames"`
		LocalizationFilename         string       `yaml:"localizationFilename"`
		AdditionalProductFilesRegexp string       `yaml:"additionalProductFilesRegexp"`
		FeaturesExtentionRegexp      string       `yaml:"featuresExtentionRegexp"`
		FeaturesCategoryName         string       `yaml:"featuresCategoryName"`
		FeaturesClassName            string       `yaml:"featuresClassName"`
		FullProductExtension         string       `yaml:"fullProductExtension"`
		FullProductSignedURL         string       `yaml:"fullProductSignedURL"`
		ImageGroups                  []ImageGroup `yaml:"imageGroups"`
	}

	Cache struct {
		CacheDir        string        `yaml:"cacheDir"`
		RetentionPeriod time.Duration `yaml:"retentionPeriod"`
	}

	Log struct {
		LogLevel      string         `yaml:"logLevel"`
		ColorLogs     bool           `yaml:"colorLogs"`
		JSONLogFormat bool           `yaml:"JSONLogFormat"`
		JSONLogFields map[string]any `yaml:"JSONLogFields"`
		HTTPTrace     bool           `yaml:"HTTPTrace"`
	}

	UIMap struct {
		TileServerURL string `yaml:"tileServerURL"`
	}

	ImageGroup struct {
		GroupName string      `yaml:"groupName"`
		Types     []ImageType `yaml:"types"`
	}

	ImageType struct {
		Name          string `yaml:"name"`
		DisplayName   string `yaml:"displayName"`
		ProductPrefix string `yaml:"productPrefix"`
		ProductRegexp string `yaml:"productRegexp"`
	}
)
