package metrics

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Register(router *gin.Engine, path string) {
	router.GET(path, gin.WrapH(promhttp.Handler()))
}

var HTTPRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "path", "status"},
)

var HTTPRequestDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"method", "path", "status"},
)

var AuthAttemptsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "auth_attempts_total",
		Help: "Count of authentication attempts",
	},
	[]string{"result"}, // success | failure
)

var FileOperationSizeBytes = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "file_operation_size_bytes",
		Help:    "Size of uploaded/downloaded files in bytes",
		Buckets: prometheus.ExponentialBuckets(1024, 2, 10), // 1KB..~
	},
	[]string{"operation"}, // upload | download
)

func InitMetrics() {
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	prometheus.MustRegister(AuthAttemptsTotal)
	prometheus.MustRegister(FileOperationSizeBytes)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		path := c.FullPath()

		c.Next()

		status := fmt.Sprintf("%d", c.Writer.Status())

		HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		HTTPRequestDuration.WithLabelValues(method, path, status).Observe(float64(c.Writer.Size()))
	}
}
