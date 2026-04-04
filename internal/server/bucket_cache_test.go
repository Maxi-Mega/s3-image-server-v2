package server

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strconv"
	"testing"
	"time"

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

			exprManager := new(expressionManager{
				exprs: map[string]map[string]map[string]*vm.Program{
					imgGroup: {
						imgType: {
							exprName: exprPrgm,
						},
					},
				},
				cacheSums: map[exprCacheKey]exprCacheEntry{},
			})

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

			result := injectFullProductURLParamsFromExpr(t.Context(), exprManager, tcSU, img, exprName, tc.fullProductProtocol, tc.fullProductRootURL)
			if result != tc.expectedURL {
				t.Fatalf("Expected %q, got %q", tc.expectedURL, result)
			}
		})
	}
}

func TestMakeSignedURL(t *testing.T) {
	t.Parallel()

	const (
		bucket    = "bucket-a"
		objectKey = "products/path/file.tif"
		imgGroup  = "grp"
		imgType   = "typ"
		exprName  = "urlParams"
	)

	baseURL := "https://signed.example.com/products/path/file.tif?X-Amz-Signature=abc"
	objLastModified := time.Date(2026, 3, 10, 8, 30, 0, 0, time.UTC)

	options := append(
		[]expr.Option{expr.Env(types.ExprEnv{})},
		types.ExprFunctions...,
	)

	exprPrgm, err := expr.Compile(`{
		"source": Files.preview.S3Path,
		"length": len(Files.preview.S3Path),
	}`, options...)
	if err != nil {
		t.Fatalf("Failed to compile expression: %v", err)
	}

	exprManager := new(expressionManager{
		exprs: map[string]map[string]map[string]*vm.Program{
			imgGroup: {
				imgType: {
					exprName: exprPrgm,
				},
			},
		},
		cacheSums: map[exprCacheKey]exprCacheEntry{},
	})

	img := image{
		name:         "img-a",
		baseDir:      "img-a",
		bucket:       bucket,
		s3Key:        "preview.jpg",
		imgGroup:     imgGroup,
		imgType:      imgType,
		lastModified: objLastModified,
		dynamicInputFiles: map[string]valueWithLastUpdate[types.DynamicInputFile]{
			"preview": {
				value: types.DynamicInputFile{
					S3Path: "preview",
				},
			},
		},
	}

	cases := []struct {
		name            string
		genReq          signedURLGenerationRequest
		mockErr         error
		wantURL         string
		wantParamsExpr  string
		expectErr       error
		assertGenerated func(t *testing.T, got signedURL)
	}{
		{
			name: "returns plain signed URL without param injection",
			genReq: signedURLGenerationRequest{
				bucket:          bucket,
				s3Key:           objectKey,
				objLastModified: objLastModified,
				img:             img,
			},
			wantURL: baseURL,
			assertGenerated: func(t *testing.T, got signedURL) {
				t.Helper()

				now := time.Now().Truncate(time.Second)
				if got.generationDate.Before(now.Add(-2*time.Second)) || got.generationDate.After(now.Add(2*time.Second)) {
					t.Fatalf("unexpected generation date: %s", got.generationDate)
				}
			},
		},
		{
			name: "injects full product params when requested",
			genReq: signedURLGenerationRequest{
				bucket:              bucket,
				s3Key:               objectKey,
				objLastModified:     objLastModified,
				img:                 img,
				injectParams:        true,
				paramsExpr:          exprName,
				exprManager:         exprManager,
				fullProductProtocol: "http://custom?url=",
				fullProductRootURL:  "https://download.example.test",
			},
			wantParamsExpr: exprName,
			wantURL:        "http://custom?length=7&source=preview&url=https%3A%2F%2Fdownload.example.test%2Fproducts%2Fpath%2Ffile.tif%3FX-Amz-Signature%3Dabc",
		},
		{
			name: "propagates S3 signed URL generation errors",
			genReq: signedURLGenerationRequest{
				bucket:          bucket,
				s3Key:           objectKey,
				objLastModified: objLastModified,
				img:             img,
			},
			mockErr:   errors.New("presign failed"),
			expectErr: errors.New("presign failed"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotBucket, gotObjectKey string

			s3Client := S3ClientMock{
				GenerateSignedURLFn: func(_ context.Context, bucketArg, objectKeyArg string) (*url.URL, error) {
					gotBucket = bucketArg
					gotObjectKey = objectKeyArg

					if tc.mockErr != nil {
						return nil, tc.mockErr
					}

					return url.Parse(baseURL)
				},
			}

			got, err := makeSignedURL(t.Context(), s3Client, tc.genReq)
			if tc.expectErr != nil {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}

				if err.Error() != tc.expectErr.Error() {
					t.Fatalf("unexpected error: got %q, want %q", err, tc.expectErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotBucket != bucket || gotObjectKey != objectKey {
				t.Fatalf("GenerateSignedURL called with (%q, %q), want (%q, %q)", gotBucket, gotObjectKey, bucket, objectKey)
			}

			if got.value.value != tc.wantURL {
				t.Fatalf("unexpected signed URL: got %q, want %q", got.value.value, tc.wantURL)
			}

			if got.value.paramsExpr != tc.wantParamsExpr {
				t.Fatalf("unexpected params expr: got %q, want %q", got.value.paramsExpr, tc.wantParamsExpr)
			}

			if !got.lastUpdate.Equal(objLastModified) {
				t.Fatalf("unexpected last update: got %s, want %s", got.lastUpdate, objLastModified)
			}

			if tc.assertGenerated != nil {
				tc.assertGenerated(t, got.value)
			}
		})
	}
}

func TestFindSignedURLsToRenew(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

	makeImage := func(name string, urls map[string]valueWithLastUpdate[signedURL]) image {
		return image{
			name:              name,
			baseDir:           name,
			bucket:            "bucket-a",
			imgGroup:          "grp",
			imgType:           "typ",
			lastModified:      baseTime,
			dynamicInputFiles: map[string]valueWithLastUpdate[types.DynamicInputFile]{},
			targets:           map[string]valueWithLastUpdate[string]{},
			linksFromCache:    map[string]valueWithLastUpdate[string]{},
			signedURLs:        urls,
		}
	}

	cases := []struct {
		name       string
		deadline   time.Time
		lifetime   time.Duration
		images     map[string]image
		wantByImg  map[string][]string
		wantParams map[string]string
		wantUpdate map[string]time.Time
	}{
		{
			name:     "returns only URLs expiring strictly before deadline",
			deadline: baseTime.Add(24 * time.Hour),
			lifetime: 6 * time.Hour,
			images: map[string]image{
				"img-a": makeImage("img-a", map[string]valueWithLastUpdate[signedURL]{
					"renew-me": {
						lastUpdate: baseTime.Add(-10 * time.Hour),
						value: signedURL{
							value:          "https://example.test/renew-me",
							paramsExpr:     "expr-renew",
							generationDate: baseTime.Add(-20 * time.Hour),
						},
					},
					"on-boundary": {
						lastUpdate: baseTime.Add(-9 * time.Hour),
						value: signedURL{
							value:          "https://example.test/on-boundary",
							paramsExpr:     "expr-boundary",
							generationDate: baseTime.Add(18 * time.Hour),
						},
					},
					"keep": {
						lastUpdate: baseTime.Add(-8 * time.Hour),
						value: signedURL{
							value:          "https://example.test/keep",
							paramsExpr:     "expr-keep",
							generationDate: baseTime.Add(19 * time.Hour),
						},
					},
				}),
			},
			wantByImg: map[string][]string{
				"img-a": {"renew-me"},
			},
			wantParams: map[string]string{
				"img-a/renew-me": "expr-renew",
			},
			wantUpdate: map[string]time.Time{
				"img-a/renew-me": baseTime.Add(-10 * time.Hour),
			},
		},
		{
			name:     "supports multiple images and custom lifetime",
			deadline: baseTime.Add(36 * time.Hour),
			lifetime: 12 * time.Hour,
			images: map[string]image{
				"img-a": makeImage("img-a", map[string]valueWithLastUpdate[signedURL]{
					"a-renew": {
						lastUpdate: baseTime.Add(-6 * time.Hour),
						value: signedURL{
							value:          "https://example.test/a-renew",
							paramsExpr:     "expr-a",
							generationDate: baseTime.Add(10 * time.Hour),
						},
					},
					"a-keep": {
						lastUpdate: baseTime.Add(-5 * time.Hour),
						value: signedURL{
							value:          "https://example.test/a-keep",
							paramsExpr:     "expr-a-keep",
							generationDate: baseTime.Add(30 * time.Hour),
						},
					},
				}),
				"img-b": makeImage("img-b", map[string]valueWithLastUpdate[signedURL]{
					"b-renew": {
						lastUpdate: baseTime.Add(-4 * time.Hour),
						value: signedURL{
							value:          "https://example.test/b-renew",
							paramsExpr:     "expr-b",
							generationDate: baseTime.Add(-1 * time.Hour),
						},
					},
				}),
				"img-c": makeImage("img-c", nil),
			},
			wantByImg: map[string][]string{
				"img-a": {"a-renew"},
				"img-b": {"b-renew"},
			},
			wantParams: map[string]string{
				"img-a/a-renew": "expr-a",
				"img-b/b-renew": "expr-b",
			},
			wantUpdate: map[string]time.Time{
				"img-a/a-renew": baseTime.Add(-6 * time.Hour),
				"img-b/b-renew": baseTime.Add(-4 * time.Hour),
			},
		},
		{
			name:     "returns empty when all URLs expire at or after deadline",
			deadline: baseTime.Add(8 * time.Hour),
			lifetime: 4 * time.Hour,
			images: map[string]image{
				"img-a": makeImage("img-a", map[string]valueWithLastUpdate[signedURL]{
					"at-deadline": {
						lastUpdate: baseTime.Add(-2 * time.Hour),
						value: signedURL{
							value:          "https://example.test/at-deadline",
							paramsExpr:     "expr-deadline",
							generationDate: baseTime.Add(4 * time.Hour),
						},
					},
					"after-deadline": {
						lastUpdate: baseTime.Add(-1 * time.Hour),
						value: signedURL{
							value:          "https://example.test/after-deadline",
							paramsExpr:     "expr-after",
							generationDate: baseTime.Add(5 * time.Hour),
						},
					},
				}),
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			bc := &bucketCache{
				bucket: "bucket-a",
				images: tc.images,
			}

			got := bc.findSignedURLsToRenew(tc.deadline, tc.lifetime)
			if len(got) != len(tc.wantByImg) {
				t.Fatalf("unexpected image count: got %d, want %d", len(got), len(tc.wantByImg))
			}

			for imgName, wantKeys := range tc.wantByImg {
				entries, ok := got[imgName]
				if !ok {
					t.Fatalf("missing image %q in renewal set", imgName)
				}

				gotKeys := make([]string, 0, len(entries))
				for s3Key, req := range entries {
					gotKeys = append(gotKeys, s3Key)

					lookupKey := imgName + "/" + s3Key
					if req.paramsExpr != tc.wantParams[lookupKey] {
						t.Fatalf("unexpected paramsExpr for %q: got %q, want %q", lookupKey, req.paramsExpr, tc.wantParams[lookupKey])
					}

					if !req.objectLastModified.Equal(tc.wantUpdate[lookupKey]) {
						t.Fatalf("unexpected last update for %q: got %s, want %s", lookupKey, req.objectLastModified, tc.wantUpdate[lookupKey])
					}

					if req.img.name != tc.images[imgName].name {
						t.Fatalf("unexpected image embedded in request for %q: got %q, want %q", lookupKey, req.img.name, tc.images[imgName].name)
					}
				}

				slices.Sort(gotKeys)
				slices.Sort(wantKeys)

				if !slices.Equal(gotKeys, wantKeys) {
					t.Fatalf("unexpected renewed keys for %q: got %v, want %v", imgName, gotKeys, wantKeys)
				}
			}
		})
	}
}
