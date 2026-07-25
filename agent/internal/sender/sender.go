package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/z9wen/VPSentinel/protocol"
)

const maxErrorBodyBytes = 4 << 10

type Client struct {
	serverURL   string
	token       string
	httpClient  *http.Client
	maxAttempts int
	baseDelay   time.Duration
}

func New(serverURL, token string, requestTimeout time.Duration) *Client {
	return &Client{
		serverURL:   serverURL,
		token:       token,
		httpClient:  &http.Client{Timeout: requestTimeout},
		maxAttempts: 4,
		baseDelay:   500 * time.Millisecond,
	}
}

// Send retries network failures, rate limits, and server errors. Authentication
// and validation errors are returned immediately because retries cannot fix them.
func (c *Client) Send(ctx context.Context, batch protocol.MetricBatch) error {
	body, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("encode metric batch: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < c.maxAttempts; attempt++ {
		if attempt > 0 {
			delay := c.backoff(attempt - 1)
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}

		retry, err := c.sendOnce(ctx, body, batch.AgentVersion)
		if err == nil {
			return nil
		}
		lastErr = err
		if !retry {
			return err
		}
	}
	return fmt.Errorf("send metric batch after %d attempts: %w", c.maxAttempts, lastErr)
}

func (c *Client) sendOnce(ctx context.Context, body []byte, agentVersion string) (bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.serverURL, bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("create metric request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", "vpsentinel-agent/"+agentVersion)

	response, err := c.httpClient.Do(request)
	if err != nil {
		// Stop only when the caller cancelled. A client timeout is treated as a
		// temporary network failure and remains eligible for a bounded retry.
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		return true, fmt.Errorf("perform metric request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		_, _ = io.Copy(io.Discard, response.Body)
		return false, nil
	}

	errorBody, _ := io.ReadAll(io.LimitReader(response.Body, maxErrorBodyBytes))
	err = fmt.Errorf("server returned %s: %s", response.Status, string(errorBody))
	retry := response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500
	return retry, err
}

func (c *Client) backoff(attempt int) time.Duration {
	// Full jitter spreads retries across [0, base*2^attempt) so a recovering
	// control plane is not hit by every agent at the same instant.
	maximum := c.baseDelay * time.Duration(1<<attempt)
	return time.Duration(rand.Int64N(int64(maximum)))
}
