package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const readinessTimeout = 5 * time.Second

func registerHealthRoutes(router *gin.Engine, deps Dependencies) {
	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), readinessTimeout)
		defer cancel()

		if err := deps.DB.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "degraded",
				"component": "postgres",
				"error":     err.Error(),
			})
			return
		}

		if err := checkMinIO(ctx, deps); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "degraded",
				"component": "minio",
				"error":     err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func checkMinIO(ctx context.Context, deps Dependencies) error {
	_, err := deps.ObjectStore.ListBuckets(ctx)
	return err
}
