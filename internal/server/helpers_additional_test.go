package server

import (
	"regexp"
	"testing"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/s3"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
)

func TestBaseDirRelativePath(t *testing.T) {
	t.Parallel()

	evt := s3Event{
		Event: s3.Event{
			ObjectKey: "a/b/c/file.json",
		},
		baseDir: "a/b",
	}

	got := evt.baseDirRelativePath()
	if got != "c@file.json" {
		t.Fatalf("unexpected path: got %q", got)
	}
}

func TestMatchesFileSelectorRegex(t *testing.T) {
	t.Parallel()

	imgType := config.ImageType{
		DynamicData: config.DynamicData{
			FileSelectors: map[string]config.FileSelector{
				types.ObjectPreview: {
					Rgx: regexp.MustCompile(`preview\.jpg$`),
				},
			},
		},
	}

	if !matchesFileSelectorRegex("root/preview.jpg", types.ObjectPreview, imgType) {
		t.Fatal("expected selector to match")
	}

	if matchesFileSelectorRegex("root/preview.png", types.ObjectPreview, imgType) {
		t.Fatal("did not expect selector to match")
	}

	if matchesFileSelectorRegex("root/preview.jpg", types.ObjectTarget, imgType) {
		t.Fatal("did not expect missing selector to match")
	}
}

func TestGetObjectType(t *testing.T) {
	t.Parallel()

	consumer := eventConsumer{}
	imgType := config.ImageType{
		DynamicData: config.DynamicData{
			FileSelectors: map[string]config.FileSelector{
				types.ObjectPreview: {
					Rgx: regexp.MustCompile(`preview\.jpg$`),
				},
				"metadata": {
					Rgx: regexp.MustCompile(`meta\.json$`),
				},
			},
		},
	}

	objType, inputFile := consumer.getObjectType("x/preview.jpg", imgType)
	if objType != types.ObjectPreview || inputFile != "" {
		t.Fatalf("unexpected preview result: %q %q", objType, inputFile)
	}

	objType, inputFile = consumer.getObjectType("x/meta.json", imgType)
	if objType != types.ObjectDynamicInput || inputFile != "metadata" {
		t.Fatalf("unexpected dynamic input result: %q %q", objType, inputFile)
	}

	objType, inputFile = consumer.getObjectType("x/other.bin", imgType)
	if objType != types.ObjectNotYetAssigned || inputFile != "" {
		t.Fatalf("unexpected fallback result: %q %q", objType, inputFile)
	}
}

func TestSignedURLIsValid(t *testing.T) {
	t.Parallel()

	valid := signedURL{
		generationDate: time.Now().Add(-1 * time.Hour),
	}
	if !valid.isValid() {
		t.Fatal("expected signed URL to be valid")
	}

	expired := signedURL{
		generationDate: time.Now().Add(-(s3.SignedURLLifetime + time.Second)),
	}
	if expired.isValid() {
		t.Fatal("expected signed URL to be expired")
	}
}

func TestToFilenameValueMap(t *testing.T) {
	t.Parallel()

	m := map[string]valueWithLastUpdate[int]{
		"a": {value: 7},
		"b": {value: 42},
	}

	got := toFilenameValueMap(m)
	if got["a"] != "7" || got["b"] != "42" {
		t.Fatalf("unexpected map values: %#v", got)
	}
}

func TestObjectTemporizerComputeEvent(t *testing.T) {
	t.Parallel()

	op := objectTemporizer{
		productsCfg: config.Products{
			TargetRelativeRgx: regexp.MustCompile(`^/targets/.*`),
		},
	}

	evt := s3Event{
		Event: s3.Event{
			ObjectType: types.ObjectNotYetAssigned,
			ObjectKey:  "prod/base/targets/a.bin",
		},
	}

	got, ok := op.computeEvent(evt, "prod/base")
	if !ok {
		t.Fatal("expected event to be assignable")
	}

	if got.ObjectType != types.ObjectTarget {
		t.Fatalf("unexpected object type: %q", got.ObjectType)
	}

	if got.baseDir != "prod/base" {
		t.Fatalf("unexpected baseDir: %q", got.baseDir)
	}
}

func TestObjectTemporizerPurge(t *testing.T) {
	t.Parallel()

	now := time.Now()
	op := objectTemporizer{
		unassignedObjects: map[string][]oot{
			"old": {
				{appendTime: now.Add(-(ootMaxLifetime + time.Second))},
			},
			"mixed": {
				{appendTime: now.Add(-(ootMaxLifetime + time.Second))},
				{appendTime: now.Add(-1 * time.Minute)},
			},
		},
	}

	op.purge(now)

	if _, exists := op.unassignedObjects["old"]; exists {
		t.Fatal("expected fully expired entry to be removed")
	}

	got := op.unassignedObjects["mixed"]
	if len(got) != 1 {
		t.Fatalf("expected one non-expired element, got %d", len(got))
	}
}

func TestDynamicFilesChecksum(t *testing.T) {
	t.Parallel()

	t0 := time.Date(2026, 1, 2, 3, 4, 5, 6, time.FixedZone("custom", 2*3600))
	t1 := t0.UTC()

	a := map[string]types.DynamicInputFile{
		"preview": {S3Path: "a/preview.jpg", Date: t0},
		"meta":    {S3Path: "a/meta.json", Date: t1},
	}
	b := map[string]types.DynamicInputFile{
		"meta":    {S3Path: "a/meta.json", Date: t1},
		"preview": {S3Path: "a/preview.jpg", Date: t1},
	}

	sumA := dynamicFilesChecksum(a)
	sumB := dynamicFilesChecksum(b)

	if sumA != sumB {
		t.Fatalf("expected stable checksum regardless of map order/time zone: %q vs %q", sumA, sumB)
	}

	b["meta"] = types.DynamicInputFile{S3Path: "a/meta-v2.json", Date: t1}

	sumC := dynamicFilesChecksum(b)
	if sumC == sumA {
		t.Fatal("expected checksum to change when file content identity changes")
	}
}

func TestValueMap2FilesMap(t *testing.T) {
	t.Parallel()

	exprMan := expressionManager{
		cacheDir: "/cache",
		fileSelectors: map[string]map[string][]string{
			"grp": {
				"typ": {"preview", "meta", "missing"},
			},
		},
	}

	img := image{
		bucket:          "bucket",
		s3Key:           "root/preview.jpg",
		imgGroup:        "grp",
		imgType:         "typ",
		previewCacheKey: "bucket/root/preview.jpg",
		lastModified:    time.Unix(10, 0),
		dynamicInputFiles: map[string]valueWithLastUpdate[types.DynamicInputFile]{
			"meta": {
				value: types.DynamicInputFile{
					S3Path:   "root/meta.json",
					CacheKey: "bucket/root/meta.json",
				},
			},
		},
	}

	got := exprMan.valueMap2FilesMap(img)

	if got[types.ObjectPreview].CacheKey != "/cache/bucket/root/preview.jpg" {
		t.Fatalf("unexpected preview cache key: %q", got[types.ObjectPreview].CacheKey)
	}

	if got["meta"].CacheKey != "/cache/bucket/root/meta.json" {
		t.Fatalf("unexpected dynamic cache key: %q", got["meta"].CacheKey)
	}

	if _, exists := got["missing"]; !exists {
		t.Fatal("expected missing selector to be present as zero value")
	}
}
