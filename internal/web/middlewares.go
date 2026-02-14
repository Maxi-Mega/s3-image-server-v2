package web

import (
	"strconv"
	"strings"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/observability"

	"github.com/gin-gonic/gin"
)

type endpoint string

const (
	endpointFront = "front"
	endpointAPI   = "api"
)

func metricsMiddleware(gatherer *observability.Metrics, endpoint endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		if c.Request.URL.Path == "/metrics" {
			return
		}

		duration := time.Since(start)
		statusCode := strconv.Itoa(c.Writer.Status())
		route := processRouteForPrometheus(c.Request.RequestURI)

		gatherer.RequestDuration.WithLabelValues(string(endpoint), route, statusCode).Observe(duration.Seconds())
		gatherer.RequestCounter.Inc()
	}
}

func processRouteForPrometheus(uri string) string {
	switch {
	case strings.HasPrefix(uri, "/api/cache/"):
		return "/api/cache"
	default:
		return uri
	}
}
