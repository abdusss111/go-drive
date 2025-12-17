package server

import (
	"net/http"

	"github.com/abduss/godrive/internal/config"
	"github.com/gin-gonic/gin"
)

func securityHeadersMiddleware(cfg config.SecurityConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.ContentTypeNosniff {
			c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		}
		if cfg.FrameOptions != "" {
			c.Writer.Header().Set("X-Frame-Options", cfg.FrameOptions)
		}
		if cfg.XSSProtection != "" {
			c.Writer.Header().Set("X-XSS-Protection", cfg.XSSProtection)
		}
		if cfg.ReferrerPolicy != "" {
			c.Writer.Header().Set("Referrer-Policy", cfg.ReferrerPolicy)
		}
		if cfg.HSTSMaxAge > 0 && c.Request.TLS != nil {
			hsts := "max-age=" + cfg.HSTSMaxAge.String()
			if cfg.HSTSIncludeSubdomains {
				hsts += "; includeSubDomains"
			}
			if cfg.HSTSPreload {
				hsts += "; preload"
			}
			c.Writer.Header().Set("Strict-Transport-Security", hsts)
		}
		c.Header("Cache-Control", "no-store")
		c.Next()
	}
}

// limitBodyMiddleware limits request body size to avoid abuse.
func limitBodyMiddleware(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxBytes && c.Request.ContentLength > 0 {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request too large"})
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
