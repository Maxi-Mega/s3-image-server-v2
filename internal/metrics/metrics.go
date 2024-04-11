package metrics

import (
	"github.com/Maxi-Mega/s3-image-server-v2/config"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const instanceLabelName = "s3_image_server"

type Metrics struct {
	RequestCounter       prometheus.Counter
	RequestDuration      *prometheus.HistogramVec
	S3EventsCounter      *prometheus.CounterVec
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
			Buckets:     prometheus.DefBuckets,
		}, []string{"endpoint", "route", "status_code"}),
		S3EventsCounter: promauto.NewCounterVec(prometheus.CounterOpts{
			Name:        "s3_events_total",
			Help:        "The total number of S3 events received by the server",
			ConstLabels: constLabels,
		}, []string{"bucket"}),
		CacheImagesPerBucket: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name:        "cache_images_number",
			Help:        "The total number of cache images",
			ConstLabels: constLabels,
		}, []string{"bucket"}),
	}
}
