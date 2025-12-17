package server

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// loggerMiddleware provides structured logging; falls back to std log if zap is nil.
func loggerMiddleware(z *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start)

		if z != nil {
			z.Info("request",
				zap.String("method", method),
				zap.String("path", path),
				zap.Int("status", status),
				zap.Duration("duration", duration),
				zap.String("client_ip", c.ClientIP()),
			)
			return
		}

		log.Printf("%s %s -> %d (%s)", method, path, status, duration)
	}
}
