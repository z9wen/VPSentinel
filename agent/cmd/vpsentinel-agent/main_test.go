package main

import (
	"context"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/z9wen/VPSentinel/agent/internal/queue"
	"github.com/z9wen/VPSentinel/protocol"
)

type countingCollector struct {
	count  atomic.Int32
	cancel context.CancelFunc
}

func (c *countingCollector) Collect(context.Context) (protocol.MetricSample, error) {
	count := c.count.Add(1)
	if count == 3 {
		c.cancel()
	}
	return protocol.MetricSample{RecordedAt: time.Now().UTC()}, nil
}

type blockingSender struct{}

func (blockingSender) Send(ctx context.Context, _ protocol.MetricBatch) error {
	<-ctx.Done()
	return ctx.Err()
}

func TestRunAgentContinuesCollectingWhileSendIsBlocked(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	collector := &countingCollector{cancel: cancel}
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	if err := runAgent(ctx, logger, collector, blockingSender{}, queue.New(10), "test-host", time.Millisecond); err != nil {
		t.Fatalf("runAgent: %v", err)
	}
	if count := collector.count.Load(); count < 3 {
		t.Fatalf("collector calls = %d, want at least 3", count)
	}
}

func TestNewBatchUsesVersionedProtocolAndStableSampleIDs(t *testing.T) {
	samples := []protocol.MetricSample{{SampleID: "sample-one"}}
	batch, err := newBatch("test-host", samples)
	if err != nil {
		t.Fatalf("newBatch: %v", err)
	}
	if batch.SchemaVersion != protocol.SchemaVersion {
		t.Fatalf("SchemaVersion = %q, want %q", batch.SchemaVersion, protocol.SchemaVersion)
	}
	if batch.BatchID == "" {
		t.Fatal("BatchID is empty")
	}
	if batch.Samples[0].SampleID != "sample-one" {
		t.Fatalf("SampleID = %q, want sample-one", batch.Samples[0].SampleID)
	}
}
