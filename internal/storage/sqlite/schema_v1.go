package sqlite

import (
	"context"
	"database/sql"
)

func init() {
	RegisterMigration(Migration{
		Version: 1,
		Name:    "baseline_schema_v1",
		Up:      migrationV1Up,
	})
}

func migrationV1Up(ctx context.Context, tx *sql.Tx) error {
	queries := []string{
		// 1. vault_metadata table
		`CREATE TABLE IF NOT EXISTS vault_metadata (
			vault_id TEXT PRIMARY KEY,
			schema_version INTEGER NOT NULL,
			crypto_suite TEXT NOT NULL,
			kdf_params_json TEXT NOT NULL,
			wrapped_key BLOB NOT NULL,
			mck BLOB NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		);`,

		// 2. records table
		`CREATE TABLE IF NOT EXISTS records (
			id TEXT PRIMARY KEY,
			vault_id TEXT NOT NULL,
			record_type TEXT NOT NULL,
			version INTEGER NOT NULL,
			payload BLOB NOT NULL,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL,
			deleted_at INTEGER
		);`,

		// 3. record_history table
		`CREATE TABLE IF NOT EXISTS record_history (
			id TEXT PRIMARY KEY,
			record_id TEXT NOT NULL,
			vault_id TEXT NOT NULL,
			version INTEGER NOT NULL,
			payload BLOB NOT NULL,
			archived_at INTEGER NOT NULL,
			FOREIGN KEY(record_id) REFERENCES records(id) ON DELETE CASCADE
		);`,

		// 4. tags table
		`CREATE TABLE IF NOT EXISTS tags (
			id TEXT PRIMARY KEY,
			vault_id TEXT NOT NULL,
			name TEXT NOT NULL,
			created_at INTEGER NOT NULL,
			UNIQUE(vault_id, name)
		);`,

		// 5. record_tags junction table
		`CREATE TABLE IF NOT EXISTS record_tags (
			record_id TEXT NOT NULL,
			tag_id TEXT NOT NULL,
			PRIMARY KEY(record_id, tag_id),
			FOREIGN KEY(record_id) REFERENCES records(id) ON DELETE CASCADE,
			FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE
		);`,

		// 6. Indices
		`CREATE INDEX IF NOT EXISTS idx_records_vault_deleted ON records(vault_id, deleted_at);`,
		`CREATE INDEX IF NOT EXISTS idx_record_history_record_id ON record_history(record_id);`,
		`CREATE INDEX IF NOT EXISTS idx_tags_vault_id ON tags(vault_id);`,
	}

	for _, q := range queries {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return err
		}
	}

	return nil
}
