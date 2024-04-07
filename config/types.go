package config

import (
	"regexp"
	"time"
)

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
		Endpoint      string        `yaml:"endpoint"`
		AccessID      string        `yaml:"accessID"`
		AccessSecret  string        `yaml:"accessSecret"`
		UseSSL        bool          `yaml:"useSSL"`
	}

	UI struct {
		WebServerPort          uint16        `yaml:"webServerPort"`
		BaseURL                string        `yaml:"baseURL"`
		WindowTitle            string        `yaml:"windowTitle"`
		ApplicationTitle       string        `yaml:"applicationTitle"`
		LogoBase64Path         string        `yaml:"logoBase64Path"`
		ScaleInitialPercentage uint          `yaml:"scaleInitialPercentage"`
		MaxImagesDisplayCount  uint          `yaml:"maxImagesDisplayCount"`
		DisplayTimeOffset      time.Duration `yaml:"displayTimeOffset"`
		Map                    UIMap         `yaml:"map"`
	}

	Products struct {
		PreviewFilename              string         `yaml:"previewFilename"`
		GeonamesFilename             string         `yaml:"geonamesFilename"`
		LocalizationFilename         string         `yaml:"localizationFilename"`
		AdditionalProductFilesRegexp string         `yaml:"additionalProductFilesRegexp"`
		AdditionalProductFilesRgx    *regexp.Regexp `yaml:"-"`
		TargetRelativeRegexp         string         `yaml:"targetRelativeRegexp"`
		TargetRelativeRgx            *regexp.Regexp `yaml:"-"`
		FeaturesExtensionRegexp      string         `yaml:"featuresExtensionRegexp"`
		FeaturesExtensionRgx         *regexp.Regexp `yaml:"-"`
		FeaturesCategoryName         string         `yaml:"featuresCategoryName"`
		FeaturesClassName            string         `yaml:"featuresClassName"`
		FullProductExtension         string         `yaml:"fullProductExtension"`
		FullProductProtocol          string         `yaml:"fullProductProtocol"`
		FullProductRootURL           string         `yaml:"fullProductRootUrl"`
		FullProductSignedURL         bool           `yaml:"fullProductSignedURL"`
		MaxObjectsAge                time.Duration  `yaml:"maxObjectsAge"`
		ImageGroups                  []ImageGroup   `yaml:"imageGroups"`
	}

	Cache struct {
		CacheDir string `yaml:"cacheDir"`
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
		Bucket    string      `yaml:"bucket"`
		Types     []ImageType `yaml:"types"`
	}

	ImageType struct {
		Name          string         `yaml:"name"`
		DisplayName   string         `yaml:"displayName"`
		ProductPrefix string         `yaml:"productPrefix"`
		ProductRegexp string         `yaml:"productRegexp"`
		ProductRgx    *regexp.Regexp `yaml:"-"`
	}
)
