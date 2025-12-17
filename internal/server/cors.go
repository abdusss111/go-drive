package server

import (
	"time"

	"github.com/abduss/godrive/internal/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func newCORS(cfg config.CORSConfig) gin.HandlerFunc {
	c := cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           cfg.MaxAge,
	}
	if len(c.AllowOrigins) == 0 {
		c.AllowAllOrigins = true
	}
	if len(c.AllowMethods) == 0 {
		c.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(c.AllowHeaders) == 0 {
		c.AllowHeaders = []string{"Authorization", "Content-Type", "Accept"}
	}
	// Ensure MaxAge non-zero.
	if c.MaxAge == 0 {
		c.MaxAge = 12 * time.Hour
	}
	return cors.New(c)
}
