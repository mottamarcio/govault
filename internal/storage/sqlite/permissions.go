package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsurePermissions ensures that the parent directory has 0700 permissions
// and the database file (plus WAL/SHM files if existing) has 0600 permissions.
func EnsurePermissions(dbPath string) error {
	if dbPath == "" || dbPath == ":memory:" {
		return nil
	}

	// 1. Ensure parent directory permissions (0700)
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create parent directory: %w", err)
	}
	if err := os.Chmod(dir, 0700); err != nil {
		return fmt.Errorf("failed to set 0700 permissions on directory %s: %w", dir, err)
	}

	// 2. Ensure database file permissions (0600) if it exists
	files := []string{
		dbPath,
		dbPath + "-wal",
		dbPath + "-shm",
	}

	for _, f := range files {
		if info, err := os.Stat(f); err == nil && !info.IsDir() {
			if err := os.Chmod(f, 0600); err != nil {
				return fmt.Errorf("failed to set 0600 permissions on file %s: %w", f, err)
			}
		}
	}

	return nil
}
