package config

import (
	"regexp"
	"time"
)

type FullProductURLParamType string

const (
	FullProductURLParamConstant = "constant"
	FullProductURLParamRegexp   = "regexp"
)

type (
	Config struct {
		S3         S3         `yaml:"s3"`
		UI         UI         `yaml:"ui"`
		Products   Products   `yaml:"products"`
		Cache      Cache      `yaml:"cache"`
		Log        Log        `yaml:"log"`
		Monitoring Monitoring `yaml:"monitoring"`
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
		FaviconPngBase64       string        `yaml:"faviconPngBase64"`
		LogoPngBase64          string        `yaml:"logoPngBase64"`
		ScaleInitialPercentage uint          `yaml:"scaleInitialPercentage"`
		MaxImagesDisplayCount  uint          `yaml:"maxImagesDisplayCount"`
		DisplayTimeOffset      time.Duration `yaml:"displayTimeOffset"`
		Map                    UIMap         `yaml:"map"`
	}

	Products struct {
		DefaultPreviewSuffix         string         `yaml:"defaultPreviewSuffix"`
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
		FullProductRootURL           string         `yaml:"fullProductRootURL"`
		FullProductSignedURL         bool           `yaml:"fullProductSignedURL"`
		MaxObjectsAge                time.Duration  `yaml:"maxObjectsAge"`
		ImageGroups                  []ImageGroup   `yaml:"imageGroups"`
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

	Monitoring struct {
		PrometheusInstanceLabel string `yaml:"prometheusInstanceLabel"`
	}

	UIMap struct {
		TileServerURL string `yaml:"tileServerURL"`
	}

	ImageGroup struct {
		GroupName                  string                `yaml:"groupName"`
		Bucket                     string                `yaml:"bucket"`
		FullPoductURLParams        []FullProductURLParam `yaml:"fullProductURLParams"`
		FullProductURLParamsRegexp string                `yaml:"fullProductURLParamsRegexp"`
		FullProductURLParamsRgx    *regexp.Regexp        `yaml:"-"`
		Types                      []ImageType           `yaml:"types"`
	}

	FullProductURLParam struct {
		Name         string                  `yaml:"name"`
		Type         FullProductURLParamType `yaml:"type"`
		Value        string                  `yaml:"value"`
		ValueMapping map[string]string       `yaml:"valueMapping"`
	}

	ImageType struct {
		Name          string         `yaml:"name"`
		DisplayName   string         `yaml:"displayName"`
		ProductPrefix string         `yaml:"productPrefix"`
		ProductRegexp string         `yaml:"productRegexp"`
		ProductRgx    *regexp.Regexp `yaml:"-"`
		PreviewSuffix string         `yaml:"previewSuffix"`
	}
)
