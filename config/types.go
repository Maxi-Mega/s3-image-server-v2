package config

import (
	"regexp"
	"time"

	"github.com/expr-lang/expr/vm"
)

type FileSelectorKind = string

const (
	FileSelectorKindCached               FileSelectorKind = "cached"
	FileSelectorKindSignedURL            FileSelectorKind = "signedURL"
	FileSelectorKindFullProductSignedURL FileSelectorKind = "fullProductSignedURL"
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
		TargetRelativeRegexp string          `yaml:"targetRelativeRegexp"`
		TargetRelativeRgx    *regexp.Regexp  `yaml:"-"`
		FullProductProtocol  string          `yaml:"fullProductProtocol"`
		FullProductRootURL   string          `yaml:"fullProductRootURL"`
		MaxObjectsAge        time.Duration   `yaml:"maxObjectsAge"`
		DynamicData          DynamicData     `yaml:"dynamicData"`
		DynamicFilters       []DynamicFilter `yaml:"dynamicFilters"`
		ImageGroups          []ImageGroup    `yaml:"imageGroups"`
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
		PMTilesURL      string `yaml:"pmtilesURL"`
		PMTilesStyleURL string `yaml:"pmtilesStyleURL"`
	}

	DynamicData struct {
		FileSelectors       map[string]FileSelector `yaml:"fileSelectors"`
		Expressions         map[string]string       `yaml:"expressions"`
		ExpressionsPrograms map[string]*vm.Program  `yaml:"-"`
	}

	FileSelector struct {
		Regex      string           `yaml:"regex"`
		Rgx        *regexp.Regexp   `yaml:"-"`
		Kind       FileSelectorKind `yaml:"kind"`
		KindParams []string         `yaml:"-"`
		Link       bool             `yaml:"link"`
	}

	DynamicFilter struct {
		Name       string `yaml:"name"`
		Expression string `yaml:"expression"`
	}

	ImageGroup struct {
		GroupName   string      `yaml:"groupName"`
		Bucket      string      `yaml:"bucket"`
		DynamicData DynamicData `yaml:"dynamicData"`
		Types       []ImageType `yaml:"types"`
	}

	ImageType struct {
		Name          string      `yaml:"name"`
		DisplayName   string      `yaml:"displayName"`
		ProductPrefix string      `yaml:"productPrefix"`
		DynamicData   DynamicData `yaml:"dynamicData"`
	}
)
