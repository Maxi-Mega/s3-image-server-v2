package config

import (
	"math"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/expr-lang/expr/vm"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// ignoreType allows ignoring values of a given type,
// while still comparing the keys of the map they're values of.
func ignoreType[T any]() cmp.Option {
	return cmp.Comparer(func(_, _ T) bool {
		return true
	})
}

func TestLoad(t *testing.T) { //nolint: maintidx
	t.Parallel()

	cases := []struct {
		filePath         string
		expectedConfig   Config
		expectedWarnings []string
		expectedError    string
	}{
		{
			filePath: "valid_cfg.yml",
			expectedConfig: Config{
				S3: S3{
					Mode:          S3ModePolling,
					PollingPeriod: 30 * time.Second,
					Endpoint:      "localhost:9000",
				},
				UI: UI{
					BaseURL:                "/",
					WindowTitle:            "S3 Image Viewer",
					ScaleInitialPercentage: 50,
					MaxImagesDisplayCount:  100,
					Map: UIMap{
						PMTilesURL:      "localhost:3000/{z}/{x}/{y}.pmtiles",
						PMTilesStyleURL: "localhost:3000/protomap-styles.json",
					},
				},
				Products: Products{
					ExternalViewers: map[string]string{
						"viewer1": "localhost:8080",
					},
					DynamicData: DynamicData{
						FileSelectors: map[string]FileSelector{
							"preview": {
								Regex: "preview.jpg$",
								Kind:  FileSelectorKindCached,
								Link:  true,
							},
							"geonames": {
								Regex: "geonames.json$",
								Kind:  FileSelectorKindCached,
								Link:  true,
							},
						},
						Expressions: map[string]string{
							"key": "'value'",
						},
					},
					ImageGroups: []ImageGroup{
						{
							GroupName: "Group 1",
							DynamicData: DynamicData{
								FileSelectors: map[string]FileSelector{
									"preview": {
										Regex: "preview.jpg$",
										Kind:  FileSelectorKindCached,
										Link:  true,
									},
									"geonames": {
										Regex: "geonames.json$",
										Kind:  FileSelectorKindCached,
										Link:  true,
									},
								},
								Expressions: map[string]string{
									"key":   "'value'",
									"image": "'DIR/DIR2/IMAGE.tif'",
								},
							},
							Types: []ImageType{
								{
									Name:        "1",
									DisplayName: "One",
									DynamicData: DynamicData{
										FileSelectors: map[string]FileSelector{
											"preview": {
												Regex: "preview.jpg$",
												Kind:  FileSelectorKindCached,
												Link:  true,
											},
											"geonames": {
												Regex: "geonames.json$",
												Kind:  FileSelectorKindCached,
												Link:  true,
											},
										},
										Expressions: map[string]string{
											"key":   "'value type'",
											"image": "'DIR/DIR2/IMAGE.tif'",
										},
										ExpressionsPrograms: map[string]*vm.Program{
											"key":   nil,
											"image": nil,
										},
									},
								},
								{
									Name:        "2",
									DisplayName: "Two",
									DynamicData: DynamicData{
										FileSelectors: map[string]FileSelector{
											"preview": {
												Regex: "preview.jpg$",
												Kind:  FileSelectorKindCached,
												Link:  true,
											},
											"geonames": {
												Regex: "geonames.json$",
												Kind:  FileSelectorKindCached,
												Link:  true,
											},
											"localization": {
												Regex: "localization.json$",
												Kind:  FileSelectorKindCached,
												Link:  true,
											},
											"product": {
												Regex:      "product.tif$",
												Kind:       FileSelectorKindFullProductSignedURL,
												KindParams: []string{"urlParams"},
												Link:       true,
											},
											"ext": {
												Regex:      "ext.typ$",
												Kind:       "externalViewerURL",
												KindParams: []string{"viewer1", "extUri"},
												Link:       true,
											},
										},
										Expressions: map[string]string{
											"key":       "'value'",
											"image":     "'DIR/DIR2/IMAGE.tif'",
											"urlParams": `{"p": "val", "n": 5}`,
											"extUri":    `_s3Uri("ext")`,
										},
										ExpressionsPrograms: map[string]*vm.Program{
											"key":       nil,
											"image":     nil,
											"urlParams": nil,
											"extUri":    nil,
										},
									},
								},
							},
						},
					},
				},
				Cache: Cache{
					CacheDir:        "/tmp/s3_image_server",
					RetentionPeriod: 7 * 24 * time.Hour,
				},
				Log: Log{
					LogLevel:      "info",
					ColorLogs:     false,
					JSONLogFormat: true,
					JSONLogFields: map[string]any{
						"f1": "field",
						"f2": 7,
					},
				},
				Monitoring: Monitoring{
					PrometheusInstanceLabel: "instance",
					RequestDurationBuckets: HistogramsBuckets{
						Min:   10 * time.Millisecond,
						Max:   5 * time.Second,
						Count: 10,
					},
					S3ListDurationBuckets: HistogramsBuckets{
						Min:   100 * time.Millisecond,
						Max:   1 * time.Second,
						Count: 5,
					},
				},
			},
			expectedWarnings: []string{
				`no file selector provided for object type "localization", in type "Group 1"/"1"`,
			},
			expectedError: "",
		},
		{
			filePath: "s3_endpoint_transform.yml",
			expectedConfig: Config{
				S3: S3{
					Mode:     S3ModeEvent,
					Endpoint: "localhost:9000", // without "https://"
				},
				UI: UI{
					BaseURL:                "/",
					WindowTitle:            "S3 Image Viewer",
					ScaleInitialPercentage: 50,
					MaxImagesDisplayCount:  100,
					Map: UIMap{
						PMTilesURL: "https://tile.openstreetmap.org/{z}/{x}/{y}.png",
					},
				},
				Products: Products{
					ImageGroups: []ImageGroup{
						{
							GroupName: "Group 1",
							DynamicData: DynamicData{
								FileSelectors: map[string]FileSelector{},
								Expressions:   map[string]string{},
							},
						},
					},
				},
				Cache: Cache{
					CacheDir:        "/tmp/s3_image_server",
					RetentionPeriod: 7 * 24 * time.Hour,
				},
				Log: Log{
					LogLevel:  "info",
					ColorLogs: true,
				},
				Monitoring: Monitoring{
					PrometheusInstanceLabel: "s3_image_server",
					RequestDurationBuckets: HistogramsBuckets{
						Min:   1 * time.Millisecond,
						Max:   1 * time.Second,
						Count: 10,
					},
					S3ListDurationBuckets: HistogramsBuckets{
						Min:   100 * time.Millisecond,
						Max:   20 * time.Second,
						Count: 10,
					},
				},
			},
			expectedError: "",
		},
		{
			filePath:      "invalid_yaml.yml",
			expectedError: "failed to parse config: yaml: construct errors:\n  line 1: cannot construct !!str `invalid...` into config.Config",
		},
		{
			filePath: "invalid_cfg.yml",
			expectedWarnings: []string{
				`no file selector provided for object type "preview", in type "grp"/"typ"`,
				`no file selector provided for object type "geonames", in type "grp"/"typ"`,
				`no file selector provided for object type "localization", in type "grp"/"typ"`,
			},
			expectedError: `the config is invalid: image type name "typ" of group "grp" is duplicate`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.filePath, func(t *testing.T) {
			t.Parallel()

			cfg, warnings, err := Load("./testdata/" + tc.filePath)
			if err != nil {
				if tc.expectedError == "" {
					t.Fatalf("Expected no error, but got %q.", err.Error())
				} else if err.Error() != tc.expectedError {
					t.Fatalf("Unexpected error: want %q, got %q.", tc.expectedError, err.Error())
				}
			} else {
				if tc.expectedError != "" {
					t.Fatal("Expected an error, but got none.")
				}
			}

			if diff := cmp.Diff(tc.expectedWarnings, warnings, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Unexpected warnings (-wanted +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.expectedConfig, cfg, ignoreType[*regexp.Regexp](), ignoreType[*vm.Program]()); diff != "" {
				t.Fatal("Unexpected config (-wanted +got):\n", diff)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	validConfig := func() Config {
		cfg := defaultConfig()
		cfg.S3.Mode = S3ModePolling
		cfg.S3.PollingPeriod = time.Second
		cfg.Products.ImageGroups = []ImageGroup{
			{
				GroupName: "grp",
				Types: []ImageType{
					{
						Name: "typ",
						DynamicData: DynamicData{
							FileSelectors: requiredObjectFileSelectors(),
						},
					},
				},
			},
		}

		return cfg
	}

	cases := []struct {
		name             string
		mutate           func(*Config)
		expectedWarnings []string
		expectedErrors   []string
	}{
		{
			name: "event mode warns when polling period is set",
			mutate: func(cfg *Config) {
				cfg.S3.Mode = S3ModeEvent
				cfg.S3.PollingPeriod = time.Second
			},
			expectedWarnings: []string{"polling period is ignored when in event mode"},
		},
		{
			name: "polling mode rejects sub-second polling period",
			mutate: func(cfg *Config) {
				cfg.S3.PollingPeriod = time.Second - time.Nanosecond
			},
			expectedErrors: []string{`polling period must be at least one second, not "999.999999ms"`},
		},
		{
			name: "unknown S3 mode",
			mutate: func(cfg *Config) {
				cfg.S3.Mode = "invalid"
			},
			expectedErrors: []string{`unknown S3 mode "invalid" (allowed values are 'polling' / 'event')`},
		},
		{
			name: "invalid dynamic filters",
			mutate: func(cfg *Config) {
				cfg.Products.DynamicFilters = []DynamicFilter{
					{Name: "", Expression: ""},
					{Name: "filter", Expression: "1"},
					{Name: "filter", Expression: "2"},
				}
			},
			expectedErrors: []string{
				"empty name for dynamic filter",
				"empty expression for dynamic filter",
				`duplicate dynamic filter name "filter"`,
			},
		},
		{
			name: "unknown products selector kind",
			mutate: func(cfg *Config) {
				cfg.Products.DynamicData.FileSelectors = map[string]FileSelector{
					"bad": {
						Regex: ".*",
						Kind:  "unknown",
					},
				}
			},
			expectedErrors: []string{`invalid products file selectors: selector "bad": unknown kind "unknown"`},
		},
		{
			name: "no image groups",
			mutate: func(cfg *Config) {
				cfg.Products.ImageGroups = nil
			},
			expectedErrors: []string{"no image groups specified"},
		},
		{
			name: "duplicate image group names",
			mutate: func(cfg *Config) {
				cfg.Products.ImageGroups = append(cfg.Products.ImageGroups, cfg.Products.ImageGroups[0])
			},
			expectedErrors: []string{`image group name "grp" is duplicate`},
		},
		{
			name: "unknown group selector kind",
			mutate: func(cfg *Config) {
				cfg.Products.ImageGroups[0].DynamicData.FileSelectors = map[string]FileSelector{
					"bad": {
						Regex: ".*",
						Kind:  "unknown",
					},
				}
			},
			expectedErrors: []string{`invalid file selectors in group "grp": selector "bad": unknown kind "unknown"`},
		},
		{
			name: "duplicate image type names",
			mutate: func(cfg *Config) {
				cfg.Products.ImageGroups[0].Types = append(cfg.Products.ImageGroups[0].Types, cfg.Products.ImageGroups[0].Types[0])
			},
			expectedErrors: []string{`image type name "typ" of group "grp" is duplicate`},
		},
		{
			name: "unknown type selector kind",
			mutate: func(cfg *Config) {
				cfg.Products.ImageGroups[0].Types[0].DynamicData.FileSelectors["bad"] = FileSelector{
					Regex: ".*",
					Kind:  "unknown",
				}
			},
			expectedErrors: []string{`invalid file selectors in type "typ"/"grp": selector "bad": unknown kind "unknown"`},
		},
		{
			name: "too high UI values",
			mutate: func(cfg *Config) {
				cfg.UI.ScaleInitialPercentage = uint(math.MaxInt) + 1
				cfg.UI.MaxImagesDisplayCount = uint(math.MaxInt) + 1
			},
			expectedErrors: []string{
				"ui.scaleInitialPercentage has a too high value",
				"ui.maxImagesDisplayCount as a too high value",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cfg := validConfig()
			tc.mutate(&cfg)

			warnings, err := cfg.validate()
			if diff := cmp.Diff(tc.expectedWarnings, warnings, cmpopts.EquateEmpty()); diff != "" {
				t.Fatalf("Unexpected warnings (-want +got):\n%s", diff)
			}

			if len(tc.expectedErrors) == 0 {
				if err != nil {
					t.Fatalf("Expected no error, got %q.", err.Error())
				}

				return
			}

			if err == nil {
				t.Fatal("Expected an error, got none.")
			}

			for _, expected := range tc.expectedErrors {
				if !strings.Contains(err.Error(), expected) {
					t.Fatalf("Expected error to contain %q, got %q.", expected, err.Error())
				}
			}
		})
	}
}

func TestProcessInvalidTargetRelativeRegexp(t *testing.T) {
	t.Parallel()

	cfg := defaultConfig()
	cfg.Products.TargetRelativeRegexp = "["

	err := cfg.process()
	if err == nil {
		t.Fatal("Expected an error, got none.")
	}

	if !strings.Contains(err.Error(), "can't parse products.targetRelativeRegexp") {
		t.Fatalf("Unexpected error: %q.", err.Error())
	}
}

func TestMergeDynamicData(t *testing.T) {
	t.Parallel()

	parent := DynamicData{
		FileSelectors: map[string]FileSelector{
			"parentOnly": {
				Regex: "parent",
				Kind:  FileSelectorKindCached,
			},
			"overridden": {
				Regex: "parent-overridden",
				Kind:  FileSelectorKindCached,
			},
		},
		Expressions: map[string]string{
			"parentOnly": "1",
			"overridden": "2",
		},
	}
	child := DynamicData{
		FileSelectors: map[string]FileSelector{
			"childOnly": {
				Regex: "child",
				Kind:  FileSelectorKindSignedURL,
			},
			"overridden": {
				Regex: "child-overridden",
				Kind:  FileSelectorKindSignedURL,
			},
		},
		Expressions: map[string]string{
			"childOnly":  "3",
			"overridden": "4",
		},
	}

	result := mergeDynamicData(child, parent)
	expected := DynamicData{
		FileSelectors: map[string]FileSelector{
			"parentOnly": {
				Regex: "parent",
				Kind:  FileSelectorKindCached,
			},
			"childOnly": {
				Regex: "child",
				Kind:  FileSelectorKindSignedURL,
			},
			"overridden": {
				Regex: "child-overridden",
				Kind:  FileSelectorKindSignedURL,
			},
		},
		Expressions: map[string]string{
			"parentOnly": "1",
			"childOnly":  "3",
			"overridden": "4",
		},
	}

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Fatalf("Unexpected merged dynamic data (-want +got):\n%s", diff)
	}

	result.FileSelectors["parentOnly"] = FileSelector{Regex: "mutated"}
	result.Expressions["parentOnly"] = "mutated"

	if parent.FileSelectors["parentOnly"].Regex != "parent" || parent.Expressions["parentOnly"] != "1" {
		t.Fatal("mergeDynamicData returned maps sharing storage with the parent.")
	}
}

func TestParseFileSelectors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name           string
		selector       FileSelector
		expectedResult FileSelector
		expectedError  string
	}{
		{
			name: "Cached valid",
			selector: FileSelector{
				Regex: "regex",
				Kind:  "cached",
				Link:  false,
			},
			expectedResult: FileSelector{
				Regex: "regex",
				Kind:  FileSelectorKindCached,
				Link:  false,
			},
		},
		{
			name: "FullProductSignedURL valid",
			selector: FileSelector{
				Regex: "regex",
				Kind:  "fullProductSignedURL(exprName)",
				Link:  true,
			},
			expectedResult: FileSelector{
				Regex:      "regex",
				Kind:       FileSelectorKindFullProductSignedURL,
				KindParams: []string{"exprName"},
				Link:       true,
			},
		},
		{
			name: "FullProductSignedURL invalid",
			selector: FileSelector{
				Regex: "regex",
				Kind:  "fullProductSignedURL()",
				Link:  true,
			},
			expectedError: `"FullProductSignedURL invalid": invalid fullProductSignedURL expression "fullProductSignedURL()"`,
		},
		{
			name: "Invalid regex",
			selector: FileSelector{
				Regex: "[",
				Kind:  "cached",
			},
			expectedError: "\"Invalid regex\": error parsing regexp: missing closing ]: `[`",
		},
		{
			name: "ExternalViewerURL valid",
			selector: FileSelector{
				Regex: "regex",
				Kind:  "externalViewerURL(viewerName, exprName)",
				Link:  true,
			},
			expectedResult: FileSelector{
				Regex:      "regex",
				Kind:       FileSelectorKindExternalViewerURL,
				KindParams: []string{"viewerName", "exprName"},
				Link:       true,
			},
		},
		{
			name: "ExternalViewerURL invalid missing viewer",
			selector: FileSelector{
				Regex: "regex",
				Kind:  "externalViewerURL(, exprName)",
				Link:  true,
			},
			expectedError: `"ExternalViewerURL invalid missing viewer": invalid externalViewerURL expression "externalViewerURL(, exprName)"`,
		},
		{
			name: "ExternalViewerURL invalid missing expression",
			selector: FileSelector{
				Regex: "regex",
				Kind:  "externalViewerURL(viewerName, )",
				Link:  true,
			},
			expectedError: `"ExternalViewerURL invalid missing expression": invalid externalViewerURL expression "externalViewerURL(viewerName, )"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m := map[string]FileSelector{tc.name: tc.selector}

			err := parseFileSelectors(m)
			if err != nil {
				if tc.expectedError == "" {
					t.Fatalf("Expected no error, but got %q.", err.Error())
				} else if err.Error() != tc.expectedError {
					t.Fatalf("Unexpected error message: want %q, got %q.", tc.expectedError, err.Error())
				}

				return
			} else if tc.expectedError != "" {
				t.Fatal("Expected an error, but got none.")
			}

			if diff := cmp.Diff(tc.expectedResult, m[tc.name], cmpopts.IgnoreTypes(&regexp.Regexp{})); diff != "" {
				t.Fatalf("Unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseDynamicDataExternalViewerURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		externalViewers  map[string]string
		expressions      map[string]string
		expectedSelector FileSelector
		expectedError    string
	}{
		{
			name: "valid external viewer selector",
			externalViewers: map[string]string{
				"viewerName": "https://viewer.example.test/?url=",
			},
			expressions: map[string]string{
				"viewerURL": `'s3://bucket/path/file.tif'`,
			},
			expectedSelector: FileSelector{
				Regex:      `product\.tif$`,
				Kind:       FileSelectorKindExternalViewerURL,
				KindParams: []string{"viewerName", "viewerURL"},
				Link:       true,
			},
		},
		{
			name:            "missing external viewer",
			externalViewers: map[string]string{},
			expressions: map[string]string{
				"viewerURL": `'s3://bucket/path/file.tif'`,
			},
			expectedError: `invalid products.imageGroups["grp"].types["typ"].dynamicData.fileSelectors["product"]: externalViewerURL references the external viewer "viewerName" which is not defined`,
		},
		{
			name: "missing expression",
			externalViewers: map[string]string{
				"viewerName": "https://viewer.example.test/?url=",
			},
			expressions:   map[string]string{},
			expectedError: `invalid products.imageGroups["grp"].types["typ"].dynamicData.fileSelectors["product"]: externalViewerURL references the expression "viewerURL" which is not defined`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dynData := DynamicData{
				FileSelectors: map[string]FileSelector{
					"product": {
						Regex: `product\.tif$`,
						Kind:  "externalViewerURL(viewerName, viewerURL)",
					},
				},
				Expressions: tc.expressions,
			}

			err := ParseDynamicData("grp", "typ", &dynData, tc.externalViewers)
			if err != nil {
				if tc.expectedError == "" {
					t.Fatalf("Expected no error, but got %q.", err.Error())
				} else if err.Error() != tc.expectedError {
					t.Fatalf("Unexpected error message: want %q, got %q.", tc.expectedError, err.Error())
				}

				return
			} else if tc.expectedError != "" {
				t.Fatal("Expected an error, but got none.")
			}

			if diff := cmp.Diff(tc.expectedSelector, dynData.FileSelectors["product"], cmpopts.IgnoreTypes(&regexp.Regexp{})); diff != "" {
				t.Fatalf("Unexpected selector (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseDynamicDataFullProductSignedURL(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		expressions      map[string]string
		expectedSelector FileSelector
		expectedError    string
	}{
		{
			name: "valid full product selector",
			expressions: map[string]string{
				"urlParams": `{"foo": "bar"}`,
			},
			expectedSelector: FileSelector{
				Regex:      `product\.tif$`,
				Kind:       FileSelectorKindFullProductSignedURL,
				KindParams: []string{"urlParams"},
				Link:       true,
			},
		},
		{
			name:          "missing expression",
			expressions:   map[string]string{},
			expectedError: `invalid products.imageGroups["grp"].types["typ"].dynamicData.fileSelectors["product"]: fullProductSignedURL references the expression "urlParams" which is not defined`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dynData := DynamicData{
				FileSelectors: map[string]FileSelector{
					"product": {
						Regex: `product\.tif$`,
						Kind:  "fullProductSignedURL(urlParams)",
					},
				},
				Expressions: tc.expressions,
			}

			err := ParseDynamicData("grp", "typ", &dynData, nil)
			if err != nil {
				if tc.expectedError == "" {
					t.Fatalf("Expected no error, but got %q.", err.Error())
				} else if err.Error() != tc.expectedError {
					t.Fatalf("Unexpected error message: want %q, got %q.", tc.expectedError, err.Error())
				}

				return
			} else if tc.expectedError != "" {
				t.Fatal("Expected an error, but got none.")
			}

			if diff := cmp.Diff(tc.expectedSelector, dynData.FileSelectors["product"], cmpopts.IgnoreTypes(&regexp.Regexp{})); diff != "" {
				t.Fatalf("Unexpected selector (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseExpressionsErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		expression    string
		expectedError string
	}{
		{
			name:          "compile error",
			expression:    "1 +",
			expectedError: `expression "expr": unexpected token`,
		},
		{
			name:          "replace regex second argument must be string",
			expression:    `_replaceRegex("value", Files.preview.S3Path, "replacement")`,
			expectedError: `expression "expr": _replaceRegex: second argument must be a string`,
		},
		{
			name:          "replace regex invalid regex",
			expression:    `_replaceRegex("value", "[", "replacement")`,
			expectedError: `expression "expr": _replaceRegex: error parsing regexp: missing closing ]`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := parseExpressions(map[string]string{"expr": tc.expression})
			if err == nil {
				t.Fatal("Expected an error, got none.")
			}

			if !strings.Contains(err.Error(), tc.expectedError) {
				t.Fatalf("Expected error to contain %q, got %q.", tc.expectedError, err.Error())
			}
		})
	}
}

func requiredObjectFileSelectors() map[string]FileSelector {
	return map[string]FileSelector{
		"preview": {
			Regex: "preview",
			Kind:  FileSelectorKindCached,
		},
		"geonames": {
			Regex: "geonames",
			Kind:  FileSelectorKindCached,
		},
		"localization": {
			Regex: "localization",
			Kind:  FileSelectorKindCached,
		},
	}
}
