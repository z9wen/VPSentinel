package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/z9wen/VPSentinel/protocol"
	"github.com/z9wen/VPSentinel/server/internal/auth"
	"github.com/z9wen/VPSentinel/server/internal/ingest"
	"github.com/z9wen/VPSentinel/server/internal/management"
)

type Dependencies struct {
	Authenticator auth.Authenticator
	Store         ingest.Store
	Logger        *slog.Logger
	Version       string
	MaxBodyBytes  int64
	Now           func() time.Time
	Management    management.Store
	Sessions      *auth.SessionManager
	AllowedOrigin string
	CookieSecure  bool
	WebhookSecret string
	GrafanaURL    string
	// TrustedProxies are the peers allowed to supply the client address in
	// X-Real-IP. Without them every client behind the bundled web proxy would
	// share one address and one rate-limit budget.
	TrustedProxies []string
}

// loginAttemptsPerMinute bounds online password guessing per client address.
const loginAttemptsPerMinute = 10

func NewRouter(dependencies Dependencies) *gin.Engine {
	if dependencies.Authenticator == nil {
		panic("api authenticator is required")
	}
	if dependencies.Store == nil {
		panic("api store is required")
	}
	if dependencies.Logger == nil {
		dependencies.Logger = slog.Default()
	}
	if dependencies.MaxBodyBytes <= 0 {
		dependencies.MaxBodyBytes = 1 << 20
	}
	if dependencies.Now == nil {
		dependencies.Now = time.Now
	}

	router := gin.New()
	// Only X-Real-IP is honoured: the bundled proxy overwrites it, whereas
	// X-Forwarded-For is appended to and can carry client-supplied values.
	router.RemoteIPHeaders = []string{"X-Real-IP"}
	if err := router.SetTrustedProxies(dependencies.TrustedProxies); err != nil {
		panic(fmt.Sprintf("api trusted proxies: %v", err))
	}
	router.Use(gin.Recovery(), securityHeaders(), cors(dependencies.AllowedOrigin), newRateLimiter(240, time.Minute).middleware(clientCredentialKey), requestLogger(dependencies.Logger))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/readyz", func(c *gin.Context) {
		if err := dependencies.Store.Ready(c.Request.Context()); err != nil {
			writeError(c, http.StatusServiceUnavailable, "not_ready", "server dependencies are not ready")
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": dependencies.Version})
	})

	router.POST("/api/v1/agent/metrics", dependencies.ingestMetrics)
	if dependencies.Management != nil {
		router.POST("/api/v1/agent/heartbeat", dependencies.agentHeartbeat)
	}
	if dependencies.Management != nil && dependencies.Sessions != nil {
		router.POST("/api/v1/auth/login", newRateLimiter(loginAttemptsPerMinute, time.Minute).middleware(clientIPKey), dependencies.login)
		router.POST("/api/v1/auth/refresh", dependencies.refresh)
		router.POST("/api/v1/auth/logout", dependencies.logout)
		router.POST("/api/v1/integrations/grafana/webhook", dependencies.grafanaWebhook)

		admin := router.Group("/api/v1")
		admin.Use(dependencies.requireAdmin())
		admin.GET("/nodes", dependencies.listNodes)
		admin.POST("/nodes", dependencies.createNode)
		admin.GET("/nodes/:id", dependencies.getNode)
		admin.DELETE("/nodes/:id", dependencies.deleteNode)
		admin.POST("/nodes/:id/tokens", dependencies.createNodeToken)
		admin.GET("/nodes/:id/tokens", dependencies.listNodeTokens)
		admin.DELETE("/nodes/:id/tokens/:tokenId", dependencies.revokeNodeToken)
		admin.GET("/nodes/:id/metrics/latest", dependencies.latestMetric)
		admin.GET("/alerts", dependencies.listAlerts)
		admin.GET("/alerts/:id", dependencies.getAlert)
	}
	return router
}

func securityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")
		c.Next()
	}
}

func cors(origin string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if origin != "" && c.GetHeader("Origin") == origin {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-VPSentinel-Webhook-Secret")
			c.Header("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			c.Header("Vary", "Origin")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func (d Dependencies) ingestMetrics(c *gin.Context) {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(c, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
		return
	}

	token, ok := bearerToken(c.GetHeader("Authorization"))
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
		return
	}
	nodeID, err := d.Authenticator.Authenticate(c.Request.Context(), token)
	if err != nil {
		writeError(c, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, d.MaxBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	var batch protocol.MetricBatch
	if err := decoder.Decode(&batch); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			writeError(c, http.StatusRequestEntityTooLarge, "body_too_large", "request body exceeds the configured limit")
			return
		}
		writeError(c, http.StatusBadRequest, "invalid_json", "request body must contain one valid metric batch")
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(c, http.StatusBadRequest, "invalid_json", "request body must contain exactly one JSON object")
		return
	}

	receivedAt := d.Now().UTC()
	if err := ingest.ValidateBatch(batch, receivedAt); err != nil {
		writeError(c, http.StatusUnprocessableEntity, "invalid_batch", err.Error())
		return
	}
	acceptedSamples, err := d.Store.StoreBatch(c.Request.Context(), nodeID, batch, receivedAt)
	if err != nil {
		d.Logger.Error("store metric batch", "error", err, "node_id", nodeID, "batch_id", batch.BatchID)
		writeError(c, http.StatusServiceUnavailable, "storage_unavailable", "metric storage is temporarily unavailable")
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"batch_id":         batch.BatchID,
		"accepted_samples": acceptedSamples,
	})
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimPrefix(header, prefix)
	if token == "" || strings.TrimSpace(token) != token || strings.ContainsAny(token, " \t\r\n") {
		return "", false
	}
	return token, true
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()
		logger.Info("http request",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"client_ip", c.ClientIP(),
			"duration", time.Since(startedAt),
		)
	}
}

func writeError(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
