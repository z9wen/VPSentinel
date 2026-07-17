package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/z9wen/VPSentinel/agent/internal/collector"
	"github.com/z9wen/VPSentinel/agent/internal/config"
	"github.com/z9wen/VPSentinel/agent/internal/queue"
	"github.com/z9wen/VPSentinel/agent/internal/sender"
	"github.com/z9wen/VPSentinel/agent/internal/version"
	"github.com/z9wen/VPSentinel/protocol"
)

const maxBatchSamples = 120

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := run(os.Args[1:], logger); err != nil {
		logger.Error("agent stopped", "error", err)
		os.Exit(1)
	}
}

func run(args []string, logger *slog.Logger) error {
	cfg, err := config.Load(args)
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("read hostname: %w", err)
	}

	hostCollector := collector.NewHostCollector(cfg.DiskPath)
	if cfg.DryRun {
		return printDryRun(ctx, os.Stdout, hostCollector, hostname)
	}

	client := sender.New(cfg.ServerURL, cfg.Token, cfg.RequestTimeout)
	sampleQueue := queue.New(cfg.QueueCapacity)

	logger.Info("agent started",
		"version", version.Version,
		"commit", version.Commit,
		"hostname", hostname,
		"interval", cfg.Interval,
	)

	return runAgent(ctx, logger, hostCollector, client, sampleQueue, hostname, cfg.Interval)
}

type metricCollector interface {
	// Collect may return a partial sample and a non-nil error. Callers should
	// still send the sample so a failed collector does not look like a dead node.
	Collect(context.Context) (protocol.MetricSample, error)
}

type metricSender interface {
	Send(context.Context, protocol.MetricBatch) error
}

type sendResult struct {
	batchID   string
	sampleIDs []string
	err       error
}

func runAgent(
	ctx context.Context,
	logger *slog.Logger,
	c metricCollector,
	s metricSender,
	q *queue.SampleQueue,
	hostname string,
	interval time.Duration,
) error {
	results := make(chan sendResult, 1)
	sending := false

	collect := func() error {
		sample, collectionErr := c.Collect(ctx)
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if sample.RecordedAt.IsZero() {
			sample.RecordedAt = time.Now().UTC()
		}
		sampleID, err := randomID()
		if err != nil {
			return fmt.Errorf("create sample ID: %w", err)
		}
		sample.SampleID = sampleID

		if q.Push(sample) {
			logger.Warn("metric queue full; dropped oldest sample")
		}
		if collectionErr != nil {
			logger.Warn("host metrics partially collected",
				"error", collectionErr,
				"failed_collectors", sample.FailedCollectors,
			)
		}
		return nil
	}

	startSend := func() error {
		if sending || q.Len() == 0 {
			return nil
		}
		batch, err := newBatch(hostname, q.Snapshot(maxBatchSamples))
		if err != nil {
			return err
		}
		sampleIDs := make([]string, len(batch.Samples))
		for index, sample := range batch.Samples {
			sampleIDs[index] = sample.SampleID
		}

		sending = true
		go func() {
			results <- sendResult{
				batchID:   batch.BatchID,
				sampleIDs: sampleIDs,
				err:       s.Send(ctx, batch),
			}
		}()
		return nil
	}

	// Collect immediately at startup instead of waiting for the first tick.
	if err := collect(); err != nil {
		if ctx.Err() != nil {
			return nil
		}
		return err
	}
	if err := startSend(); err != nil {
		return err
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Info("agent shutting down", "queued_samples", q.Len())
			return nil

		case <-ticker.C:
			if err := collect(); err != nil {
				if ctx.Err() != nil {
					return nil
				}
				logger.Warn("metric collection cycle failed", "error", err)
			}
			if err := startSend(); err != nil {
				return err
			}

		case result := <-results:
			sending = false
			if result.err != nil {
				logger.Warn("metric batch send failed",
					"error", result.err,
					"batch_id", result.batchID,
					"queued_samples", q.Len(),
				)
				// The sender already performed bounded retries. Wait for the next
				// collection tick before another outer attempt to avoid a hot loop.
				continue
			}

			removed := q.Remove(result.sampleIDs)
			logger.Info("metric batch sent",
				"batch_id", result.batchID,
				"sample_count", len(result.sampleIDs),
				"acknowledged_samples", removed,
			)
			if err := startSend(); err != nil {
				return err
			}
		}
	}
}

func printDryRun(ctx context.Context, output io.Writer, c metricCollector, hostname string) error {
	sample, collectionErr := c.Collect(ctx)
	if sample.RecordedAt.IsZero() {
		sample.RecordedAt = time.Now().UTC()
	}
	sampleID, err := randomID()
	if err != nil {
		return fmt.Errorf("create sample ID: %w", err)
	}
	sample.SampleID = sampleID

	batch, err := newBatch(hostname, []protocol.MetricSample{sample})
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(batch); err != nil {
		return fmt.Errorf("print dry-run batch: %w", err)
	}
	if collectionErr != nil {
		return fmt.Errorf("dry-run completed with collector failures: %w", collectionErr)
	}
	return nil
}

func newBatch(hostname string, samples []protocol.MetricSample) (protocol.MetricBatch, error) {
	batchID, err := randomID()
	if err != nil {
		return protocol.MetricBatch{}, fmt.Errorf("create batch ID: %w", err)
	}
	return protocol.MetricBatch{
		SchemaVersion: protocol.SchemaVersion,
		BatchID:       batchID,
		Hostname:      hostname,
		AgentVersion:  version.Version,
		Samples:       samples,
	}, nil
}

func randomID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
