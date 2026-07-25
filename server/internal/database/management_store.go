package database

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/z9wen/VPSentinel/server/internal/auth"
	"github.com/z9wen/VPSentinel/server/internal/management"
)

func (s *PostgresStore) EnsureAdmin(ctx context.Context, email, password string) error {
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback(ctx) }()

	var userID int64
	var existingHash string
	err = transaction.QueryRow(ctx, "SELECT id, password_hash FROM users WHERE email = LOWER($1) FOR UPDATE", email).Scan(&userID, &existingHash)
	if err == nil && bcrypt.CompareHashAndPassword([]byte(existingHash), []byte(password)) == nil {
		return transaction.Commit(ctx)
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("load admin user: %w", err)
	}
	userMissing := errors.Is(err, pgx.ErrNoRows)
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	if userMissing {
		if _, err := transaction.Exec(ctx, "INSERT INTO users (email, password_hash) VALUES (LOWER($1), $2)", email, string(hash)); err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}
	} else {
		if _, err := transaction.Exec(ctx, "UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1", userID, string(hash)); err != nil {
			return fmt.Errorf("update admin user: %w", err)
		}
		// Changing the configured password invalidates all long-lived sessions.
		if _, err := transaction.Exec(ctx, "UPDATE refresh_sessions SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL", userID); err != nil {
			return fmt.Errorf("revoke admin sessions: %w", err)
		}
	}
	return transaction.Commit(ctx)
}

func (s *PostgresStore) EnsureBootstrapToken(ctx context.Context, nodeID, token string) error {
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	_, err = transaction.Exec(ctx, `
		INSERT INTO nodes (id, name, hostname, environment, agent_version, last_seen_at)
		VALUES ($1, $1, $1, 'production', 'pending', NULL)
		ON CONFLICT (id) DO NOTHING
	`, nodeID)
	if err != nil {
		return fmt.Errorf("ensure bootstrap node: %w", err)
	}
	_, err = transaction.Exec(ctx, `
		INSERT INTO agent_tokens (node_id, label, token_hash)
		VALUES ($1, 'bootstrap', $2)
		ON CONFLICT (token_hash) DO NOTHING
	`, nodeID, auth.HashToken(token))
	if err != nil {
		return fmt.Errorf("ensure bootstrap token: %w", err)
	}
	return transaction.Commit(ctx)
}

func (s *PostgresStore) PasswordHash(ctx context.Context, email string) (string, error) {
	var hash string
	err := s.pool.QueryRow(ctx, "SELECT password_hash FROM users WHERE email = LOWER($1)", email).Scan(&hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", management.ErrNotFound
	}
	return hash, err
}

func (s *PostgresStore) CreateRefreshSession(ctx context.Context, email string, tokenHash []byte, expiresAt time.Time) error {
	command, err := s.pool.Exec(ctx, `
		INSERT INTO refresh_sessions (user_id, token_hash, expires_at)
		SELECT id, $2, $3 FROM users WHERE email = LOWER($1)
	`, email, tokenHash, expiresAt)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return management.ErrNotFound
	}
	return nil
}

func (s *PostgresStore) RotateRefreshSession(ctx context.Context, oldHash, newHash []byte, expiresAt time.Time) (string, error) {
	transaction, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = transaction.Rollback(ctx) }()
	var userID int64
	var email string
	err = transaction.QueryRow(ctx, `
		UPDATE refresh_sessions rs SET revoked_at = NOW()
		FROM users u
		WHERE rs.user_id = u.id AND rs.token_hash = $1 AND rs.revoked_at IS NULL AND rs.expires_at > NOW()
		RETURNING rs.user_id, u.email
	`, oldHash).Scan(&userID, &email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", management.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if _, err := transaction.Exec(ctx, `
		INSERT INTO refresh_sessions (user_id, token_hash, expires_at) VALUES ($1, $2, $3)
	`, userID, newHash, expiresAt); err != nil {
		return "", err
	}
	if err := transaction.Commit(ctx); err != nil {
		return "", err
	}
	return email, nil
}

func (s *PostgresStore) RevokeRefreshSession(ctx context.Context, tokenHash []byte) error {
	_, err := s.pool.Exec(ctx, "UPDATE refresh_sessions SET revoked_at = NOW() WHERE token_hash = $1 AND revoked_at IS NULL", tokenHash)
	return err
}

func scanNode(row pgx.Row, now time.Time) (management.Node, error) {
	var node management.Node
	err := row.Scan(&node.ID, &node.Name, &node.Hostname, &node.Environment, &node.DisplayAddress,
		&node.AgentVersion, &node.OSName, &node.KernelVersion, &node.LastSeenAt, &node.CreatedAt)
	node.Status = management.Status(node.LastSeenAt, now)
	return node, err
}

const nodeColumns = `id, name, hostname, environment, display_address, agent_version, os_name, kernel_version, last_seen_at, created_at`

func (s *PostgresStore) ListNodes(ctx context.Context, now time.Time) ([]management.Node, error) {
	rows, err := s.pool.Query(ctx, "SELECT "+nodeColumns+" FROM nodes ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	nodes := make([]management.Node, 0)
	for rows.Next() {
		node, err := scanNode(rows, now)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, rows.Err()
}

func (s *PostgresStore) GetNode(ctx context.Context, id string, now time.Time) (management.Node, error) {
	node, err := scanNode(s.pool.QueryRow(ctx, "SELECT "+nodeColumns+" FROM nodes WHERE id = $1", id), now)
	if errors.Is(err, pgx.ErrNoRows) {
		return management.Node{}, management.ErrNotFound
	}
	return node, err
}

func (s *PostgresStore) CreateNode(ctx context.Context, node management.Node) (management.Node, error) {
	if node.Hostname == "" {
		node.Hostname = node.Name
	}
	if node.Environment == "" {
		node.Environment = "production"
	}
	created, err := scanNode(s.pool.QueryRow(ctx, `
		INSERT INTO nodes (id, name, hostname, environment, display_address, agent_version, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, 'pending', NULL)
		RETURNING `+nodeColumns,
		node.ID, node.Name, node.Hostname, node.Environment, node.DisplayAddress), time.Now().UTC())
	return created, err
}

func (s *PostgresStore) DeleteNode(ctx context.Context, id string) error {
	command, err := s.pool.Exec(ctx, "DELETE FROM nodes WHERE id = $1", id)
	if err == nil && command.RowsAffected() == 0 {
		return management.ErrNotFound
	}
	return err
}

func (s *PostgresStore) CreateToken(ctx context.Context, nodeID, label string, tokenHash []byte) (management.Token, error) {
	var token management.Token
	err := s.pool.QueryRow(ctx, `
		INSERT INTO agent_tokens (node_id, label, token_hash) VALUES ($1, $2, $3)
		RETURNING id, node_id, label, created_at, last_used_at, revoked_at
	`, nodeID, label, tokenHash).Scan(&token.ID, &token.NodeID, &token.Label, &token.CreatedAt, &token.LastUsedAt, &token.RevokedAt)
	if err != nil {
		return management.Token{}, err
	}
	return token, nil
}

func (s *PostgresStore) RevokeToken(ctx context.Context, nodeID string, tokenID int64) error {
	command, err := s.pool.Exec(ctx, `
		UPDATE agent_tokens SET revoked_at = NOW()
		WHERE id = $1 AND node_id = $2 AND revoked_at IS NULL
	`, tokenID, nodeID)
	if err == nil && command.RowsAffected() == 0 {
		return management.ErrNotFound
	}
	return err
}

func (s *PostgresStore) ListTokens(ctx context.Context, nodeID string) ([]management.Token, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, node_id, label, created_at, last_used_at, revoked_at
		FROM agent_tokens WHERE node_id = $1 ORDER BY created_at DESC
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tokens := make([]management.Token, 0)
	for rows.Next() {
		var token management.Token
		if err := rows.Scan(&token.ID, &token.NodeID, &token.Label, &token.CreatedAt, &token.LastUsedAt, &token.RevokedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (s *PostgresStore) LatestMetric(ctx context.Context, nodeID string) (management.Metric, error) {
	var metric management.Metric
	err := s.pool.QueryRow(ctx, `
		SELECT node_id, recorded_at, received_at, cpu_percent,
			memory_used_bytes, memory_total_bytes, disk_used_bytes, disk_total_bytes,
			load_1, load_5, load_15, uptime_seconds, network_rx_bytes, network_tx_bytes, failed_collectors
		FROM metric_samples WHERE node_id = $1 ORDER BY recorded_at DESC LIMIT 1
	`, nodeID).Scan(&metric.NodeID, &metric.RecordedAt, &metric.ReceivedAt, &metric.CPUPercent,
		&metric.MemoryUsedBytes, &metric.MemoryTotalBytes, &metric.DiskUsedBytes, &metric.DiskTotalBytes,
		&metric.Load1, &metric.Load5, &metric.Load15, &metric.UptimeSeconds, &metric.NetworkRXBytes,
		&metric.NetworkTXBytes, &metric.FailedCollectors)
	if errors.Is(err, pgx.ErrNoRows) {
		return management.Metric{}, management.ErrNotFound
	}
	return metric, err
}

const alertColumns = `id, fingerprint, node_id, rule_name, severity, state, summary, starts_at, ends_at, received_at, raw_payload`

func (s *PostgresStore) ListAlerts(ctx context.Context, active bool, limit int) ([]management.Alert, error) {
	query := "SELECT " + alertColumns + " FROM alert_events ORDER BY received_at DESC LIMIT $1"
	if active {
		query = "SELECT " + alertColumns + " FROM (SELECT DISTINCT ON (fingerprint) * FROM alert_events ORDER BY fingerprint, received_at DESC) latest WHERE state = 'firing' ORDER BY received_at DESC LIMIT $1"
	}
	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	alerts := make([]management.Alert, 0)
	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	return alerts, rows.Err()
}

func scanAlert(row pgx.Row) (management.Alert, error) {
	var alert management.Alert
	var raw []byte
	err := row.Scan(&alert.ID, &alert.Fingerprint, &alert.NodeID, &alert.RuleName, &alert.Severity,
		&alert.State, &alert.Summary, &alert.StartsAt, &alert.EndsAt, &alert.ReceivedAt, &raw)
	alert.RawPayload = json.RawMessage(raw)
	return alert, err
}

func (s *PostgresStore) GetAlert(ctx context.Context, id int64) (management.Alert, error) {
	alert, err := scanAlert(s.pool.QueryRow(ctx, "SELECT "+alertColumns+" FROM alert_events WHERE id = $1", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return management.Alert{}, management.ErrNotFound
	}
	return alert, err
}

func (s *PostgresStore) RecordAlert(ctx context.Context, input management.AlertInput, receivedAt time.Time) (bool, error) {
	command, err := s.pool.Exec(ctx, `
		INSERT INTO alert_events (fingerprint, node_id, rule_name, severity, state, summary, starts_at, ends_at, received_at, raw_payload)
		VALUES ($1, (SELECT id FROM nodes WHERE id = NULLIF($2, '')), $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (fingerprint, state, starts_at) DO NOTHING
	`, input.Fingerprint, input.NodeID, input.RuleName, input.Severity, input.State, input.Summary,
		input.StartsAt, input.EndsAt, receivedAt, input.RawPayload)
	return command.RowsAffected() > 0, err
}

func (s *PostgresStore) Heartbeat(ctx context.Context, nodeID, hostname, version string, receivedAt time.Time) error {
	command, err := s.pool.Exec(ctx, `
		UPDATE nodes SET hostname = COALESCE(NULLIF($2, ''), hostname), agent_version = COALESCE(NULLIF($3, ''), agent_version),
			last_seen_at = GREATEST(last_seen_at, $4), updated_at = NOW()
		WHERE id = $1
	`, nodeID, hostname, version, receivedAt)
	if err == nil && command.RowsAffected() == 0 {
		return management.ErrNotFound
	}
	return err
}

func (s *PostgresStore) CleanupMetrics(ctx context.Context, before time.Time) (int64, error) {
	command, err := s.pool.Exec(ctx, "DELETE FROM metric_samples WHERE recorded_at < $1", before)
	if err != nil {
		return 0, err
	}
	_, _ = s.pool.Exec(ctx, "DELETE FROM metric_batches WHERE received_at < $1", before)
	_, _ = s.pool.Exec(ctx, "DELETE FROM refresh_sessions WHERE expires_at < NOW() OR revoked_at < NOW() - INTERVAL '7 days'")
	return command.RowsAffected(), nil
}
