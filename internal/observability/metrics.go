package observability

import (
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const instanceLabelName = "s3_image_server"

type Metrics struct {
	RequestCounter       prometheus.Counter
	RequestDuration      *prometheus.HistogramVec
	S3EventsCounter      *prometheus.CounterVec
	S3ListDuration       *prometheus.HistogramVec
	CacheImagesPerBucket *prometheus.GaugeVec
}

func New(cfg config.Monitoring) *Metrics {
	constLabels := map[string]string{
		instanceLabelName: cfg.PrometheusInstanceLabel,
	}

	return &Metrics{
		RequestCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name:        "request_total",
			Help:        "The total number of requests received by the server",
			ConstLabels: constLabels,
		}),
		RequestDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:        "request_duration",
			Help:        "The duration of requests being handled by the server",
			ConstLabels: constLabels,
			Buckets:     prometheus.ExponentialBucketsRange((100 * time.Microsecond).Seconds(), (100 * time.Millisecond).Seconds(), 10),
		}, []string{"endpoint", "route", "status_code"}),
		S3EventsCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "s3_events_total",
			Help:        "The total number of S3 events received by the server",
			ConstLabels: constLabels,
		}, []string{"bucket"}),
		S3ListDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:        "s3_list_duration_seconds",
			Help:        "The duration spent listing objects from S3",
			ConstLabels: constLabels,
			Buckets:     prometheus.ExponentialBucketsRange((100 * time.Millisecond).Seconds(), (10 * time.Second).Seconds(), 10),
		}, []string{"bucket"}),
		CacheImagesPerBucket: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name:        "cache_images_number",
			Help:        "The total number of cache images",
			ConstLabels: constLabels,
		}, []string{"bucket"}),
	}
}
