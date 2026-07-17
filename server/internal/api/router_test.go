package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/z9wen/VPSentinel/protocol"
	"github.com/z9wen/VPSentinel/server/internal/auth"
	"github.com/z9wen/VPSentinel/server/internal/ingest"
)

const testToken = "0123456789abcdef"

func TestIngestMetricsAcceptsAuthenticatedBatch(t *testing.T) {
	now := time.Date(2026, time.July, 17, 8, 0, 0, 0, time.UTC)
	store := ingest.NewMemoryStore(10)
	router := testRouter(store, now, 1<<20)
	body, err := json.Marshal(apiTestBatch(now))
	if err != nil {
		t.Fatalf("marshal batch: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/agent/metrics", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+testToken)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusAccepted, response.Body.String())
	}
	batches := store.Batches()
	if len(batches) != 1 || batches[0].NodeID != "node-one" {
		t.Fatalf("stored batches = %#v, want one batch for node-one", batches)
	}
}

func TestIngestMetricsTreatsRetriedSamplesAsIdempotent(t *testing.T) {
	now := time.Date(2026, time.July, 17, 8, 0, 0, 0, time.UTC)
	store := ingest.NewMemoryStore(10)
	router := testRouter(store, now, 1<<20)
	body, err := json.Marshal(apiTestBatch(now))
	if err != nil {
		t.Fatalf("marshal batch: %v", err)
	}

	for attempt := 0; attempt < 2; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/api/v1/agent/metrics", bytes.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("Authorization", "Bearer "+testToken)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusAccepted {
			t.Fatalf("attempt %d status = %d, want %d", attempt+1, response.Code, http.StatusAccepted)
		}
		if attempt == 1 && !strings.Contains(response.Body.String(), `"accepted_samples":0`) {
			t.Fatalf("retry body = %s, want zero accepted samples", response.Body.String())
		}
	}

	if batches := store.Batches(); len(batches) != 1 {
		t.Fatalf("stored batches = %d, want 1", len(batches))
	}
}

func TestIngestMetricsRejectsUnauthorizedRequest(t *testing.T) {
	now := time.Now().UTC()
	router := testRouter(ingest.NewMemoryStore(10), now, 1<<20)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/agent/metrics", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestIngestMetricsRejectsUnknownJSONField(t *testing.T) {
	now := time.Now().UTC()
	router := testRouter(ingest.NewMemoryStore(10), now, 1<<20)
	body := `{"schema_version":"v1","unexpected":true}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/agent/metrics", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+testToken)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusBadRequest, response.Body.String())
	}
}

func TestIngestMetricsEnforcesBodyLimit(t *testing.T) {
	now := time.Now().UTC()
	router := testRouter(ingest.NewMemoryStore(10), now, 32)
	body := `{"schema_version":"` + strings.Repeat("x", 64) + `"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/agent/metrics", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+testToken)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)
	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d; body=%s", response.Code, http.StatusRequestEntityTooLarge, response.Body.String())
	}
}

func TestHealthReadinessAndVersion(t *testing.T) {
	router := testRouter(ingest.NewMemoryStore(10), time.Now().UTC(), 1<<20)
	for path, wantStatus := range map[string]int{
		"/healthz": http.StatusOK,
		"/readyz":  http.StatusOK,
		"/version": http.StatusOK,
	} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != wantStatus {
			t.Errorf("GET %s status = %d, want %d", path, response.Code, wantStatus)
		}
	}
}

func testRouter(store ingest.Store, now time.Time, maxBodyBytes int64) http.Handler {
	gin.SetMode(gin.TestMode)
	return NewRouter(Dependencies{
		Authenticator: auth.NewStatic(testToken, "node-one"),
		Store:         store,
		Logger:        slog.New(slog.NewJSONHandler(io.Discard, nil)),
		Version:       "test-version",
		MaxBodyBytes:  maxBodyBytes,
		Now:           func() time.Time { return now },
	})
}

func apiTestBatch(recordedAt time.Time) protocol.MetricBatch {
	cpu := 10.5
	return protocol.MetricBatch{
		SchemaVersion: protocol.SchemaVersion,
		BatchID:       "11111111111111111111111111111111",
		Hostname:      "test-host",
		AgentVersion:  "test-version",
		Samples: []protocol.MetricSample{{
			SampleID:   "22222222222222222222222222222222",
			RecordedAt: recordedAt,
			CPUPercent: &cpu,
		}},
	}
}
