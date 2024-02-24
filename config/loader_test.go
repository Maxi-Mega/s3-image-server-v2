package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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
					EndPoint: "localhost:9000",
				},
				UI: UI{
					Map: UIMap{
						TileServerURL: "localhost:3000/{z}/{x}/{y}.png",
					},
				},
				Products: Products{
					ImageGroups: []ImageGroup{
						{
							GroupName: "Group 1",
							Types: []ImageType{
								{
									Name:        "1",
									DisplayName: "One",
								},
								{
									Name:        "2",
									DisplayName: "Two",
								},
							},
						},
					},
				},
				Log: Log{
					JSONLogFormat: true,
					JSONLogFields: map[string]any{
						"f1": "field",
						"f2": 7,
					},
				},
			},
			expectedError: "",
		},
		{
			filePath:       "invalid_cfg.yml",
			expectedConfig: Config{},
			expectedError:  "failed to parse config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid...` into config.Config",
		},
	}

	for _, tc := range cases { //nolint:paralleltest // Not a problem since Go 1.22
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

			if diff := cmp.Diff(tc.expectedConfig, cfg); diff != "" {
				t.Fatal("Unexpected config (-wanted +got):\n", diff)
			}
		})
	}
}
