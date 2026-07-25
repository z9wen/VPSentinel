package database

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/z9wen/VPSentinel/protocol"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Ready(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *PostgresStore) StoreBatch(
	ctx context.Context,
	nodeID string,
	batch protocol.MetricBatch,
	receivedAt time.Time,
) (int, error) {
	if len(batch.Samples) == 0 {
		return 0, fmt.Errorf("metric batch must contain at least one sample")
	}
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin metric batch transaction: %w", err)
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	latestSample := batch.Samples[len(batch.Samples)-1]
	if _, err := transaction.Exec(ctx, `
		INSERT INTO nodes (
			id, name, hostname, agent_version, os_name, kernel_version, last_seen_at
		) VALUES ($1, $2, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6)
		ON CONFLICT (id) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			agent_version = EXCLUDED.agent_version,
			os_name = COALESCE(EXCLUDED.os_name, nodes.os_name),
			kernel_version = COALESCE(EXCLUDED.kernel_version, nodes.kernel_version),
			last_seen_at = GREATEST(nodes.last_seen_at, EXCLUDED.last_seen_at),
			updated_at = NOW()
	`,
		nodeID,
		batch.Hostname,
		batch.AgentVersion,
		latestSample.OSName,
		latestSample.KernelVersion,
		receivedAt,
	); err != nil {
		return 0, fmt.Errorf("upsert node metadata: %w", err)
	}

	acceptedSamples := 0
	for _, sample := range batch.Samples {
		memoryUsed, err := postgresBigInt(sample.MemoryUsedBytes)
		if err != nil {
			return 0, fmt.Errorf("sample %s memory_used_bytes: %w", sample.SampleID, err)
		}
		memoryTotal, err := postgresBigInt(sample.MemoryTotalBytes)
		if err != nil {
			return 0, fmt.Errorf("sample %s memory_total_bytes: %w", sample.SampleID, err)
		}
		diskUsed, err := postgresBigInt(sample.DiskUsedBytes)
		if err != nil {
			return 0, fmt.Errorf("sample %s disk_used_bytes: %w", sample.SampleID, err)
		}
		diskTotal, err := postgresBigInt(sample.DiskTotalBytes)
		if err != nil {
			return 0, fmt.Errorf("sample %s disk_total_bytes: %w", sample.SampleID, err)
		}
		uptime, err := postgresBigInt(sample.UptimeSeconds)
		if err != nil {
			return 0, fmt.Errorf("sample %s uptime_seconds: %w", sample.SampleID, err)
		}
		networkRX, err := postgresBigInt(sample.NetworkRXBytes)
		if err != nil {
			return 0, fmt.Errorf("sample %s network_rx_bytes: %w", sample.SampleID, err)
		}
		networkTX, err := postgresBigInt(sample.NetworkTXBytes)
		if err != nil {
			return 0, fmt.Errorf("sample %s network_tx_bytes: %w", sample.SampleID, err)
		}
		failedCollectors := sample.FailedCollectors
		if failedCollectors == nil {
			failedCollectors = []string{}
		}

		commandTag, err := transaction.Exec(ctx, `
			INSERT INTO metric_samples (
				node_id, sample_id, recorded_at, received_at,
				cpu_percent, memory_used_bytes, memory_total_bytes,
				disk_used_bytes, disk_total_bytes,
				load_1, load_5, load_15, uptime_seconds,
				network_rx_bytes, network_tx_bytes, failed_collectors
			) VALUES (
				$1, $2, $3, $4,
				$5, $6, $7,
				$8, $9,
				$10, $11, $12, $13,
				$14, $15, $16
			)
			ON CONFLICT (node_id, sample_id) DO NOTHING
		`,
			nodeID,
			sample.SampleID,
			sample.RecordedAt,
			receivedAt,
			nullable(sample.CPUPercent),
			memoryUsed,
			memoryTotal,
			diskUsed,
			diskTotal,
			nullable(sample.Load1),
			nullable(sample.Load5),
			nullable(sample.Load15),
			uptime,
			networkRX,
			networkTX,
			failedCollectors,
		)
		if err != nil {
			return 0, fmt.Errorf("insert metric sample %s: %w", sample.SampleID, err)
		}
		acceptedSamples += int(commandTag.RowsAffected())
	}

	if _, err := transaction.Exec(ctx, `
		INSERT INTO metric_batches (node_id, batch_id, received_at, accepted_samples)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (node_id, batch_id) DO NOTHING
	`, nodeID, batch.BatchID, receivedAt, acceptedSamples); err != nil {
		return 0, fmt.Errorf("record metric batch: %w", err)
	}

	if err := transaction.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit metric batch: %w", err)
	}
	return acceptedSamples, nil
}

func nullable[T any](value *T) any {
	if value == nil {
		return nil
	}
	return *value
}

func postgresBigInt(value *uint64) (any, error) {
	if value == nil {
		return nil, nil
	}
	if *value > math.MaxInt64 {
		return nil, fmt.Errorf("value %d exceeds PostgreSQL BIGINT", *value)
	}
	return int64(*value), nil
}
