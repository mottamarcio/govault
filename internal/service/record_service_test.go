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

func setupTestRecordService(t *testing.T) (*service.VaultService, *service.RecordService, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "govault-record-service-test-*")
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
	tagRepo := sqlite.NewTagRepository(db)
	recordRepo := sqlite.NewRecordRepository(db, tagRepo)
	historyRepo := sqlite.NewHistoryRepository(db)

	vaultService := service.NewVaultService(db, metaRepo, dbPath)
	recordService := service.NewRecordService(vaultService.Session(), recordRepo, tagRepo, historyRepo)

	cleanup := func() {
		_ = db.Close()
		_ = os.RemoveAll(tempDir)
	}

	return vaultService, recordService, cleanup
}

func TestRecordServiceCRUD(t *testing.T) {
	vaultSvc, recSvc, cleanup := setupTestRecordService(t)
	defer cleanup()

	ctx := context.Background()

	// 1. Initial actions when locked should fail
	_, err := recSvc.Create(ctx, service.RecordInput{
		Type:    domain.RecordTypeLogin,
		Payload: domain.LoginPayload{Username: "test"},
	})
	if err != service.ErrVaultLocked {
		t.Fatalf("expected ErrVaultLocked, got %v", err)
	}

	// 2. Init and unlock vault
	if err := vaultSvc.Init(ctx, "masterpassword123"); err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if _, err := vaultSvc.Unlock(ctx, "masterpassword123"); err != nil {
		t.Fatalf("Unlock failed: %v", err)
	}

	// 3. Create Login record
	loginInput := service.RecordInput{
		Title: "Personal Gmail",
		Type:  domain.RecordTypeLogin,
		Tags:  []string{"personal", "email"},
		Payload: domain.LoginPayload{
			Username: "me@gmail.com",
			Password: "strongPassword!@#",
			URI:      "https://mail.google.com",
			Notes:    "2FA enabled",
		},
	}
	createdLogin, err := recSvc.Create(ctx, loginInput)
	if err != nil {
		t.Fatalf("Create login failed: %v", err)
	}
	if createdLogin.ID == "" || createdLogin.Version != 1 {
		t.Fatalf("unexpected created record: %+v", createdLogin)
	}

	// 4. Retrieve by ID
	fetchedLogin, err := recSvc.GetByID(ctx, createdLogin.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	payload, ok := fetchedLogin.Payload.(domain.LoginPayload)
	if !ok || payload.Username != "me@gmail.com" || payload.Password != "strongPassword!@#" {
		t.Fatalf("decrypted payload mismatch: %+v", fetchedLogin.Payload)
	}
	if len(fetchedLogin.Tags) != 2 {
		t.Fatalf("tags mismatch: %+v", fetchedLogin.Tags)
	}

	// 5. Create Note record
	noteInput := service.RecordInput{
		Type: domain.RecordTypeNote,
		Tags: []string{"work"},
		Payload: domain.NotePayload{
			Content: "Top Secret Strategy\nConfidential notes",
		},
	}
	createdNote, err := recSvc.Create(ctx, noteInput)
	if err != nil {
		t.Fatalf("Create note failed: %v", err)
	}

	// 6. List records
	records, err := recSvc.List(ctx, false)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// 7. Search records in-memory
	searchResults, err := recSvc.Search(ctx, service.SearchFilter{
		Query: "Strategy",
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(searchResults) != 1 || searchResults[0].ID != createdNote.ID {
		t.Fatalf("search result mismatch: %+v", searchResults)
	}
}

func TestRecordServiceHistoryAndTrash(t *testing.T) {
	vaultSvc, recSvc, cleanup := setupTestRecordService(t)
	defer cleanup()

	ctx := context.Background()
	_ = vaultSvc.Init(ctx, "mypassword")
	_, _ = vaultSvc.Unlock(ctx, "mypassword")

	// 1. Create initial record (Version 1)
	input := service.RecordInput{
		Type: domain.RecordTypeAPIKey,
		Tags: []string{"aws"},
		Payload: domain.APIKeyPayload{
			Service: "AWS",
			Key:     "AKIA_V1",
			Secret:  "SECRET_V1",
		},
	}
	rec, err := recSvc.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 2. Update to Version 2
	updateInput := service.RecordInput{
		Type: domain.RecordTypeAPIKey,
		Tags: []string{"aws", "prod"},
		Payload: domain.APIKeyPayload{
			Service: "AWS",
			Key:     "AKIA_V2",
			Secret:  "SECRET_V2",
		},
	}
	updatedRec, err := recSvc.Update(ctx, rec.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updatedRec.Version != 2 {
		t.Fatalf("expected version 2, got %d", updatedRec.Version)
	}

	// 3. Inspect history
	history, err := recSvc.ListHistory(ctx, rec.ID)
	if err != nil {
		t.Fatalf("ListHistory failed: %v", err)
	}
	if len(history) != 1 || history[0].Version != 1 {
		t.Fatalf("expected 1 history entry with version 1, got %+v", history)
	}

	histV1, err := recSvc.GetHistoryRevision(ctx, rec.ID, 1)
	if err != nil {
		t.Fatalf("GetHistoryRevision failed: %v", err)
	}
	v1Payload := histV1.Payload.(domain.APIKeyPayload)
	if v1Payload.Key != "AKIA_V1" {
		t.Fatalf("historical payload mismatch: got %s, want AKIA_V1", v1Payload.Key)
	}

	// 4. Soft Delete
	if err := recSvc.Delete(ctx, rec.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	activeList, _ := recSvc.List(ctx, false)
	if len(activeList) != 0 {
		t.Fatalf("expected 0 active records, got %d", len(activeList))
	}

	trashList, err := recSvc.ListTrash(ctx)
	if err != nil || len(trashList) != 1 {
		t.Fatalf("expected 1 trash record, got %d (err: %v)", len(trashList), err)
	}

	// 5. Restore
	if err := recSvc.Restore(ctx, rec.ID); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}

	restoredList, _ := recSvc.List(ctx, false)
	if len(restoredList) != 1 {
		t.Fatalf("expected 1 active record after restore, got %d", len(restoredList))
	}

	// 6. Delete and Purge
	_ = recSvc.Delete(ctx, rec.ID)
	if err := recSvc.PurgeTrash(ctx); err != nil {
		t.Fatalf("PurgeTrash failed: %v", err)
	}

	finalTrash, _ := recSvc.ListTrash(ctx)
	if len(finalTrash) != 0 {
		t.Fatalf("expected empty trash after purge, got %d", len(finalTrash))
	}
}
