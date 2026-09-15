package sqlite

import (
	"context"
	"testing"
)

func TestSchemaV1TablesExist(t *testing.T) {
	ctx := context.Background()
	db, err := Open("")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	requiredTables := []string{
		"vault_metadata",
		"records",
		"record_history",
		"tags",
		"record_tags",
		"schema_version",
	}

	for _, table := range requiredTables {
		var count int
		err := db.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?;", table).Scan(&count)
		if err != nil {
			t.Fatalf("failed to query table %s: %v", table, err)
		}
		if count != 1 {
			t.Errorf("expected table %q to exist, but not found", table)
		}
	}
}
