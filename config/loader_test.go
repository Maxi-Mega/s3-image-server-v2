package config

import (
	"regexp"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestLoad(t *testing.T) {
	t.Parallel()

	cases := []struct {
		filePath       string
		expectedConfig Config
		expectedError  string
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
					MaxImagesDisplayCount:  10,
					Map: UIMap{
						TileServerURL: "localhost:3000/{z}/{x}/{y}.png",
					},
				},
				Products: Products{
					DefaultPreviewSuffix: "preview.jpg",
					GeonamesFilename:     "geonames.json",
					LocalizationFilename: "localization.json",
					ImageGroups: []ImageGroup{
						{
							GroupName: "Group 1",
							Types: []ImageType{
								{
									Name:        "1",
									DisplayName: "One",
								},
								{
									Name:          "2",
									DisplayName:   "Two",
									PreviewSuffix: "other",
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
			expectedError: "",
		},
		{
			filePath:       "invalid_yaml.yml",
			expectedConfig: Config{},
			expectedError:  "failed to parse config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid...` into config.Config",
		},
		{
			filePath:       "invalid_cfg.yml",
			expectedConfig: Config{},
			expectedError:  `the config is invalid: image type name "typ" of group "grp" is duplicate`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.filePath, func(t *testing.T) {
			t.Parallel()

			cfg, err := Load("./testdata/" + tc.filePath)
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

			if diff := cmp.Diff(tc.expectedConfig, cfg, cmpopts.IgnoreTypes(&regexp.Regexp{})); diff != "" {
				t.Fatal("Unexpected config (-wanted +got):\n", diff)
			}
		})
	}
}
