package ingest

import (
	"strings"
	"testing"
	"time"

	"github.com/z9wen/VPSentinel/protocol"
)

func TestValidateBatchAcceptsPartialSample(t *testing.T) {
	now := time.Now().UTC()
	batch := validBatch(now)
	batch.Samples[0].FailedCollectors = []string{"disk"}
	if err := ValidateBatch(batch, now); err != nil {
		t.Fatalf("ValidateBatch: %v", err)
	}
}

func TestValidateBatchRejectsDuplicateSampleID(t *testing.T) {
	now := time.Now().UTC()
	batch := validBatch(now)
	batch.Samples = append(batch.Samples, batch.Samples[0])
	if err := ValidateBatch(batch, now); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("ValidateBatch error = %v, want duplicate sample error", err)
	}
}

func validBatch(now time.Time) protocol.MetricBatch {
	cpu := 12.5
	return protocol.MetricBatch{
		SchemaVersion: protocol.SchemaVersion,
		BatchID:       "11111111111111111111111111111111",
		Hostname:      "test-host",
		AgentVersion:  "test",
		Samples: []protocol.MetricSample{{
			SampleID:   "22222222222222222222222222222222",
			RecordedAt: now,
			CPUPercent: &cpu,
		}},
	}
}
