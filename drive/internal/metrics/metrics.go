package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Register attaches the Prometheus metrics endpoint to the router.
func Register(router *gin.Engine, path string) {
	router.GET(path, gin.WrapH(promhttp.Handler()))
}
