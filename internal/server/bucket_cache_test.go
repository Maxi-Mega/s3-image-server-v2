package server

import (
	"strconv"
	"testing"
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
