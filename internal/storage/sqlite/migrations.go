package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Migration represents a single versioned database migration step.
type Migration struct {
	Version int
	Name    string
	Up      func(ctx context.Context, tx *sql.Tx) error
}

var migrations = []Migration{}

// RegisterMigration adds a migration to the migration sequence.
func RegisterMigration(m Migration) {
	migrations = append(migrations, m)
}

// CurrentVersion queries the current schema version from schema_version table.
func (db *DB) CurrentVersion(ctx context.Context) (int, error) {
	// Create schema_version table if it doesn't exist
	_, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY,
			applied_at INTEGER NOT NULL
		);
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to ensure schema_version table: %w", err)
	}

	var version sql.NullInt64
	err = db.QueryRowContext(ctx, "SELECT MAX(version) FROM schema_version;").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to query schema version: %w", err)
	}

	if !version.Valid {
		return 0, nil
	}

	return int(version.Int64), nil
}

// Migrate executes all pending migrations within individual transactions.
func (db *DB) Migrate(ctx context.Context) error {
	current, err := db.CurrentVersion(ctx)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if m.Version <= current {
			continue
		}

		if err := db.runMigration(ctx, m); err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", m.Version, m.Name, err)
		}
	}

	return nil
}

func (db *DB) runMigration(ctx context.Context, m Migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", err)
	}
	defer tx.Rollback()

	if err := m.Up(ctx, tx); err != nil {
		return err
	}

	now := time.Now().Unix()
	_, err = tx.ExecContext(ctx, "INSERT INTO schema_version (version, applied_at) VALUES (?, ?);", m.Version, now)
	if err != nil {
		return fmt.Errorf("failed to record migration version: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	return nil
}
