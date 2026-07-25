package api

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/z9wen/VPSentinel/server/internal/auth"
	"github.com/z9wen/VPSentinel/server/internal/management"
)

const refreshCookieName = "vpsentinel_refresh"

func (d Dependencies) decodeJSON(c *gin.Context, destination any) bool {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, d.MaxBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "request body must contain valid JSON")
		return false
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeError(c, http.StatusBadRequest, "invalid_json", "request body must contain exactly one JSON object")
		return false
	}
	return true
}

func (d Dependencies) login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if !d.decodeJSON(c, &input) {
		return
	}
	hash, err := d.Management.PasswordHash(c.Request.Context(), input.Email)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)) != nil {
		writeError(c, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
		return
	}
	pair, err := d.Sessions.IssuePair(strings.ToLower(input.Email), d.Now().UTC())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "session_error", "could not create session")
		return
	}
	if err := d.Management.CreateRefreshSession(c.Request.Context(), input.Email, auth.HashToken(pair.RefreshToken), pair.RefreshExpiresAt); err != nil {
		writeError(c, http.StatusServiceUnavailable, "session_error", "could not persist session")
		return
	}
	d.setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	c.JSON(http.StatusOK, accessResponse(pair))
}

func (d Dependencies) refresh(c *gin.Context) {
	token, err := c.Cookie(refreshCookieName)
	if err != nil {
		if c.Request.ContentLength == 0 {
			writeError(c, http.StatusUnauthorized, "refresh_token_required", "refresh token is required")
			return
		}
		var input struct {
			RefreshToken string `json:"refresh_token"`
		}
		if !d.decodeJSON(c, &input) {
			return
		}
		token = input.RefreshToken
	}
	email, err := d.Sessions.VerifyRefresh(token, d.Now().UTC())
	if err != nil {
		writeError(c, http.StatusUnauthorized, "invalid_refresh_token", "refresh token is expired or invalid")
		return
	}
	pair, err := d.Sessions.IssuePair(email, d.Now().UTC())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "session_error", "could not rotate session")
		return
	}
	storedEmail, err := d.Management.RotateRefreshSession(c.Request.Context(), auth.HashToken(token), auth.HashToken(pair.RefreshToken), pair.RefreshExpiresAt)
	if err != nil || storedEmail != email {
		writeError(c, http.StatusUnauthorized, "invalid_refresh_token", "refresh token was revoked or already used")
		return
	}
	d.setRefreshCookie(c, pair.RefreshToken, pair.RefreshExpiresAt)
	c.JSON(http.StatusOK, accessResponse(pair))
}

func (d Dependencies) logout(c *gin.Context) {
	if token, err := c.Cookie(refreshCookieName); err == nil {
		_ = d.Management.RevokeRefreshSession(c.Request.Context(), auth.HashToken(token))
	}
	http.SetCookie(c.Writer, &http.Cookie{Name: refreshCookieName, Value: "", Path: "/api/v1/auth", MaxAge: -1, HttpOnly: true, Secure: d.CookieSecure, SameSite: http.SameSiteStrictMode})
	c.Status(http.StatusNoContent)
}

func (d Dependencies) setRefreshCookie(c *gin.Context, token string, expiresAt time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{Name: refreshCookieName, Value: token, Path: "/api/v1/auth", Expires: expiresAt, HttpOnly: true, Secure: d.CookieSecure, SameSite: http.SameSiteStrictMode})
}

func accessResponse(pair auth.TokenPair) gin.H {
	// Refresh tokens stay in an HttpOnly cookie and are never exposed to JavaScript.
	return gin.H{"access_token": pair.AccessToken, "token_type": pair.TokenType, "expires_at": pair.AccessExpiresAt}
}

func (d Dependencies) requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			writeError(c, http.StatusUnauthorized, "unauthorized", "access token required")
			return
		}
		email, err := d.Sessions.VerifyAccess(token, d.Now().UTC())
		if err != nil {
			writeError(c, http.StatusUnauthorized, "unauthorized", "access token is expired or invalid")
			return
		}
		c.Set("admin_email", email)
		c.Next()
	}
}

func (d Dependencies) listNodes(c *gin.Context) {
	nodes, err := d.Management.ListNodes(c.Request.Context(), d.Now().UTC())
	if err != nil {
		writeError(c, http.StatusServiceUnavailable, "storage_unavailable", "could not load nodes")
		return
	}
	for index := range nodes {
		nodes[index].GrafanaURL = d.grafanaNodeURL(nodes[index].ID)
	}
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}

func (d Dependencies) getNode(c *gin.Context) {
	node, err := d.Management.GetNode(c.Request.Context(), c.Param("id"), d.Now().UTC())
	if !handleManagementError(c, err, "node") {
		return
	}
	node.GrafanaURL = d.grafanaNodeURL(node.ID)
	c.JSON(http.StatusOK, node)
}

func (d Dependencies) createNode(c *gin.Context) {
	var input struct {
		Name           string  `json:"name"`
		Hostname       string  `json:"hostname"`
		Environment    string  `json:"environment"`
		DisplayAddress *string `json:"display_address"`
	}
	if !d.decodeJSON(c, &input) {
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" || len(input.Name) > 100 || len(input.Hostname) > 255 || len(input.Environment) > 50 {
		writeError(c, http.StatusUnprocessableEntity, "invalid_node", "name is required and node fields must be within their size limits")
		return
	}
	id, err := randomID(12)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "id_error", "could not create node ID")
		return
	}
	node, err := d.Management.CreateNode(c.Request.Context(), management.Node{ID: id, Name: input.Name, Hostname: strings.TrimSpace(input.Hostname), Environment: strings.TrimSpace(input.Environment), DisplayAddress: input.DisplayAddress})
	if err != nil {
		writeError(c, http.StatusServiceUnavailable, "storage_unavailable", "could not create node")
		return
	}
	node.GrafanaURL = d.grafanaNodeURL(node.ID)
	c.JSON(http.StatusCreated, node)
}

func (d Dependencies) deleteNode(c *gin.Context) {
	if !handleManagementError(c, d.Management.DeleteNode(c.Request.Context(), c.Param("id")), "node") {
		return
	}
	c.Status(http.StatusNoContent)
}

func (d Dependencies) createNodeToken(c *gin.Context) {
	var input struct {
		Label string `json:"label"`
	}
	if !d.decodeJSON(c, &input) {
		return
	}
	if input.Label == "" {
		input.Label = "agent"
	}
	if len(input.Label) > 100 {
		writeError(c, http.StatusUnprocessableEntity, "invalid_token_label", "token label must not exceed 100 characters")
		return
	}
	if _, err := d.Management.GetNode(c.Request.Context(), c.Param("id"), d.Now().UTC()); !handleManagementError(c, err, "node") {
		return
	}
	plain, err := randomSecret(32)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "token_error", "could not generate token")
		return
	}
	token, err := d.Management.CreateToken(c.Request.Context(), c.Param("id"), input.Label, auth.HashToken(plain))
	if err != nil {
		writeError(c, http.StatusServiceUnavailable, "storage_unavailable", "could not create token")
		return
	}
	// This is the only response that contains the plaintext agent token.
	c.JSON(http.StatusCreated, gin.H{"token": plain, "metadata": token})
}

func (d Dependencies) revokeNodeToken(c *gin.Context) {
	tokenID, err := strconv.ParseInt(c.Param("tokenId"), 10, 64)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_token_id", "token ID must be an integer")
		return
	}
	if !handleManagementError(c, d.Management.RevokeToken(c.Request.Context(), c.Param("id"), tokenID), "token") {
		return
	}
	c.Status(http.StatusNoContent)
}

func (d Dependencies) listNodeTokens(c *gin.Context) {
	if _, err := d.Management.GetNode(c.Request.Context(), c.Param("id"), d.Now().UTC()); !handleManagementError(c, err, "node") {
		return
	}
	tokens, err := d.Management.ListTokens(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeError(c, http.StatusServiceUnavailable, "storage_unavailable", "could not load tokens")
		return
	}
	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

func (d Dependencies) latestMetric(c *gin.Context) {
	metric, err := d.Management.LatestMetric(c.Request.Context(), c.Param("id"))
	if !handleManagementError(c, err, "metric") {
		return
	}
	c.JSON(http.StatusOK, metric)
}

func (d Dependencies) listAlerts(c *gin.Context) {
	active := c.Query("active") == "true"
	limit := 100
	if requested, err := strconv.Atoi(c.Query("limit")); err == nil && requested > 0 && requested <= 500 {
		limit = requested
	}
	alerts, err := d.Management.ListAlerts(c.Request.Context(), active, limit)
	if err != nil {
		writeError(c, http.StatusServiceUnavailable, "storage_unavailable", "could not load alerts")
		return
	}
	c.JSON(http.StatusOK, gin.H{"alerts": alerts})
}

func (d Dependencies) getAlert(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_alert_id", "alert ID must be an integer")
		return
	}
	alert, err := d.Management.GetAlert(c.Request.Context(), id)
	if !handleManagementError(c, err, "alert") {
		return
	}
	c.JSON(http.StatusOK, alert)
}

func (d Dependencies) agentHeartbeat(c *gin.Context) {
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
	var input struct {
		Hostname     string `json:"hostname"`
		AgentVersion string `json:"agent_version"`
	}
	if !d.decodeJSON(c, &input) {
		return
	}
	if err := d.Management.Heartbeat(c.Request.Context(), nodeID, input.Hostname, input.AgentVersion, d.Now().UTC()); err != nil {
		writeError(c, http.StatusServiceUnavailable, "storage_unavailable", "could not store heartbeat")
		return
	}
	c.Status(http.StatusAccepted)
}

type grafanaWebhook struct {
	Status string `json:"status"`
	Alerts []struct {
		Status      string            `json:"status"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		StartsAt    time.Time         `json:"startsAt"`
		EndsAt      time.Time         `json:"endsAt"`
		Fingerprint string            `json:"fingerprint"`
	} `json:"alerts"`
}

func (d Dependencies) grafanaWebhook(c *gin.Context) {
	provided := c.GetHeader("X-VPSentinel-Webhook-Secret")
	if bearer, ok := bearerToken(c.GetHeader("Authorization")); ok {
		provided = bearer
	}
	if len(provided) != len(d.WebhookSecret) || subtle.ConstantTimeCompare([]byte(provided), []byte(d.WebhookSecret)) != 1 {
		writeError(c, http.StatusUnauthorized, "unauthorized", "valid webhook secret required")
		return
	}
	var payload grafanaWebhook
	// Grafana adds fields across versions, so this integration deliberately
	// accepts unknown fields while still enforcing the global body-size limit.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, d.MaxBodyBytes)
	if err := json.NewDecoder(c.Request.Body).Decode(&payload); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_json", "request body must contain a valid Grafana webhook")
		return
	}
	raw, _ := json.Marshal(payload)
	accepted := 0
	for _, alert := range payload.Alerts {
		state := strings.ToLower(alert.Status)
		if state == "" {
			state = strings.ToLower(payload.Status)
		}
		if state != "firing" && state != "resolved" {
			continue
		}
		if alert.StartsAt.IsZero() {
			alert.StartsAt = d.Now().UTC()
		}
		fingerprint := alert.Fingerprint
		if fingerprint == "" {
			sum := sha256.Sum256([]byte(fmt.Sprint(alert.Labels) + alert.StartsAt.String()))
			fingerprint = hex.EncodeToString(sum[:])
		}
		ruleName := firstNonEmpty(alert.Labels["alertname"], alert.Labels["rulename"], "Grafana alert")
		severity := firstNonEmpty(alert.Labels["severity"], "warning")
		summary := firstNonEmpty(alert.Annotations["summary"], alert.Annotations["description"], ruleName)
		var endsAt *time.Time
		if !alert.EndsAt.IsZero() {
			endsAt = &alert.EndsAt
		}
		inserted, err := d.Management.RecordAlert(c.Request.Context(), management.AlertInput{Fingerprint: fingerprint, NodeID: alert.Labels["node_id"], RuleName: ruleName, Severity: severity, State: state, Summary: summary, StartsAt: alert.StartsAt, EndsAt: endsAt, RawPayload: raw}, d.Now().UTC())
		if err != nil {
			d.Logger.Error("record Grafana alert", "error", err, "fingerprint", fingerprint)
			writeError(c, http.StatusServiceUnavailable, "storage_unavailable", "could not record alerts")
			return
		}
		if inserted {
			accepted++
		}
	}
	c.JSON(http.StatusAccepted, gin.H{"accepted_events": accepted})
}

func handleManagementError(c *gin.Context, err error, resource string) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, management.ErrNotFound) {
		writeError(c, http.StatusNotFound, "not_found", resource+" was not found")
	} else {
		writeError(c, http.StatusServiceUnavailable, "storage_unavailable", "could not load "+resource)
	}
	return false
}

func (d Dependencies) grafanaNodeURL(nodeID string) string {
	return strings.TrimRight(d.GrafanaURL, "/") + "/d/vpsentinel-node/node-detail?var-node_id=" + url.QueryEscape(nodeID)
}

func randomID(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func randomSecret(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
