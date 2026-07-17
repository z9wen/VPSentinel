package ingest

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/z9wen/VPSentinel/protocol"
)

const (
	maxBatchSamples   = 120
	maxClockSkew      = 5 * time.Minute
	maxSampleAge      = 24 * time.Hour
	maxPostgresBigInt = uint64(1<<63 - 1)
)

func ValidateBatch(batch protocol.MetricBatch, now time.Time) error {
	if batch.SchemaVersion != protocol.SchemaVersion {
		return fmt.Errorf("unsupported schema version %q", batch.SchemaVersion)
	}
	if !validID(batch.BatchID) {
		return errors.New("batch_id must be a 128-bit hexadecimal ID")
	}
	if batch.Hostname == "" || len(batch.Hostname) > 255 {
		return errors.New("hostname must contain between 1 and 255 characters")
	}
	if batch.AgentVersion == "" || len(batch.AgentVersion) > 64 {
		return errors.New("agent_version must contain between 1 and 64 characters")
	}
	if len(batch.Samples) == 0 || len(batch.Samples) > maxBatchSamples {
		return fmt.Errorf("samples must contain between 1 and %d entries", maxBatchSamples)
	}

	seenSampleIDs := make(map[string]struct{}, len(batch.Samples))
	for index, sample := range batch.Samples {
		if err := validateSample(sample, now); err != nil {
			return fmt.Errorf("samples[%d]: %w", index, err)
		}
		if _, exists := seenSampleIDs[sample.SampleID]; exists {
			return fmt.Errorf("samples[%d]: duplicate sample_id", index)
		}
		seenSampleIDs[sample.SampleID] = struct{}{}
	}
	return nil
}

func validateSample(sample protocol.MetricSample, now time.Time) error {
	if !validID(sample.SampleID) {
		return errors.New("sample_id must be a 128-bit hexadecimal ID")
	}
	if sample.RecordedAt.IsZero() {
		return errors.New("recorded_at is required")
	}
	if sample.RecordedAt.After(now.Add(maxClockSkew)) {
		return errors.New("recorded_at is too far in the future")
	}
	if sample.RecordedAt.Before(now.Add(-maxSampleAge)) {
		return errors.New("recorded_at is too old")
	}
	if sample.CPUPercent != nil && (!finite(*sample.CPUPercent) || *sample.CPUPercent < 0 || *sample.CPUPercent > 100) {
		return errors.New("cpu_percent must be between 0 and 100")
	}
	if err := validateNonNegative("load_1", sample.Load1); err != nil {
		return err
	}
	if err := validateNonNegative("load_5", sample.Load5); err != nil {
		return err
	}
	if err := validateNonNegative("load_15", sample.Load15); err != nil {
		return err
	}
	if sample.MemoryUsedBytes != nil && sample.MemoryTotalBytes != nil && *sample.MemoryUsedBytes > *sample.MemoryTotalBytes {
		return errors.New("memory_used_bytes must not exceed memory_total_bytes")
	}
	if sample.DiskUsedBytes != nil && sample.DiskTotalBytes != nil && *sample.DiskUsedBytes > *sample.DiskTotalBytes {
		return errors.New("disk_used_bytes must not exceed disk_total_bytes")
	}
	for name, value := range map[string]*uint64{
		"memory_used_bytes":  sample.MemoryUsedBytes,
		"memory_total_bytes": sample.MemoryTotalBytes,
		"disk_used_bytes":    sample.DiskUsedBytes,
		"disk_total_bytes":   sample.DiskTotalBytes,
		"uptime_seconds":     sample.UptimeSeconds,
		"network_rx_bytes":   sample.NetworkRXBytes,
		"network_tx_bytes":   sample.NetworkTXBytes,
	} {
		if value != nil && *value > maxPostgresBigInt {
			return fmt.Errorf("%s exceeds PostgreSQL BIGINT", name)
		}
	}
	if len(sample.OSName) > 128 || len(sample.KernelVersion) > 255 {
		return errors.New("host metadata is too long")
	}
	if len(sample.FailedCollectors) > 8 {
		return errors.New("too many failed_collectors entries")
	}
	for _, collector := range sample.FailedCollectors {
		if collector == "" || len(collector) > 32 {
			return errors.New("failed_collectors entries must contain between 1 and 32 characters")
		}
	}
	return nil
}

func validateNonNegative(name string, value *float64) error {
	if value != nil && (!finite(*value) || *value < 0) {
		return fmt.Errorf("%s must be a finite non-negative number", name)
	}
	return nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func validID(value string) bool {
	if len(value) != 32 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 16
}
