package sender

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/z9wen/VPSentinel/protocol"
)

// roundTripFunc 让普通函数实现 http.RoundTripper 接口，因此测试无需真的监听端口。
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestSendAddsAuthenticationAndRetriesServerError(t *testing.T) {
	attempts := 0
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		attempts++
		if request.Header.Get("Authorization") != "Bearer secret" {
			t.Errorf("Authorization = %q", request.Header.Get("Authorization"))
		}
		if attempts == 1 {
			return response(http.StatusServiceUnavailable, "temporary failure"), nil
		}
		return response(http.StatusAccepted, ""), nil
	})

	client := New("https://monitor.example/metrics", "secret", time.Second)
	client.httpClient.Transport = transport
	client.baseDelay = time.Millisecond
	batch := protocol.MetricBatch{AgentVersion: "test", Samples: []protocol.MetricSample{{}}}
	if err := client.Send(context.Background(), batch); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("attempts = %d, want 2", attempts)
	}
}

func TestSendDoesNotRetryUnauthorized(t *testing.T) {
	attempts := 0
	transport := roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		attempts++
		return response(http.StatusUnauthorized, "unauthorized"), nil
	})

	client := New("https://monitor.example/metrics", "wrong", time.Second)
	client.httpClient.Transport = transport
	client.baseDelay = time.Millisecond
	err := client.Send(context.Background(), protocol.MetricBatch{AgentVersion: "test"})
	if err == nil {
		t.Fatal("Send returned nil error, want unauthorized error")
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func response(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
