package api

import (
	"crypto/sha256"
	"encoding/hex"
	"math"
	"net/http"
	"strconv"
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

// rateKey decides which requests share a fixed-window budget.
type rateKey func(*gin.Context) string

// clientCredentialKey gives every credential its own budget per client address,
// so one agent cannot exhaust another agent's budget behind the same NAT. The
// credential is hashed so tokens are not kept in process diagnostics.
func clientCredentialKey(c *gin.Context) string {
	sum := sha256.Sum256([]byte(c.GetHeader("Authorization")))
	return c.ClientIP() + ":" + hex.EncodeToString(sum[:8])
}

// clientIPKey limits by address alone, which is what bounds password guessing.
func clientIPKey(c *gin.Context) string {
	return c.ClientIP()
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{limit: limit, window: window, clients: make(map[string]rateWindow)}
}

func (l *rateLimiter) middleware(key rateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientKey := key(c)
		now := time.Now()

		l.mu.Lock()
		entry := l.clients[clientKey]
		if entry.started.IsZero() || now.Sub(entry.started) >= l.window {
			entry = rateWindow{started: now}
		}
		entry.count++
		l.clients[clientKey] = entry
		limited := entry.count > l.limit
		retryAfter := l.window - now.Sub(entry.started)
		if len(l.clients) > 10_000 {
			for client, candidate := range l.clients {
				if now.Sub(candidate.started) >= l.window {
					delete(l.clients, client)
				}
			}
		}
		l.mu.Unlock()

		if limited {
			c.Header("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
			writeError(c, http.StatusTooManyRequests, "rate_limited", "request rate limit exceeded")
			return
		}
		c.Next()
	}
}
