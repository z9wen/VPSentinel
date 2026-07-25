package management

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var ErrNotFound = errors.New("resource not found")

type Node struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Hostname       string     `json:"hostname"`
	Environment    string     `json:"environment"`
	DisplayAddress *string    `json:"display_address"`
	AgentVersion   string     `json:"agent_version"`
	OSName         *string    `json:"os_name"`
	KernelVersion  *string    `json:"kernel_version"`
	LastSeenAt     *time.Time `json:"last_seen_at"`
	CreatedAt      time.Time  `json:"created_at"`
	Status         string     `json:"status"`
	GrafanaURL     string     `json:"grafana_url,omitempty"`
}

type Metric struct {
	NodeID           string    `json:"node_id"`
	RecordedAt       time.Time `json:"recorded_at"`
	ReceivedAt       time.Time `json:"received_at"`
	CPUPercent       *float64  `json:"cpu_percent"`
	MemoryUsedBytes  *int64    `json:"memory_used_bytes"`
	MemoryTotalBytes *int64    `json:"memory_total_bytes"`
	DiskUsedBytes    *int64    `json:"disk_used_bytes"`
	DiskTotalBytes   *int64    `json:"disk_total_bytes"`
	Load1            *float64  `json:"load_1"`
	Load5            *float64  `json:"load_5"`
	Load15           *float64  `json:"load_15"`
	UptimeSeconds    *int64    `json:"uptime_seconds"`
	NetworkRXBytes   *int64    `json:"network_rx_bytes"`
	NetworkTXBytes   *int64    `json:"network_tx_bytes"`
	FailedCollectors []string  `json:"failed_collectors"`
}

type Token struct {
	ID         int64      `json:"id"`
	NodeID     string     `json:"node_id"`
	Label      string     `json:"label"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
}

type Alert struct {
	ID          int64           `json:"id"`
	Fingerprint string          `json:"fingerprint"`
	NodeID      *string         `json:"node_id"`
	RuleName    string          `json:"rule_name"`
	Severity    string          `json:"severity"`
	State       string          `json:"state"`
	Summary     string          `json:"summary"`
	StartsAt    time.Time       `json:"starts_at"`
	EndsAt      *time.Time      `json:"ends_at"`
	ReceivedAt  time.Time       `json:"received_at"`
	RawPayload  json.RawMessage `json:"raw_payload,omitempty"`
}

type AlertInput struct {
	Fingerprint string
	NodeID      string
	RuleName    string
	Severity    string
	State       string
	Summary     string
	StartsAt    time.Time
	EndsAt      *time.Time
	RawPayload  json.RawMessage
}

type Store interface {
	PasswordHash(context.Context, string) (string, error)
	CreateRefreshSession(context.Context, string, []byte, time.Time) error
	RotateRefreshSession(context.Context, []byte, []byte, time.Time) (string, error)
	RevokeRefreshSession(context.Context, []byte) error
	ListNodes(context.Context, time.Time) ([]Node, error)
	GetNode(context.Context, string, time.Time) (Node, error)
	CreateNode(context.Context, Node) (Node, error)
	DeleteNode(context.Context, string) error
	CreateToken(context.Context, string, string, []byte) (Token, error)
	ListTokens(context.Context, string) ([]Token, error)
	RevokeToken(context.Context, string, int64) error
	LatestMetric(context.Context, string) (Metric, error)
	ListAlerts(context.Context, bool, int) ([]Alert, error)
	GetAlert(context.Context, int64) (Alert, error)
	RecordAlert(context.Context, AlertInput, time.Time) (bool, error)
	Heartbeat(context.Context, string, string, string, time.Time) error
	CleanupMetrics(context.Context, time.Time) (int64, error)
}

func Status(lastSeen *time.Time, now time.Time) string {
	if lastSeen == nil {
		return "pending"
	}
	age := now.Sub(*lastSeen)
	if age <= 2*time.Minute {
		return "online"
	}
	if age <= 5*time.Minute {
		return "degraded"
	}
	return "offline"
}
