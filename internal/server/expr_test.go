package server

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/s3"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
)

const (
	imgGroup = "grp"
	imgType  = "typ"
)

func setupExprManTest(t *testing.T, dynData *config.DynamicData, externalViewers map[string]string, files map[string]string) *expressionManager {
	t.Helper()

	err := config.ParseDynamicData(imgGroup, imgType, dynData, externalViewers)
	if err != nil {
		t.Fatal("Invalid dynamic data:", err)
	}

	cfg := config.Config{
		Products: config.Products{
			ImageGroups: []config.ImageGroup{
				{
					GroupName: imgGroup,
					Types: []config.ImageType{
						{
							Name:        imgType,
							DynamicData: *dynData,
						},
					},
				},
			},
		},
		Cache: config.Cache{
			CacheDir: t.TempDir(),
		},
	}
	exprMan := newExpressionManager(cfg)

	for filePath, content := range files {
		fullPath := filepath.Join(cfg.Cache.CacheDir, filePath)

		err = os.MkdirAll(filepath.Dir(fullPath), 0755)
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.Create(fullPath)
		if err != nil {
			t.Fatal(err)
		}

		defer f.Close()

		_, err = f.WriteString(content)
		if err != nil {
			t.Fatal(err)
		}
	}

	return exprMan
}

func TestExprProductBasePath(t *testing.T) {
	t.Parallel()

	dynamicData := config.DynamicData{
		FileSelectors: map[string]config.FileSelector{
			types.ObjectPreview: {
				Regex: `preview\.jpg$`,
				Kind:  config.FileSelectorKindCached,
			},
		},
		Expressions: map[string]string{
			types.ExprProductBasePath: "__testCounter__(); Files.preview.S3Path[:lastIndexOf(Files.preview.S3Path, '/')]",
		},
	}
	externalViewers := map[string]string{}
	files := map[string]string{
		"prod/1/2/3/preview.jpg": "some jpg data",
	}
	exprMan := setupExprManTest(t, &dynamicData, externalViewers, files)

	for i := range 2 {
		exprCallCounter := new(atomic.Int64)
		ctx := context.WithValue(t.Context(), types.ExprTestCounterKey{}, exprCallCounter)

		s3Evt := s3.Event{
			Bucket:             "prod",
			ObjectKey:          "1/2/3/preview.jpg",
			ObjectLastModified: time.Now(),
		}

		for j := range 2 { // 1st call should run the expression, 2nd should use the cache
			basepath, err := exprMan.productBasePath(ctx, imgGroup, imgType, s3Evt)
			if err != nil {
				t.Fatalf("[call %d-%d] Failed to get base path: %v", j+1, i+1, err)
			}

			if basepath != "1/2/3" {
				t.Fatalf("[call %d-%d] Expected base path to be %q, got %q", j+1, i+1, "1/2/3", basepath)
			}

			if exprCallCounter.Load() != 1 { // should not increase, due to cache
				t.Fatalf("[call %d-%d] Expected expression call counter to be %d, got %d", j+1, i+1, 1, exprCallCounter.Load())
			}
		}
	}
}

func TestExprExternalViewerURL(t *testing.T) {
	t.Parallel()

	const (
		inputFile = "product"
		exprName  = "viewerURL"
	)

	objectLastModified := time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC)
	dynamicData := config.DynamicData{
		FileSelectors: map[string]config.FileSelector{
			inputFile: {
				Regex: `product\.tif$`,
				Kind:  "externalViewerURL(viewer, viewerURL)",
			},
		},
		Expressions: map[string]string{
			exprName: "__testCounter__(); _s3Uri('product')",
		},
	}
	externalViewers := map[string]string{
		"viewer": "https://viewer.example.test/?url=",
	}
	exprMan := setupExprManTest(t, &dynamicData, externalViewers, nil)
	img := image{
		lastModified: objectLastModified,
		bucket:       "preview-bucket",
		s3Key:        "products/1/preview.jpg",
		imgGroup:     imgGroup,
		imgType:      imgType,
		dynamicInputFiles: map[string]valueWithLastUpdate[types.DynamicInputFile]{
			inputFile: {
				value: types.DynamicInputFile{
					S3Bucket: "product-bucket",
					S3Path:   "products/1/product.tif",
					Date:     objectLastModified,
				},
				lastUpdate: objectLastModified,
			},
		},
		previewCacheKey: "products/1/preview.jpg",
	}
	exprCallCounter := new(atomic.Int64)
	ctx := context.WithValue(t.Context(), types.ExprTestCounterKey{}, exprCallCounter)

	for i := range 2 {
		result, err := exprMan.externalViewerURL(ctx, img, exprName)
		if err != nil {
			t.Fatalf("[call %d] Failed to get external viewer URL input: %v", i+1, err)
		}

		if result != "s3://product-bucket/products/1/product.tif" {
			t.Fatalf("[call %d] Expected %q, got %q", i+1, "s3://product-bucket/products/1/product.tif", result)
		}

		if exprCallCounter.Load() != 1 {
			t.Fatalf("[call %d] Expected expression call counter to be %d, got %d", i+1, 1, exprCallCounter.Load())
		}
	}
}

func TestExprExternalViewerURLUnexpectedOutputType(t *testing.T) {
	t.Parallel()

	dynamicData := config.DynamicData{
		FileSelectors: map[string]config.FileSelector{
			"product": {
				Regex: `product\.tif$`,
				Kind:  "externalViewerURL(viewer, viewerURL)",
			},
		},
		Expressions: map[string]string{
			"viewerURL": "42",
		},
	}
	externalViewers := map[string]string{
		"viewer": "https://viewer.example.test/?url=",
	}
	exprMan := setupExprManTest(t, &dynamicData, externalViewers, nil)
	img := image{
		lastModified:      time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC),
		bucket:            "preview-bucket",
		s3Key:             "products/1/preview.jpg",
		imgGroup:          imgGroup,
		imgType:           imgType,
		dynamicInputFiles: map[string]valueWithLastUpdate[types.DynamicInputFile]{},
		previewCacheKey:   "products/1/preview.jpg",
	}

	_, err := exprMan.externalViewerURL(t.Context(), img, "viewerURL")
	if err == nil {
		t.Fatal("Expected an error, but got none.")
	}

	if err.Error() != "unexpected output type: want string, got int" {
		t.Fatalf("Unexpected error: want %q, got %q", "unexpected output type: want string, got int", err.Error())
	}
}
