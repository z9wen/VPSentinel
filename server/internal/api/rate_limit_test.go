package api

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/z9wen/VPSentinel/server/internal/auth"
	"github.com/z9wen/VPSentinel/server/internal/ingest"
)

func proxiedLoginRouter(t *testing.T, trustedProxies []string) http.Handler {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	gin.SetMode(gin.TestMode)
	return NewRouter(Dependencies{
		Authenticator:  auth.NewStatic(testToken, "node-one"),
		Store:          ingest.NewMemoryStore(10),
		Management:     &fakeManagementStore{passwordHash: string(hash)},
		Sessions:       auth.NewSessionManager("0123456789abcdef0123456789abcdef", 15*time.Minute, 24*time.Hour),
		Logger:         slog.New(slog.NewJSONHandler(io.Discard, nil)),
		WebhookSecret:  "webhook-secret-123",
		TrustedProxies: trustedProxies,
	})
}

func failedLogin(router http.Handler, peer, realIP string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"admin@example.com","password":"wrong-password"}`))
	request.RemoteAddr = peer + ":40000"
	if realIP != "" {
		request.Header.Set("X-Real-IP", realIP)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestLoginIsRateLimitedPerClientBehindTrustedProxy(t *testing.T) {
	router := proxiedLoginRouter(t, []string{"172.16.0.0/12"})
	const proxy = "172.18.0.5"

	for attempt := 1; attempt <= loginAttemptsPerMinute; attempt++ {
		if response := failedLogin(router, proxy, "203.0.113.10"); response.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d status = %d, want %d", attempt, response.Code, http.StatusUnauthorized)
		}
	}
	limited := failedLogin(router, proxy, "203.0.113.10")
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("status after %d attempts = %d, want %d", loginAttemptsPerMinute, limited.Code, http.StatusTooManyRequests)
	}
	if seconds, err := strconv.Atoi(limited.Header().Get("Retry-After")); err != nil || seconds < 1 || seconds > 60 {
		t.Fatalf("Retry-After = %q, want 1-60 seconds", limited.Header().Get("Retry-After"))
	}

	// Another client behind the same proxy keeps its own budget.
	if response := failedLogin(router, proxy, "203.0.113.20"); response.Code != http.StatusUnauthorized {
		t.Fatalf("second client status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestUntrustedPeerCannotSpoofClientAddress(t *testing.T) {
	router := proxiedLoginRouter(t, []string{"172.16.0.0/12"})
	const attacker = "198.51.100.7"

	for attempt := 1; attempt <= loginAttemptsPerMinute; attempt++ {
		failedLogin(router, attacker, "203.0.113."+strconv.Itoa(attempt))
	}
	if response := failedLogin(router, attacker, "203.0.113.99"); response.Code != http.StatusTooManyRequests {
		t.Fatalf("spoofed X-Real-IP status = %d, want %d", response.Code, http.StatusTooManyRequests)
	}
}
