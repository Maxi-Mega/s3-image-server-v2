package server

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
)

func TestGetCacheKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		bucket   string
		imgName  string
		subDir   string
		filename string
		expected string
	}{
		{
			bucket:   "prod",
			imgName:  "1/2/3",
			filename: "preview.jpg",
			expected: "prod/1/2/3/preview.jpg",
		},
		{
			bucket:   "b-313",
			imgName:  "A/B_C/",
			subDir:   targetsDirName,
			filename: "preview.jpg",
			expected: "b-313/A/B_C/__targets__/preview.jpg",
		},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			t.Parallel()

			bc := bucketCache{bucket: tc.bucket}

			result := bc.getCacheKey(tc.imgName, tc.subDir, tc.filename)
			if result != tc.expected {
				t.Fatalf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestInjectFullProductURLParams(t *testing.T) {
	t.Parallel()

	cases := []struct {
		objectKey   string
		signedURL   string
		imgGroup    config.ImageGroup
		expectedURL string
	}{
		{
			objectKey:   "any",
			signedURL:   "https://example.com/file.tif",
			imgGroup:    config.ImageGroup{GroupName: "No regex"},
			expectedURL: "https://example.com/file.tif",
		},
		{
			objectKey: "some/path/to/file.tif",
			signedURL: "https://example.com/file.tif",
			imgGroup: config.ImageGroup{
				FullPoductURLParams: []config.FullProductURLParam{
					{
						Name:  "const",
						Type:  config.FullProductURLParamConstant,
						Value: "val",
					},
					{
						Name: "dyn",
						Type: config.FullProductURLParamRegexp,
					},
					{
						Name:         "mapped",
						Type:         config.FullProductURLParamRegexp,
						ValueMapping: map[string]string{"from": "1", "to": "2"},
					},
				},
				FullProductURLParamsRgx: regexp.MustCompile(`^\w+/(?P<dyn>\w+)/(?P<mapped>[[:alpha:]]+)/file\.tif$`),
			},
			expectedURL: "https://example.com/file.tif?const=val&dyn=path&mapped=2",
		},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			t.Parallel()

			result := injectFullProductURLParams(tc.imgGroup, tc.objectKey, tc.signedURL)
			if result != tc.expectedURL {
				t.Fatalf("Expected %q, got %q", tc.expectedURL, result)
			}
		})
	}
}
