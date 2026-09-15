package sqlite

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilePermissions(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "govault-perm-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "subdir", "vault.db")

	// Ensure permissions creates directory with 0700
	if err := EnsurePermissions(dbPath); err != nil {
		t.Fatalf("EnsurePermissions failed: %v", err)
	}

	dirInfo, err := os.Stat(filepath.Dir(dbPath))
	if err != nil {
		t.Fatalf("failed to stat parent dir: %v", err)
	}
	if dirInfo.Mode().Perm() != 0700 {
		t.Errorf("expected dir permissions 0700, got %o", dirInfo.Mode().Perm())
	}

	// Create dummy db and wal files
	if err := os.WriteFile(dbPath, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	walPath := dbPath + "-wal"
	if err := os.WriteFile(walPath, []byte("wal"), 0644); err != nil {
		t.Fatalf("failed to write test wal file: %v", err)
	}

	if err := EnsurePermissions(dbPath); err != nil {
		t.Fatalf("EnsurePermissions failed: %v", err)
	}

	fileInfo, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("failed to stat db file: %v", err)
	}
	if fileInfo.Mode().Perm() != 0600 {
		t.Errorf("expected db file permissions 0600, got %o", fileInfo.Mode().Perm())
	}

	walInfo, err := os.Stat(walPath)
	if err != nil {
		t.Fatalf("failed to stat wal file: %v", err)
	}
	if walInfo.Mode().Perm() != 0600 {
		t.Errorf("expected wal file permissions 0600, got %o", walInfo.Mode().Perm())
	}
}
