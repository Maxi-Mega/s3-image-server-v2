package config

import (
	"regexp"
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

func TestLoad(t *testing.T) {
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
					PollingMode:   true,
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
										},
										Expressions: map[string]string{
											"key":       "'value'",
											"image":     "'DIR/DIR2/IMAGE.tif'",
											"urlParams": `{"p": "val", "n": 5}`,
										},
										ExpressionsPrograms: map[string]*vm.Program{
											"key":       nil,
											"image":     nil,
											"urlParams": nil,
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
					PrometheusInstanceLabel: "s3_image_server",
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
					PollingMode:   true,
					PollingPeriod: 30 * time.Second,
					Endpoint:      "localhost:9000", // without "https://"
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
				},
			},
			expectedError: "",
		},
		{
			filePath:      "invalid_yaml.yml",
			expectedError: "failed to parse config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid...` into config.Config",
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
					t.Fatalf("Unexpected error: want %q, got %q.", tc.expectedError, err.Error())
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
