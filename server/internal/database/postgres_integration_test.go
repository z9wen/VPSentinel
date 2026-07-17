package database

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/z9wen/VPSentinel/protocol"
)

func TestPostgresMigrationsAndIdempotentStore(t *testing.T) {
	databaseURL := os.Getenv("VPSENTINEL_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("VPSENTINEL_TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	pool, err := Connect(ctx, databaseURL, logger)
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer pool.Close()

	if err := Migrate(ctx, pool, logger); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := Migrate(ctx, pool, logger); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}

	const nodeID = "integration-test-node"
	if _, err := pool.Exec(ctx, "DELETE FROM nodes WHERE id = $1", nodeID); err != nil {
		t.Fatalf("clean test node before test: %v", err)
	}
	defer func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = pool.Exec(cleanupCtx, "DELETE FROM nodes WHERE id = $1", nodeID)
	}()

	now := time.Now().UTC().Truncate(time.Microsecond)
	cpu := 12.5
	memoryUsed := uint64(1024)
	memoryTotal := uint64(2048)
	batch := protocol.MetricBatch{
		SchemaVersion: protocol.SchemaVersion,
		BatchID:       "11111111111111111111111111111111",
		Hostname:      "integration-host",
		AgentVersion:  "integration-test",
		Samples: []protocol.MetricSample{{
			SampleID:         "22222222222222222222222222222222",
			RecordedAt:       now,
			CPUPercent:       &cpu,
			MemoryUsedBytes:  &memoryUsed,
			MemoryTotalBytes: &memoryTotal,
			OSName:           "linux",
			KernelVersion:    "test-kernel",
		}},
	}

	store := NewPostgresStore(pool)
	accepted, err := store.StoreBatch(ctx, nodeID, batch, now)
	if err != nil {
		t.Fatalf("first StoreBatch: %v", err)
	}
	if accepted != 1 {
		t.Fatalf("first accepted samples = %d, want 1", accepted)
	}
	accepted, err = store.StoreBatch(ctx, nodeID, batch, now.Add(time.Second))
	if err != nil {
		t.Fatalf("second StoreBatch: %v", err)
	}
	if accepted != 0 {
		t.Fatalf("retry accepted samples = %d, want 0", accepted)
	}

	var sampleCount int
	if err := pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM metric_samples WHERE node_id = $1",
		nodeID,
	).Scan(&sampleCount); err != nil {
		t.Fatalf("count stored samples: %v", err)
	}
	if sampleCount != 1 {
		t.Fatalf("stored sample count = %d, want 1", sampleCount)
	}
}
