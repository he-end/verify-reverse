package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func RateLimitRegister(maxRequests int, window time.Duration) gin.HandlerFunc {
	mu := &sync.Mutex{}
	buckets := map[string]*rateBucket{}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		mu.Lock()
		b, ok := buckets[ip]
		if !ok || now.After(b.resetAt) {
			b = &rateBucket{resetAt: now.Add(window)}
			buckets[ip] = b
		}
		mu.Unlock()

		b.Lock()
		if b.count >= maxRequests {
			b.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "terlalu banyak percobaan, coba lagi nanti",
			})
			return
		}
		b.count++
		b.Unlock()
		c.Next()
	}
}

type rateBucket struct {
	sync.Mutex
	count   int
	resetAt time.Time
}
