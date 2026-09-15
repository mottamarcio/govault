package sqlite

import (
	"context"
	"database/sql"
	"testing"
)

func TestMigrations(t *testing.T) {
	ctx := context.Background()
	db, err := Open("") // in-memory db
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Initial version should be 0
	v, err := db.CurrentVersion(ctx)
	if err != nil {
		t.Fatalf("CurrentVersion failed: %v", err)
	}
	if v != 0 {
		t.Fatalf("expected version 0, got %d", v)
	}

	// Run migrations
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	// Post migration version should be >= 1
	v, err = db.CurrentVersion(ctx)
	if err != nil {
		t.Fatalf("CurrentVersion failed: %v", err)
	}
	if v < 1 {
		t.Fatalf("expected version >= 1, got %d", v)
	}

	// Verify idempotency (re-running migrate is no-op)
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("subsequent Migrate failed: %v", err)
	}
}

func TestMigrationRollback(t *testing.T) {
	ctx := context.Background()
	db, err := Open("")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	// Snapshot existing migrations and restore after test
	originalMigrations := migrations
	defer func() {
		migrations = originalMigrations
	}()

	// Register a faulty migration with invalid SQL
	faultyMigration := Migration{
		Version: 999,
		Name:    "faulty_step",
		Up: func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "INVALID SQL SYNTAX STATEMENT;")
			return err
		},
	}
	RegisterMigration(faultyMigration)

	err = db.Migrate(ctx)
	if err == nil {
		t.Fatalf("expected migration to fail and roll back")
	}

	// Schema version must not have reached 999
	v, _ := db.CurrentVersion(ctx)
	if v >= 999 {
		t.Fatalf("version should not be 999 after rollback, got %d", v)
	}
}
