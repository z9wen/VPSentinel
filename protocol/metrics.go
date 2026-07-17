// Package protocol defines the versioned wire format shared by VPSentinel
// agents and the ingestion server.
package protocol

import "time"

const SchemaVersion = "v1"

// MetricSample is a point-in-time host snapshot. Pointer fields distinguish a
// genuine zero value from a collector that failed to produce a value.
type MetricSample struct {
	SampleID         string    `json:"sample_id"`
	RecordedAt       time.Time `json:"recorded_at"`
	CPUPercent       *float64  `json:"cpu_percent,omitempty"`
	MemoryUsedBytes  *uint64   `json:"memory_used_bytes,omitempty"`
	MemoryTotalBytes *uint64   `json:"memory_total_bytes,omitempty"`
	DiskUsedBytes    *uint64   `json:"disk_used_bytes,omitempty"`
	DiskTotalBytes   *uint64   `json:"disk_total_bytes,omitempty"`
	Load1            *float64  `json:"load_1,omitempty"`
	Load5            *float64  `json:"load_5,omitempty"`
	Load15           *float64  `json:"load_15,omitempty"`
	UptimeSeconds    *uint64   `json:"uptime_seconds,omitempty"`
	NetworkRXBytes   *uint64   `json:"network_rx_bytes,omitempty"`
	NetworkTXBytes   *uint64   `json:"network_tx_bytes,omitempty"`
	OSName           string    `json:"os_name,omitempty"`
	KernelVersion    string    `json:"kernel_version,omitempty"`
	FailedCollectors []string  `json:"failed_collectors,omitempty"`
}

// MetricBatch groups samples into one authenticated ingestion request.
type MetricBatch struct {
	SchemaVersion string         `json:"schema_version"`
	BatchID       string         `json:"batch_id"`
	Hostname      string         `json:"hostname"`
	AgentVersion  string         `json:"agent_version"`
	Samples       []MetricSample `json:"samples"`
}
