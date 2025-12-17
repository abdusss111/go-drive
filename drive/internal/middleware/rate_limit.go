package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter ограничивает количество запросов с одного IP
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*Client
	duration time.Duration
	maxReq   int
}

type Client struct {
	last  time.Time
	count int
}

func NewRateLimiter(duration time.Duration, maxReq int) *RateLimiter {
	rl := &RateLimiter{
		clients:  make(map[string]*Client),
		duration: duration,
		maxReq:   maxReq,
	}

	// Запускаем очистку старых записей
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, client := range rl.clients {
			if now.Sub(client.last) > rl.duration {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Middleware возвращает Gin middleware для ограничения запросов
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		client, exists := rl.clients[ip]
		if !exists {
			client = &Client{last: time.Now()}
			rl.clients[ip] = client
		}

		now := time.Now()
		if now.Sub(client.last) > rl.duration {
			client.count = 0
			client.last = now
		}

		client.count++
		if client.count > rl.maxReq {
			rl.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "rate limit exceeded",
				"code":  429,
			})
			c.Abort()
			return
		}

		rl.mu.Unlock()
		c.Next()
	}
}
