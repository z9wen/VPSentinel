package api

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateWindow struct {
	started time.Time
	count   int
}

type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]rateWindow
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{limit: limit, window: window, clients: make(map[string]rateWindow)}
}

func (l *rateLimiter) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Hash the credential before using it as a map key so tokens are not kept
		// in process diagnostics. The IP keeps unauthenticated traffic separate.
		sum := sha256.Sum256([]byte(c.GetHeader("Authorization")))
		key := c.ClientIP() + ":" + hex.EncodeToString(sum[:8])
		now := time.Now()

		l.mu.Lock()
		entry := l.clients[key]
		if entry.started.IsZero() || now.Sub(entry.started) >= l.window {
			entry = rateWindow{started: now}
		}
		entry.count++
		l.clients[key] = entry
		limited := entry.count > l.limit
		if len(l.clients) > 10_000 {
			for client, candidate := range l.clients {
				if now.Sub(candidate.started) >= l.window {
					delete(l.clients, client)
				}
			}
		}
		l.mu.Unlock()

		if limited {
			c.Header("Retry-After", "60")
			writeError(c, http.StatusTooManyRequests, "rate_limited", "request rate limit exceeded")
			return
		}
		c.Next()
	}
}
