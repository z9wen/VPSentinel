package database

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const migrationAdvisoryLockID int64 = 0x565053454E54494E // "VPSENTIN"

//go:embed migrations/*.sql
var migrationFiles embed.FS

type migration struct {
	version  string
	name     string
	SQL      string
	checksum string
}

func Migrate(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	connection, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer connection.Release()

	if _, err := connection.Exec(ctx, "SELECT pg_advisory_lock($1)", migrationAdvisoryLockID); err != nil {
		return fmt.Errorf("acquire migration advisory lock: %w", err)
	}
	defer func() {
		unlockCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = connection.Exec(unlockCtx, "SELECT pg_advisory_unlock($1)", migrationAdvisoryLockID)
	}()

	if _, err := connection.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			checksum TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	for _, candidate := range migrations {
		var appliedChecksum string
		err := connection.QueryRow(ctx,
			"SELECT checksum FROM schema_migrations WHERE version = $1",
			candidate.version,
		).Scan(&appliedChecksum)
		switch {
		case err == nil:
			if appliedChecksum != candidate.checksum {
				return fmt.Errorf("migration %s checksum changed after it was applied", candidate.version)
			}
			continue
		case !errors.Is(err, pgx.ErrNoRows):
			return fmt.Errorf("read migration %s state: %w", candidate.version, err)
		}

		transaction, err := connection.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", candidate.version, err)
		}
		if _, err := transaction.Exec(ctx, candidate.SQL); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("apply migration %s: %w", candidate.version, err)
		}
		if _, err := transaction.Exec(ctx,
			"INSERT INTO schema_migrations (version, name, checksum) VALUES ($1, $2, $3)",
			candidate.version,
			candidate.name,
			candidate.checksum,
		); err != nil {
			_ = transaction.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", candidate.version, err)
		}
		if err := transaction.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", candidate.version, err)
		}
		logger.Info("database migration applied", "version", candidate.version, "name", candidate.name)
	}
	return nil
}

func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationFiles, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read embedded migrations: %w", err)
	}
	result := make([]migration, 0, len(entries))
	seenVersions := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		parts := strings.SplitN(strings.TrimSuffix(entry.Name(), ".sql"), "_", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid migration filename %q", entry.Name())
		}
		if _, duplicate := seenVersions[parts[0]]; duplicate {
			return nil, fmt.Errorf("duplicate migration version %q", parts[0])
		}
		seenVersions[parts[0]] = struct{}{}

		contents, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}
		checksum := sha256.Sum256(contents)
		result = append(result, migration{
			version:  parts[0],
			name:     parts[1],
			SQL:      string(contents),
			checksum: hex.EncodeToString(checksum[:]),
		})
	}
	if len(result) == 0 {
		return nil, errors.New("no embedded database migrations found")
	}
	return result, nil
}
