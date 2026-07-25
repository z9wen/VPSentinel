package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/z9wen/VPSentinel/server/internal/auth"
	"github.com/z9wen/VPSentinel/server/internal/ingest"
	"github.com/z9wen/VPSentinel/server/internal/management"
)

type fakeManagementStore struct {
	passwordHash string
	refreshHash  []byte
}

func (s *fakeManagementStore) PasswordHash(context.Context, string) (string, error) {
	return s.passwordHash, nil
}
func (s *fakeManagementStore) CreateRefreshSession(_ context.Context, _ string, hash []byte, _ time.Time) error {
	s.refreshHash = hash
	return nil
}
func (s *fakeManagementStore) RotateRefreshSession(_ context.Context, oldHash, newHash []byte, _ time.Time) (string, error) {
	if !bytes.Equal(oldHash, s.refreshHash) {
		return "", management.ErrNotFound
	}
	s.refreshHash = newHash
	return "admin@example.com", nil
}
func (s *fakeManagementStore) RevokeRefreshSession(context.Context, []byte) error { return nil }
func (s *fakeManagementStore) ListNodes(context.Context, time.Time) ([]management.Node, error) {
	return []management.Node{{ID: "node-one", Name: "Production web", Status: "online"}}, nil
}
func (s *fakeManagementStore) GetNode(context.Context, string, time.Time) (management.Node, error) {
	return management.Node{}, management.ErrNotFound
}
func (s *fakeManagementStore) CreateNode(context.Context, management.Node) (management.Node, error) {
	return management.Node{}, nil
}
func (s *fakeManagementStore) DeleteNode(context.Context, string) error { return nil }
func (s *fakeManagementStore) CreateToken(context.Context, string, string, []byte) (management.Token, error) {
	return management.Token{}, nil
}
func (s *fakeManagementStore) RevokeToken(context.Context, string, int64) error { return nil }
func (s *fakeManagementStore) ListTokens(context.Context, string) ([]management.Token, error) {
	return []management.Token{}, nil
}
func (s *fakeManagementStore) LatestMetric(context.Context, string) (management.Metric, error) {
	return management.Metric{}, management.ErrNotFound
}
func (s *fakeManagementStore) ListAlerts(context.Context, bool, int) ([]management.Alert, error) {
	return []management.Alert{}, nil
}
func (s *fakeManagementStore) GetAlert(context.Context, int64) (management.Alert, error) {
	return management.Alert{}, management.ErrNotFound
}
func (s *fakeManagementStore) RecordAlert(context.Context, management.AlertInput, time.Time) (bool, error) {
	return true, nil
}
func (s *fakeManagementStore) Heartbeat(context.Context, string, string, string, time.Time) error {
	return nil
}
func (s *fakeManagementStore) CleanupMetrics(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func TestAdminLoginAndAccessToken(t *testing.T) {
	now := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	managementStore := &fakeManagementStore{passwordHash: string(passwordHash)}
	router := adminTestRouter(managementStore, now)

	loginBody := []byte(`{"email":"admin@example.com","password":"correct-password"}`)
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginRequest.Header.Set("Content-Type", "application/json")
	loginResponse := httptest.NewRecorder()
	router.ServeHTTP(loginResponse, loginRequest)
	if loginResponse.Code != http.StatusOK {
		t.Fatalf("login status = %d; body=%s", loginResponse.Code, loginResponse.Body.String())
	}
	var login struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(loginResponse.Body.Bytes(), &login); err != nil || login.AccessToken == "" {
		t.Fatalf("decode login response: %v; body=%s", err, loginResponse.Body.String())
	}
	cookies := loginResponse.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != refreshCookieName || !cookies[0].HttpOnly {
		t.Fatalf("refresh cookie = %#v, want one HttpOnly cookie", cookies)
	}

	nodesRequest := httptest.NewRequest(http.MethodGet, "/api/v1/nodes", nil)
	nodesRequest.Header.Set("Authorization", "Bearer "+login.AccessToken)
	nodesResponse := httptest.NewRecorder()
	router.ServeHTTP(nodesResponse, nodesRequest)
	if nodesResponse.Code != http.StatusOK || !bytes.Contains(nodesResponse.Body.Bytes(), []byte("node-one")) {
		t.Fatalf("nodes status = %d; body=%s", nodesResponse.Code, nodesResponse.Body.String())
	}
}

func TestRefreshTokenIsRotated(t *testing.T) {
	now := time.Now().UTC()
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	store := &fakeManagementStore{passwordHash: string(hash)}
	router := adminTestRouter(store, now)

	loginRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"email":"admin@example.com","password":"correct-password"}`))
	loginResponse := httptest.NewRecorder()
	router.ServeHTTP(loginResponse, loginRequest)
	oldCookie := loginResponse.Result().Cookies()[0]

	refreshRequest := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	refreshRequest.AddCookie(oldCookie)
	refreshResponse := httptest.NewRecorder()
	router.ServeHTTP(refreshResponse, refreshRequest)
	if refreshResponse.Code != http.StatusOK {
		t.Fatalf("refresh status = %d; body=%s", refreshResponse.Code, refreshResponse.Body.String())
	}
	newCookie := refreshResponse.Result().Cookies()[0]
	if newCookie.Value == oldCookie.Value {
		t.Fatal("refresh endpoint did not rotate the refresh token")
	}
}

func adminTestRouter(store management.Store, now time.Time) http.Handler {
	gin.SetMode(gin.TestMode)
	return NewRouter(Dependencies{
		Authenticator: auth.NewStatic(testToken, "node-one"),
		Store:         ingest.NewMemoryStore(10),
		Management:    store,
		Sessions:      auth.NewSessionManager("0123456789abcdef0123456789abcdef", 15*time.Minute, 24*time.Hour),
		Logger:        slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Now:           func() time.Time { return now },
		WebhookSecret: "webhook-secret-123",
		GrafanaURL:    "/grafana",
	})
}
