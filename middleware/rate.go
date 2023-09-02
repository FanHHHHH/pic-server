package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

var (
	visitors = make(map[string]*rate.Limiter)
	mu       sync.Mutex
)

func getVisitor(ip string, filepath string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	id := ip + filepath

	limiter, exists := visitors[id]
	if !exists {
		limiter = rate.NewLimiter(rate.Every(30*time.Millisecond), 1)
		visitors[id] = limiter
	}
	return limiter
}

func RateLimit(c *gin.Context) {

	limiter := getVisitor(c.ClientIP(), c.Param("filepath"))
	if !limiter.Allow() {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
			"error": "rate limit exceeded",
		})
		return
	}
	c.Next()
}
