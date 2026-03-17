package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func WaitForDB(ctx context.Context, db *sql.DB, attempts int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := db.PingContext(ctx); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("database is not ready after %d attempts: %w", attempts, lastErr)
}

func RunMigrations(ctx context.Context, db *sql.DB, dir string) error {
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("migrations directory %q is not available: %w", dir, err)
	}

	if _, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
	filename TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`); err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		name := filepath.Base(file)

		var applied bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE filename = $1)`, name).Scan(&applied); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if applied {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		upSQL, err := extractUpSQL(string(content))
		if err != nil {
			return fmt.Errorf("parse migration %s: %w", name, err)
		}
		if strings.TrimSpace(upSQL) == "" {
			continue
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, upSQL); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (filename) VALUES ($1)`, name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("mark migration %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	return nil
}

func extractUpSQL(content string) (string, error) {
	upIndex := strings.Index(content, "-- +goose Up")
	if upIndex == -1 {
		return "", fmt.Errorf("missing goose up marker")
	}
	downIndex := strings.Index(content, "-- +goose Down")
	if downIndex == -1 {
		return strings.TrimSpace(content[upIndex+len("-- +goose Up"):]), nil
	}
	return strings.TrimSpace(content[upIndex+len("-- +goose Up") : downIndex]), nil
}
