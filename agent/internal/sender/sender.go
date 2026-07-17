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

// Send 会重试网络错误、429 和 5xx。认证或请求格式错误不会因重试而恢复。
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
		// 只有调用者的 context 已经结束时才停止。http.Client 自己的超时
		// 属于临时网络故障，仍然值得重试。
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
	// Full jitter: 随机等待 [0, base*2^attempt)，避免大量 agent 同时重试。
	maximum := c.baseDelay * time.Duration(1<<attempt)
	return time.Duration(rand.Int64N(int64(maximum)))
}
