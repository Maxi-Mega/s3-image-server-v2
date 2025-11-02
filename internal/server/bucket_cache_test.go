package server

import (
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
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

func TestInjectFullProductURLParamsFromExpr(t *testing.T) {
	t.Parallel()

	cases := []struct {
		signedURL           string
		expression          string
		fullProductProtocol string
		fullProductRootURL  string
		expectedURL         string
	}{
		{
			signedURL: "https://example.com/file.tif",
			expression: `{
				"str": "value",
				"length": len(Files.preview.S3Path),
			}`,
			fullProductProtocol: "http://custom?url=",
			fullProductRootURL:  "https://some.other.host",
			expectedURL:         "http://custom?length=3&str=value&url=https%3A%2F%2Fsome.other.host%2Ffile.tif",
		},
	}

	const (
		exprName  = "urlParams"
		imgGroup  = "grp"
		imgType   = "typ"
		objectKey = "key"
	)

	for i, tc := range cases {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			t.Parallel()

			options := append(
				[]expr.Option{expr.Env(types.ExprEnv{})},
				types.ExprFunctions...,
			)

			exprPrgm, err := expr.Compile(tc.expression, options...)
			if err != nil {
				t.Fatalf("Failed to compile expression: %v", err)
			}

			bc := bucketCache{
				exprManager: &expressionManager{
					exprs: map[string]map[string]map[string]*vm.Program{
						imgGroup: {
							imgType: {
								exprName: exprPrgm,
							},
						},
					},
					cacheSums: map[exprCacheKey]exprCacheEntry{},
				},
				cfg: config.Config{
					Products: config.Products{
						FullProductProtocol: tc.fullProductProtocol,
						FullProductRootURL:  tc.fullProductRootURL,
					},
				},
			}

			img := image{
				lastModified:      time.Now(),
				s3Key:             objectKey,
				imgGroup:          imgGroup,
				imgType:           imgType,
				dynamicInputFiles: map[string]valueWithLastUpdate[types.DynamicInputFile]{},
				previewCacheKey:   objectKey,
			}

			tcSU, err := url.Parse(tc.signedURL)
			if err != nil {
				t.Fatalf("Failed to parse signed URL: %v", err)
			}

			result := bc.injectFullProductURLParamsFromExpr(t.Context(), tcSU, img, exprName)
			if result != tc.expectedURL {
				t.Fatalf("Expected %q, got %q", tc.expectedURL, result)
			}
		})
	}
}
