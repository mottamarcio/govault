package sqlite

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/mottamarcio/govault/internal/crypto/cipher"
)

func TestTagRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	vaultID := "vault-tag-test"

	// 1. Create tags
	t1, err := tagRepo.Create(ctx, vaultID, "finance")
	if err != nil {
		t.Fatalf("Create tag failed: %v", err)
	}
	if t1.Name != "finance" {
		t.Errorf("expected tag name finance, got %s", t1.Name)
	}

	// Idempotent creation (same name returns existing)
	t1dup, err := tagRepo.Create(ctx, vaultID, "finance")
	if err != nil {
		t.Fatalf("Create duplicate tag failed: %v", err)
	}
	if t1dup.ID != t1.ID {
		t.Errorf("expected duplicate create to return same tag ID")
	}

	t2, _ := tagRepo.Create(ctx, vaultID, "work")

	// 2. List tags
	tags, err := tagRepo.List(ctx, vaultID)
	if err != nil {
		t.Fatalf("List tags failed: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}

	// 3. Delete tag
	if err := tagRepo.Delete(ctx, vaultID, t2.ID); err != nil {
		t.Fatalf("Delete tag failed: %v", err)
	}

	tags, _ = tagRepo.List(ctx, vaultID)
	if len(tags) != 1 {
		t.Fatalf("expected 1 tag after deletion, got %d", len(tags))
	}
}

func TestRecordCRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	recordRepo := NewRecordRepository(db, tagRepo)
	vaultID := "vault-record-test"

	t1, _ := tagRepo.Create(ctx, vaultID, "email")
	t2, _ := tagRepo.Create(ctx, vaultID, "personal")

	payload1 := make([]byte, cipher.MinPayloadSize+20)
	payload1[0] = 0xAA

	rec := &Record{
		VaultID:    vaultID,
		RecordType: "login",
		Payload:    payload1,
	}

	// 1. Create record with tags
	if err := recordRepo.Create(ctx, rec, []string{t1.ID, t2.ID}); err != nil {
		t.Fatalf("Create record failed: %v", err)
	}

	if rec.ID == "" || rec.Version != 1 {
		t.Fatalf("expected generated ID and version 1")
	}

	// 2. GetByID
	fetched, err := recordRepo.GetByID(ctx, vaultID, rec.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if !bytes.Equal(fetched.Payload, payload1) {
		t.Errorf("payload mismatch")
	}

	// Verify associated tags
	tagIDs, err := tagRepo.GetRecordTags(ctx, rec.ID)
	if err != nil {
		t.Fatalf("GetRecordTags failed: %v", err)
	}
	if len(tagIDs) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tagIDs))
	}

	// 3. Soft Delete
	if err := recordRepo.SoftDelete(ctx, vaultID, rec.ID); err != nil {
		t.Fatalf("SoftDelete failed: %v", err)
	}

	// Excluded from normal Get and List
	if _, err := recordRepo.GetByID(ctx, vaultID, rec.ID); !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound for soft-deleted record")
	}

	listActive, _ := recordRepo.List(ctx, vaultID, false)
	if len(listActive) != 0 {
		t.Fatalf("expected 0 active records")
	}

	listAll, _ := recordRepo.List(ctx, vaultID, true)
	if len(listAll) != 1 || listAll[0].DeletedAt == nil {
		t.Fatalf("expected 1 deleted record in full list")
	}

	// 4. Restore
	if err := recordRepo.Restore(ctx, vaultID, rec.ID); err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	listActive, _ = recordRepo.List(ctx, vaultID, false)
	if len(listActive) != 1 {
		t.Fatalf("expected 1 active record after restore")
	}

	// 5. Purge Deleted
	recordRepo.SoftDelete(ctx, vaultID, rec.ID)
	purged, err := recordRepo.PurgeDeleted(ctx, vaultID)
	if err != nil {
		t.Fatalf("PurgeDeleted failed: %v", err)
	}
	if purged != 1 {
		t.Fatalf("expected 1 record purged, got %d", purged)
	}

	listAll, _ = recordRepo.List(ctx, vaultID, true)
	if len(listAll) != 0 {
		t.Fatalf("expected 0 records after purge")
	}
}

func TestRecordUpdateAndHistory(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	tagRepo := NewTagRepository(db)
	recordRepo := NewRecordRepository(db, tagRepo)
	historyRepo := NewHistoryRepository(db)
	vaultID := "vault-history-test"

	payloadV1 := make([]byte, cipher.MinPayloadSize+10)
	payloadV1[0] = 0x01

	rec := &Record{
		VaultID:    vaultID,
		RecordType: "login",
		Payload:    payloadV1,
	}

	if err := recordRepo.Create(ctx, rec, nil); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 1. First Update (Version 1 -> Version 2)
	payloadV2 := make([]byte, cipher.MinPayloadSize+10)
	payloadV2[0] = 0x02
	rec.Payload = payloadV2

	if err := recordRepo.Update(ctx, rec, nil); err != nil {
		t.Fatalf("Update v1->v2 failed: %v", err)
	}
	if rec.Version != 2 {
		t.Fatalf("expected record version 2, got %d", rec.Version)
	}

	// 2. Second Update (Version 2 -> Version 3)
	payloadV3 := make([]byte, cipher.MinPayloadSize+10)
	payloadV3[0] = 0x03
	rec.Payload = payloadV3

	if err := recordRepo.Update(ctx, rec, nil); err != nil {
		t.Fatalf("Update v2->v3 failed: %v", err)
	}
	if rec.Version != 3 {
		t.Fatalf("expected record version 3, got %d", rec.Version)
	}

	// 3. Query History
	histories, err := historyRepo.ListByRecordID(ctx, rec.ID)
	if err != nil {
		t.Fatalf("ListByRecordID failed: %v", err)
	}
	if len(histories) != 2 {
		t.Fatalf("expected 2 history entries (v1, v2), got %d", len(histories))
	}

	if histories[0].Version != 2 || histories[1].Version != 1 {
		t.Errorf("history versions order mismatch: %d, %d", histories[0].Version, histories[1].Version)
	}
	if !bytes.Equal(histories[1].Payload, payloadV1) {
		t.Errorf("history v1 payload mismatch")
	}

	// 4. Optimistic locking conflict test
	staleRec := &Record{
		ID:         rec.ID,
		VaultID:    vaultID,
		RecordType: "login",
		Version:    1, // Stale version! Current is 3
		Payload:    payloadV1,
	}
	err = recordRepo.Update(ctx, staleRec, nil)
	if !errors.Is(err, ErrOptimisticLockingConflict) {
		t.Fatalf("expected ErrOptimisticLockingConflict, got %v", err)
	}
}
