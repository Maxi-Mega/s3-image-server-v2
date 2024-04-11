package web

import (
	"strconv"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/metrics"

	"github.com/gin-gonic/gin"
)

type endpoint string

const (
	endpointFront = "front"
	endpointAPI   = "api"
)

func metricsMiddleware(gatherer *metrics.Metrics, endpoint endpoint) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		statusCode := strconv.Itoa(c.Writer.Status())
		route := c.FullPath()

		gatherer.RequestDuration.WithLabelValues(string(endpoint), route, statusCode).Observe(duration.Seconds())
		gatherer.RequestCounter.Inc()
	}
}
