package server

import (
	"context"
	"net/url"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/observability"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/s3"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"

	"github.com/prometheus/client_golang/prometheus"
)

func TestApplyRegeneratedSignedURLs(t *testing.T) {
	t.Parallel()

	baseTime := time.Date(2026, 4, 4, 12, 0, 0, 0, time.UTC)

	makeSignedURLValue := func(lastUpdate, generationDate time.Time, value string) valueWithLastUpdate[signedURL] {
		return valueWithLastUpdate[signedURL]{
			value: signedURL{
				value:          value,
				generationDate: generationDate,
			},
			lastUpdate: lastUpdate,
		}
	}

	makeImage := func(entries map[string]valueWithLastUpdate[signedURL]) image {
		return image{
			name:              "img-a",
			baseDir:           "img-a",
			bucket:            "bucket-a",
			imgGroup:          "grp",
			imgType:           "typ",
			lastModified:      baseTime,
			targets:           map[string]valueWithLastUpdate[string]{},
			dynamicInputFiles: map[string]valueWithLastUpdate[types.DynamicInputFile]{},
			linksFromCache:    map[string]valueWithLastUpdate[string]{},
			signedURLs:        entries,
		}
	}

	cases := []struct {
		name          string
		cacheBuckets  map[string]*bucketCache
		newURLs       map[string]map[string]map[string]valueWithLastUpdate[signedURL]
		wantTotal     int
		wantURL       string
		wantUpdatedAt time.Time
	}{
		{
			name: "applies only matching current version",
			cacheBuckets: map[string]*bucketCache{
				"bucket-a": {
					bucket: "bucket-a",
					images: map[string]image{
						"img-a": makeImage(map[string]valueWithLastUpdate[signedURL]{
							"keep-current": makeSignedURLValue(
								baseTime.Add(2*time.Hour),
								baseTime.Add(2*time.Hour),
								"https://example.test/current",
							),
							"same-version": makeSignedURLValue(
								baseTime.Add(1*time.Hour),
								baseTime.Add(1*time.Hour),
								"https://example.test/old",
							),
						}),
					},
				},
			},
			newURLs: map[string]map[string]map[string]valueWithLastUpdate[signedURL]{
				"bucket-a": {
					"img-a": {
						"missing-key": makeSignedURLValue(
							baseTime.Add(1*time.Hour),
							baseTime.Add(3*time.Hour),
							"https://example.test/missing",
						),
						"keep-current": makeSignedURLValue(
							baseTime.Add(1*time.Hour),
							baseTime.Add(3*time.Hour),
							"https://example.test/stale-version",
						),
						"same-version": makeSignedURLValue(
							baseTime.Add(1*time.Hour),
							baseTime.Add(3*time.Hour),
							"https://example.test/renewed",
						),
					},
				},
			},
			wantTotal:     1,
			wantURL:       "https://example.test/renewed",
			wantUpdatedAt: baseTime.Add(3 * time.Hour),
		},
		{
			name: "does not overwrite a newer URL for the same object version",
			cacheBuckets: map[string]*bucketCache{
				"bucket-a": {
					bucket: "bucket-a",
					images: map[string]image{
						"img-a": makeImage(map[string]valueWithLastUpdate[signedURL]{
							"same-version": makeSignedURLValue(
								baseTime.Add(1*time.Hour),
								baseTime.Add(5*time.Hour),
								"https://example.test/current-newer",
							),
						}),
					},
				},
			},
			newURLs: map[string]map[string]map[string]valueWithLastUpdate[signedURL]{
				"bucket-a": {
					"img-a": {
						"same-version": makeSignedURLValue(
							baseTime.Add(1*time.Hour),
							baseTime.Add(4*time.Hour),
							"https://example.test/older-renewal",
						),
					},
				},
			},
			wantURL:       "https://example.test/current-newer",
			wantUpdatedAt: baseTime.Add(5 * time.Hour),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := Server{
				cache: &cache{
					buckets: tc.cacheBuckets,
				},
			}

			gotTotal := srv.applyRegeneratedSignedURLs(tc.newURLs)
			if gotTotal != tc.wantTotal {
				t.Fatalf("unexpected updated URL count: got %d, want %d", gotTotal, tc.wantTotal)
			}

			got := srv.cache.buckets["bucket-a"].images["img-a"].signedURLs["same-version"]
			if got.value.value != tc.wantURL {
				t.Fatalf("unexpected final signed URL: got %q, want %q", got.value.value, tc.wantURL)
			}

			if !got.value.generationDate.Equal(tc.wantUpdatedAt) {
				t.Fatalf("unexpected generation date: got %s, want %s", got.value.generationDate, tc.wantUpdatedAt)
			}
		})
	}
}

func TestRunSignedURLRegenerationLoop(t *testing.T) {
	t.Parallel()

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		startTime := time.Now()
		firstTickTime := startTime.Add(s3.SignedURLRegenerationPeriod)
		deadline := firstTickTime.Add(24 * time.Hour)

		renewGenerationDate := startTime.Add(-(s3.SignedURLLifetime + time.Hour))
		keepGenerationDate := deadline.Add(24 * time.Hour).Add(-s3.SignedURLLifetime)
		boundaryGenerationDate := deadline.Add(-s3.SignedURLLifetime)
		objectLastModified := startTime.Add(-2 * time.Hour)

		calls := make(map[string]int)

		s3Client := S3ClientMock{
			GenerateSignedURLFn: func(_ context.Context, bucket, objectKey string) (*url.URL, error) {
				if bucket != "bucket-a" {
					t.Fatalf("GenerateSignedURL called with bucket %q, want %q", bucket, "bucket-a")
				}

				calls[objectKey]++

				switch objectKey {
				case "renew-a.tif":
					return url.Parse("https://signed.example.test/renew-a.tif?new=1")
				case "renew-b.tif":
					return url.Parse("https://signed.example.test/renew-b.tif?new=1")
				default:
					t.Fatalf("unexpected GenerateSignedURL call for %q", objectKey)

					return nil, nil //nolint: nilnil
				}
			},
		}

		srv := Server{
			gatherer: &observability.Metrics{
				S3SignedURLRegenCounter: prometheus.NewCounter(prometheus.CounterOpts{
					Name: "test_s3_signed_url_regen_total",
					Help: "test counter",
				}),
			},
			s3Client: s3Client,
			cache: &cache{
				buckets: map[string]*bucketCache{
					"bucket-a": {
						bucket:   "bucket-a",
						s3Client: s3Client,
						images: map[string]image{
							"img-a": {
								name:              "img-a",
								baseDir:           "img-a",
								bucket:            "bucket-a",
								imgGroup:          "grp",
								imgType:           "typ",
								lastModified:      objectLastModified,
								targets:           map[string]valueWithLastUpdate[string]{},
								dynamicInputFiles: map[string]valueWithLastUpdate[types.DynamicInputFile]{},
								linksFromCache:    map[string]valueWithLastUpdate[string]{},
								signedURLs: map[string]valueWithLastUpdate[signedURL]{
									"renew-a.tif": {
										value: signedURL{
											value:          "https://signed.example.test/renew-a.tif?old=1",
											generationDate: renewGenerationDate,
										},
										lastUpdate: objectLastModified,
									},
									"keep.tif": {
										value: signedURL{
											value:          "https://signed.example.test/keep.tif?old=1",
											generationDate: keepGenerationDate,
										},
										lastUpdate: objectLastModified,
									},
									"boundary.tif": {
										value: signedURL{
											value:          "https://signed.example.test/boundary.tif?old=1",
											generationDate: boundaryGenerationDate,
										},
										lastUpdate: objectLastModified,
									},
								},
							},
							"img-b": {
								name:              "img-b",
								baseDir:           "img-b",
								bucket:            "bucket-a",
								imgGroup:          "grp",
								imgType:           "typ",
								lastModified:      objectLastModified,
								targets:           map[string]valueWithLastUpdate[string]{},
								dynamicInputFiles: map[string]valueWithLastUpdate[types.DynamicInputFile]{},
								linksFromCache:    map[string]valueWithLastUpdate[string]{},
								signedURLs: map[string]valueWithLastUpdate[signedURL]{
									"renew-b.tif": {
										value: signedURL{
											value:          "https://signed.example.test/renew-b.tif?old=1",
											generationDate: renewGenerationDate,
										},
										lastUpdate: objectLastModified,
									},
								},
							},
						},
					},
				},
			},
		}

		go srv.runSignedURLRegenerationLoop(ctx)

		synctest.Wait()

		time.Sleep(s3.SignedURLRegenerationPeriod)
		synctest.Wait()

		imgA := srv.cache.buckets["bucket-a"].images["img-a"].signedURLs
		imgB := srv.cache.buckets["bucket-a"].images["img-b"].signedURLs

		if got := imgA["renew-a.tif"]; got.value.value != "https://signed.example.test/renew-a.tif?new=1" {
			t.Fatalf("unexpected renewed URL for renew-a.tif: got %q", got.value.value)
		}

		if got := imgB["renew-b.tif"]; got.value.value != "https://signed.example.test/renew-b.tif?new=1" {
			t.Fatalf("unexpected renewed URL for renew-b.tif: got %q", got.value.value)
		}

		if got := imgA["keep.tif"]; got.value.value != "https://signed.example.test/keep.tif?old=1" {
			t.Fatalf("keep.tif should not have been renewed: got %q", got.value.value)
		}

		if got := imgA["boundary.tif"]; got.value.value != "https://signed.example.test/boundary.tif?old=1" {
			t.Fatalf("boundary.tif should not have been renewed: got %q", got.value.value)
		}

		for name, got := range map[string]valueWithLastUpdate[signedURL]{
			"renew-a.tif": imgA["renew-a.tif"],
			"renew-b.tif": imgB["renew-b.tif"],
		} {
			if !got.lastUpdate.Equal(objectLastModified) {
				t.Fatalf("unexpected last update for %s: got %s, want %s", name, got.lastUpdate, objectLastModified)
			}

			if !got.value.generationDate.After(renewGenerationDate) {
				t.Fatalf("expected generation date to move forward for %s: old %s, new %s", name, renewGenerationDate, got.value.generationDate)
			}
		}

		if calls["renew-a.tif"] != 1 || calls["renew-b.tif"] != 1 {
			t.Fatalf("unexpected renewal calls: %#v", calls)
		}

		if len(calls) != 2 {
			t.Fatalf("unexpected number of renewed files: got %d, want 2", len(calls))
		}

		cancel()
		synctest.Wait()
	})
}
