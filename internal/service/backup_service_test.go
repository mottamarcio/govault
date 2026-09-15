package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
)

func setupTestBackupEnv(t *testing.T) (*sqlite.DB, *service.VaultService, *service.RecordService, *service.BackupService, string, string) {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test-vault.db")

	ctx := context.Background()
	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test sqlite: %v", err)
	}

	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	metaRepo := sqlite.NewMetadataRepository(db)
	tagRepo := sqlite.NewTagRepository(db)
	recordRepo := sqlite.NewRecordRepository(db, tagRepo)
	historyRepo := sqlite.NewHistoryRepository(db)

	vaultService := service.NewVaultService(db, metaRepo, dbPath)

	password := "MasterSecret123!"
	if err := vaultService.Init(ctx, password); err != nil {
		t.Fatalf("failed to init vault: %v", err)
	}

	session, err := vaultService.Unlock(ctx, password)
	if err != nil {
		t.Fatalf("failed to unlock vault: %v", err)
	}

	recordService := service.NewRecordService(session, recordRepo, tagRepo, historyRepo)
	backupService := service.NewBackupService(metaRepo, recordRepo, tagRepo, historyRepo, dbPath)

	return db, vaultService, recordService, backupService, dbPath, password
}

func TestBackupExportAndVerify(t *testing.T) {
	db, _, recordService, backupService, _, password := setupTestBackupEnv(t)
	defer db.Close()

	ctx := context.Background()
	tempDir := t.TempDir()
	backupPath := filepath.Join(tempDir, "vault-export.gvault")

	// 1. Create a secret record
	loginPayload := domain.LoginPayload{
		Username: "alice",
		Password: "AliceSecretPassword123!",
		URI:      "https://example.com",
	}
	_, err := recordService.Create(ctx, service.RecordInput{
		Title:   "Alice Login",
		Type:    domain.RecordTypeLogin,
		Tags:    []string{"Personal", "Email"},
		Payload: loginPayload,
	})
	if err != nil {
		t.Fatalf("failed to create record: %v", err)
	}

	// 2. Export backup with active master password
	sess := recordService.Session()
	err = backupService.CreateBackup(ctx, sess, backupPath, service.BackupCreateOptions{})
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(backupPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("expected 0600 permissions, got %o", info.Mode().Perm())
	}

	// 3. Structural verification without password
	resUnauth, err := backupService.VerifyBackup(ctx, backupPath, "")
	if err != nil {
		t.Fatalf("structural verify failed: %v", err)
	}
	if resUnauth.IsVerified {
		t.Fatalf("expected IsVerified=false for unauthenticated check")
	}
	if resUnauth.Inspection == nil || resUnauth.Inspection.FormatVersion != 1 {
		t.Fatalf("invalid inspection header")
	}

	// 4. Cryptographic verification with password
	resAuth, err := backupService.VerifyBackup(ctx, backupPath, password)
	if err != nil {
		t.Fatalf("cryptographic verify failed: %v", err)
	}
	if !resAuth.IsVerified {
		t.Fatalf("expected IsVerified=true for authenticated check")
	}
	if resAuth.Manifest == nil || resAuth.Manifest.EntryCount != 1 {
		t.Fatalf("manifest entry count mismatch: got %v", resAuth.Manifest)
	}

	// 5. Verification with wrong password fails
	_, err = backupService.VerifyBackup(ctx, backupPath, "WrongPassword999!")
	if err != service.ErrInvalidPassphrase {
		t.Fatalf("expected ErrInvalidPassphrase, got %v", err)
	}

	// 6. Export with custom export passphrase
	customBackupPath := filepath.Join(tempDir, "vault-custom.gvault")
	customPassphrase := "OneTimeExportPassphrase456!"
	err = backupService.CreateBackup(ctx, sess, customBackupPath, service.BackupCreateOptions{
		ExportPassphrase: customPassphrase,
	})
	if err != nil {
		t.Fatalf("failed to create custom passphrase backup: %v", err)
	}

	resCustom, err := backupService.VerifyBackup(ctx, customBackupPath, customPassphrase)
	if err != nil || !resCustom.IsVerified {
		t.Fatalf("failed to verify custom passphrase backup: %v", err)
	}
}

func TestBackupRestore(t *testing.T) {
	db, vaultService, recordService, backupService, dbPath, password := setupTestBackupEnv(t)
	defer db.Close()

	ctx := context.Background()
	tempDir := t.TempDir()
	backupPath := filepath.Join(tempDir, "restore-test.gvault")

	// 1. Create a record
	_, err := recordService.Create(ctx, service.RecordInput{
		ID:      "test-rec-1",
		Title:   "Important Note",
		Type:    domain.RecordTypeNote,
		Tags:    []string{"Notes"},
		Payload: domain.NotePayload{Content: "Important Note\nSecret Content"},
	})
	if err != nil {
		t.Fatalf("failed to create record: %v", err)
	}

	// Export
	sess := recordService.Session()
	err = backupService.CreateBackup(ctx, sess, backupPath, service.BackupCreateOptions{})
	if err != nil {
		t.Fatalf("failed to create backup: %v", err)
	}

	// 2. Modify original record
	_, err = recordService.Create(ctx, service.RecordInput{
		ID:      "test-rec-2",
		Title:   "Temporary Note",
		Type:    domain.RecordTypeNote,
		Payload: domain.NotePayload{Content: "Will be overwritten"},
	})
	if err != nil {
		t.Fatalf("failed to create second record: %v", err)
	}

	// Lock and close DB before restore replacement
	_ = vaultService.Lock()
	_ = db.Close()

	// 3. Restore backup
	err = backupService.RestoreBackup(ctx, backupPath, password, service.BackupRestoreOptions{})
	if err != nil {
		t.Fatalf("restore backup failed: %v", err)
	}

	// 4. Reopen and verify restored state
	restoredDB, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to reopen restored db: %v", err)
	}
	defer restoredDB.Close()

	metaRepo := sqlite.NewMetadataRepository(restoredDB)
	tagRepo := sqlite.NewTagRepository(restoredDB)
	recordRepo := sqlite.NewRecordRepository(restoredDB, tagRepo)
	historyRepo := sqlite.NewHistoryRepository(restoredDB)

	newVaultService := service.NewVaultService(restoredDB, metaRepo, dbPath)
	newSession, err := newVaultService.Unlock(ctx, password)
	if err != nil {
		t.Fatalf("failed to unlock restored vault: %v", err)
	}

	newRecordService := service.NewRecordService(newSession, recordRepo, tagRepo, historyRepo)

	records, err := newRecordService.List(ctx, false)
	if err != nil {
		t.Fatalf("failed to list restored records: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 restored record, got %d", len(records))
	}
	if records[0].ID != "test-rec-1" {
		t.Fatalf("restored record ID mismatch: %s", records[0].ID)
	}
}

func TestBackupMerge(t *testing.T) {
	db, _, recordService, backupService, _, password := setupTestBackupEnv(t)
	defer db.Close()

	ctx := context.Background()
	tempDir := t.TempDir()
	backupPath := filepath.Join(tempDir, "merge-test.gvault")

	// 1. Create a record in first vault and export
	_, err := recordService.Create(ctx, service.RecordInput{
		ID:      "conflict-rec",
		Title:   "Backup Record",
		Type:    domain.RecordTypeLogin,
		Tags:    []string{"ImportedTag"},
		Payload: domain.LoginPayload{Username: "backup_user"},
	})
	if err != nil {
		t.Fatalf("failed to create record: %v", err)
	}

	sess := recordService.Session()
	err = backupService.CreateBackup(ctx, sess, backupPath, service.BackupCreateOptions{})
	if err != nil {
		t.Fatalf("failed to export backup: %v", err)
	}

	// 2. Modify existing record in active vault
	_, err = recordService.Create(ctx, service.RecordInput{
		ID:      "other-rec",
		Title:   "Local Record",
		Type:    domain.RecordTypeNote,
		Payload: domain.NotePayload{Content: "Local only"},
	})
	if err != nil {
		t.Fatalf("failed to create local record: %v", err)
	}

	// 3. Test Merge with ConflictSkip
	mergeRes, err := backupService.MergeBackup(ctx, sess, backupPath, password, service.ConflictSkip)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if mergeRes.SkippedCount != 1 {
		t.Fatalf("expected 1 skipped record, got %d", mergeRes.SkippedCount)
	}

	// 4. Test Merge with ConflictRename
	mergeResRename, err := backupService.MergeBackup(ctx, sess, backupPath, password, service.ConflictRename)
	if err != nil {
		t.Fatalf("merge rename failed: %v", err)
	}
	if mergeResRename.RenamedCount != 1 || mergeResRename.ImportedCount != 1 {
		t.Fatalf("expected 1 renamed and 1 imported, got %+v", mergeResRename)
	}
}
