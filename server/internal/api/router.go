package api

import (
	"encoding/json"
	"errors"
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
)

type Dependencies struct {
	Authenticator auth.Authenticator
	Store         ingest.Store
	Logger        *slog.Logger
	Version       string
	MaxBodyBytes  int64
	Now           func() time.Time
}

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
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Recovery(), requestLogger(dependencies.Logger))

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
	return router
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
