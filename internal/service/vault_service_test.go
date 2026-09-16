package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
)

func setupTestVaultService(t *testing.T) (*service.VaultService, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "govault-service-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "vault.db")
	db, err := sqlite.Open(dbPath)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to open test db: %v", err)
	}

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		db.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("failed to apply migrations: %v", err)
	}

	metaRepo := sqlite.NewMetadataRepository(db)
	vaultService := service.NewVaultService(db, metaRepo, dbPath)

	cleanup := func() {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
	}

	return vaultService, cleanup
}

func TestVaultServiceLifecycle(t *testing.T) {
	svc, cleanup := setupTestVaultService(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Initial status should be uninitialized
	_, err := svc.Inspect(ctx)
	if err != service.ErrVaultNotInitialized {
		t.Fatalf("expected ErrVaultNotInitialized, got %v", err)
	}

	// 2. Unlock uninitialized should fail
	_, err = svc.Unlock(ctx, "masterpassword")
	if err != service.ErrVaultNotInitialized {
		t.Fatalf("expected ErrVaultNotInitialized on unlock, got %v", err)
	}

	// 3. Init with empty password fails
	if err := svc.Init(ctx, ""); err != service.ErrEmptyPassword {
		t.Fatalf("expected ErrEmptyPassword, got %v", err)
	}

	// 4. Init with valid password
	if err := svc.Init(ctx, "correct-horse-battery-staple"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// 5. Init again fails
	if err := svc.Init(ctx, "another-pass"); err != service.ErrVaultAlreadyExists {
		t.Fatalf("expected ErrVaultAlreadyExists, got %v", err)
	}

	// 6. Inspect succeeds
	header, err := svc.Inspect(ctx)
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}
	if header.VaultID == "" || header.SchemaVersion != 1 {
		t.Fatalf("unexpected header: %+v", header)
	}

	// 7. Unlock with wrong password fails
	_, err = svc.Unlock(ctx, "wrong-password")
	if err != service.ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}
	if svc.IsUnlocked() {
		t.Fatal("vault should remain locked after failed unlock")
	}

	// 8. Unlock with correct password succeeds
	session, err := svc.Unlock(ctx, "correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
	if !session.IsUnlocked() || !svc.IsUnlocked() {
		t.Fatal("vault should be unlocked")
	}

	vaultKey, err := session.VaultKey()
	if err != nil || len(vaultKey) != 32 {
		t.Fatalf("VaultKey() returned invalid key: len %d, err %v", len(vaultKey), err)
	}

	// 9. Lock
	if err := svc.Lock(); err != nil {
		t.Fatalf("Lock failed: %v", err)
	}
	if svc.IsUnlocked() {
		t.Fatal("vault should be locked after Lock()")
	}
}

func TestVaultPasswordChangeAndPurge(t *testing.T) {
	svc, cleanup := setupTestVaultService(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Init
	masterPass := "old-secret-pass-123"
	if err := svc.Init(ctx, masterPass); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Get initial vault key
	sess1, err := svc.Unlock(ctx, masterPass)
	if err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}
	key1, _ := sess1.VaultKey()
	svc.Lock()

	// 2. Change password with wrong old password fails
	newPass := "new-ultra-secret-pass-456"
	err = svc.ChangeMasterPassword(ctx, "wrong-old-pass", newPass)
	if err != service.ErrInvalidPassword {
		t.Fatalf("expected ErrInvalidPassword, got %v", err)
	}

	// 3. Change password with correct old password succeeds
	if err := svc.ChangeMasterPassword(ctx, masterPass, newPass); err != nil {
		t.Fatalf("ChangeMasterPassword failed: %v", err)
	}

	// 4. Unlock with old password fails
	_, err = svc.Unlock(ctx, masterPass)
	if err != service.ErrInvalidPassword {
		t.Fatalf("expected old password to be rejected, got %v", err)
	}

	// 5. Unlock with new password succeeds and unwraps identical VaultKey
	sess2, err := svc.Unlock(ctx, newPass)
	if err != nil {
		t.Fatalf("Unlock with new password failed: %v", err)
	}
	key2, _ := sess2.VaultKey()

	if string(key1) != string(key2) {
		t.Fatal("VaultKey changed across master password rotation")
	}

	// 6. Purge
	if err := svc.Purge(ctx); err != nil {
		t.Fatalf("Purge failed: %v", err)
	}
}

func TestVaultServiceIsInitialized(t *testing.T) {
	svc, cleanup := setupTestVaultService(t)
	defer cleanup()

	ctx := context.Background()

	// Initial status should be false
	init, err := svc.IsInitialized(ctx)
	if err != nil {
		t.Fatalf("unexpected error checking IsInitialized: %v", err)
	}
	if init {
		t.Fatal("expected IsInitialized to return false on fresh vault")
	}

	// Initialize vault
	if err := svc.Init(ctx, "correct-horse-battery-staple"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	// Status should now be true
	init, err = svc.IsInitialized(ctx)
	if err != nil {
		t.Fatalf("unexpected error checking IsInitialized: %v", err)
	}
	if !init {
		t.Fatal("expected IsInitialized to return true after Init")
	}
}

