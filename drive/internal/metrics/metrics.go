package metrics

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Register(router *gin.Engine, path string) {
	router.GET(path, gin.WrapH(promhttp.Handler()))
}

var (
	metricsOnce sync.Once

	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	AuthAttemptsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Authentication attempts",
		},
		[]string{"result"}, // success | failure
	)

	FileOperationSizeBytes = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "file_operation_size_bytes",
			Help:    "Upload and download file sizes",
			Buckets: prometheus.ExponentialBuckets(1024, 2, 10),
		},
		[]string{"operation"}, // upload | download
	)
)

func InitMetrics() {
	metricsOnce.Do(func() {
		prometheus.MustRegister(
			HTTPRequestsTotal,
			HTTPRequestDuration,
			AuthAttemptsTotal,
			FileOperationSizeBytes,
		)
	})
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		c.Next()

		status := fmt.Sprintf("%d", c.Writer.Status())

		HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		HTTPRequestDuration.
			WithLabelValues(method, path, status).
			Observe(float64(c.Writer.Size()))
	}
}
