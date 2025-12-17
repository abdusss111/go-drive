package server

import (
	"sync"
	"time"

	"github.com/abduss/godrive/internal/config"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type clientLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimitMiddleware applies IP-based token bucket limiting.
func rateLimitMiddleware(cfg config.RateLimitConfig) gin.HandlerFunc {
	limiters := sync.Map{}
	cleanupTicker := time.NewTicker(5 * time.Minute)

	// Periodically clean old entries.
	go func() {
		for range cleanupTicker.C {
			now := time.Now()
			limiters.Range(func(key, value any) bool {
				cl := value.(clientLimiter)
				if now.Sub(cl.lastSeen) > 10*time.Minute {
					limiters.Delete(key)
				}
				return true
			})
		}
	}()

	newLimiter := func() *rate.Limiter {
		r := rate.Limit(cfg.RPS)
		if cfg.RPS <= 0 {
			r = rate.Inf
		}
		return rate.NewLimiter(r, cfg.Burst)
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		val, ok := limiters.Load(ip)
		if !ok {
			val = clientLimiter{limiter: newLimiter(), lastSeen: time.Now()}
			limiters.Store(ip, val)
		}

		cl := val.(clientLimiter)
		cl.lastSeen = time.Now()
		if !cl.limiter.Allow() {
			c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
			return
		}
		limiters.Store(ip, cl)
		c.Next()
	}
}
