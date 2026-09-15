package sqlite

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB with GoVault-specific SQLite configuration and path metadata.
type DB struct {
	*sql.DB
	path string
}

// Open initializes and configures a SQLite connection with required PRAGMAs and file permissions.
func Open(path string) (*DB, error) {
	if err := EnsurePermissions(path); err != nil {
		return nil, fmt.Errorf("failed to enforce permissions: %w", err)
	}

	dsn := path
	if path == "" {
		dsn = ":memory:"
	}

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// SQLite connection pool limits
	sqlDB.SetMaxOpenConns(1) // Single writer constraint prevents database locking conflicts

	db := &DB{
		DB:   sqlDB,
		path: path,
	}

	if err := db.applyPragmas(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to apply pragmas: %w", err)
	}

	// Re-check permissions on created WAL/SHM files
	if err := EnsurePermissions(path); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to re-apply permissions: %w", err)
	}

	return db, nil
}

func (db *DB) applyPragmas() error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
	}

	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("failed executing %q: %w", p, err)
		}
	}

	return nil
}

// Path returns the filesystem path of the database.
func (db *DB) Path() string {
	return db.path
}
