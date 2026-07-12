package server

import (
	"testing"
	"time"
)

var (
	benchmarkGetCacheValue any
	benchmarkGetCacheOK    bool
)

func BenchmarkExpressionManagerGetCache(b *testing.B) {
	const (
		imgBucket = "bucket"
		imgKey    = "path/to/preview.jpg"
		exprName  = "expr"
		sum       = "sum"
	)

	cases := []struct {
		name       string
		entry      exprCacheEntry
		requestSum string
		wantOK     bool
	}{
		{
			name: "hit",
			entry: exprCacheEntry{
				sum:   sum,
				ts:    time.Now(),
				value: "cached value",
			},
			requestSum: sum,
			wantOK:     true,
		},
		{
			name: "checksum_miss",
			entry: exprCacheEntry{
				sum:   "other sum",
				ts:    time.Now(),
				value: "cached value",
			},
			requestSum: sum,
			wantOK:     false,
		},
		{
			name: "expired_entry",
			entry: exprCacheEntry{
				sum:   sum,
				ts:    time.Now().Add(-exprCacheTTL - time.Second),
				value: "cached value",
			},
			requestSum: sum,
			wantOK:     false,
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			exprMan := &expressionManager{
				cacheSums: map[exprCacheKey]exprCacheEntry{
					{bucket: imgBucket, s3key: imgKey, exprName: exprName}: tc.entry,
				},
			}

			b.ReportAllocs()
			b.ResetTimer()

			var (
				value any
				ok    bool
			)

			for range b.N {
				value, ok = exprMan.getCache(imgBucket, imgKey, exprName, tc.requestSum)
			}

			benchmarkGetCacheValue = value
			benchmarkGetCacheOK = ok

			if ok != tc.wantOK {
				b.Fatalf("Expected cache result %t, got %t", tc.wantOK, ok)
			}
		})
	}
}
