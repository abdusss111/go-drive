package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const correlationIDKey = "correlation_id"

// CorrelationIDMiddleware adds a correlation ID to each request for tracing.
func CorrelationIDMiddleware(log *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get or generate correlation ID
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Store in context
		c.Set(correlationIDKey, correlationID)
		c.Header("X-Correlation-ID", correlationID)

		// Add to logger context
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Log request completion
		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info("HTTP request",
			zap.String("correlation_id", correlationID),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// GetCorrelationID extracts the correlation ID from the Gin context.
func GetCorrelationID(c *gin.Context) string {
	if id, exists := c.Get(correlationIDKey); exists {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return ""
}

