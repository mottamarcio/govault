package sqlite

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mottamarcio/govault/internal/crypto/cipher"
	"github.com/mottamarcio/govault/internal/crypto/kdf"
	"github.com/mottamarcio/govault/internal/crypto/keys"
)

func setupTestDB(t *testing.T) (*DB, func()) {
	t.Helper()
	db, err := Open("")
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	if err := db.Migrate(context.Background()); err != nil {
		db.Close()
		t.Fatalf("Migrate failed: %v", err)
	}

	return db, func() {
		db.Close()
	}
}

func sampleEnvelope() *keys.WrappedKeyEnvelope {
	params, _ := kdf.DefaultArgon2Params()
	return &keys.WrappedKeyEnvelope{
		KDFParams:  params,
		WrappedKey: make([]byte, cipher.MinPayloadSize+32),
		MCK:        make([]byte, 32),
	}
}

func TestMetadataCreateGet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewMetadataRepository(db)

	// 1. Get on empty db returns ErrVaultNotInitialized
	_, err := repo.Get(ctx)
	if !errors.Is(err, ErrVaultNotInitialized) {
		t.Fatalf("expected ErrVaultNotInitialized, got %v", err)
	}

	// 2. Create metadata
	env := sampleEnvelope()
	env.WrappedKey[0] = 0xAA
	meta := &VaultMetadata{
		VaultID:       "vault-uuid-001",
		SchemaVersion: 1,
		CryptoSuite:   "v1",
		Envelope:      env,
		CreatedAt:     time.Now().Truncate(time.Second),
		UpdatedAt:     time.Now().Truncate(time.Second),
	}

	if err := repo.Create(ctx, meta); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// 3. Get metadata
	retrieved, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if retrieved.VaultID != meta.VaultID {
		t.Errorf("vault ID mismatch: expected %s, got %s", meta.VaultID, retrieved.VaultID)
	}
	if retrieved.SchemaVersion != meta.SchemaVersion {
		t.Errorf("schema version mismatch: expected %d, got %d", meta.SchemaVersion, retrieved.SchemaVersion)
	}
	if retrieved.CryptoSuite != meta.CryptoSuite {
		t.Errorf("crypto suite mismatch: expected %s, got %s", meta.CryptoSuite, retrieved.CryptoSuite)
	}
	if !bytes.Equal(retrieved.Envelope.WrappedKey, meta.Envelope.WrappedKey) {
		t.Errorf("wrapped key mismatch")
	}
	if !bytes.Equal(retrieved.Envelope.MCK, meta.Envelope.MCK) {
		t.Errorf("MCK mismatch")
	}
}

func TestSingleVaultConstraint(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewMetadataRepository(db)

	meta1 := &VaultMetadata{
		VaultID:       "vault-uuid-001",
		SchemaVersion: 1,
		CryptoSuite:   "v1",
		Envelope:      sampleEnvelope(),
	}

	if err := repo.Create(ctx, meta1); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}

	meta2 := &VaultMetadata{
		VaultID:       "vault-uuid-002",
		SchemaVersion: 1,
		CryptoSuite:   "v1",
		Envelope:      sampleEnvelope(),
	}

	err := repo.Create(ctx, meta2)
	if !errors.Is(err, ErrVaultAlreadyExists) {
		t.Fatalf("expected ErrVaultAlreadyExists, got %v", err)
	}
}

func TestMetadataUpdateEnvelope(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewMetadataRepository(db)

	env1 := sampleEnvelope()
	env1.WrappedKey[0] = 0x11
	meta := &VaultMetadata{
		VaultID:       "vault-uuid-001",
		SchemaVersion: 1,
		CryptoSuite:   "v1",
		Envelope:      env1,
	}

	if err := repo.Create(ctx, meta); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update envelope (mock password rotation)
	env2 := sampleEnvelope()
	env2.WrappedKey[0] = 0x99
	env2.MCK[0] = 0x88

	if err := repo.UpdateEnvelope(ctx, meta.VaultID, env2); err != nil {
		t.Fatalf("UpdateEnvelope failed: %v", err)
	}

	retrieved, err := repo.Get(ctx)
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	if retrieved.Envelope.WrappedKey[0] != 0x99 || retrieved.Envelope.MCK[0] != 0x88 {
		t.Fatalf("updated envelope values not persisted properly")
	}

	// Non-existent vault update fails
	err = repo.UpdateEnvelope(ctx, "non-existent-id", env2)
	if err == nil {
		t.Fatalf("expected update on non-existent vault to fail")
	}
}

func TestMetadataInspect(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewMetadataRepository(db)

	// Inspect before init
	_, err := repo.Inspect(ctx)
	if !errors.Is(err, ErrVaultNotInitialized) {
		t.Fatalf("expected ErrVaultNotInitialized from Inspect, got %v", err)
	}

	meta := &VaultMetadata{
		VaultID:       "vault-uuid-inspect",
		SchemaVersion: 1,
		CryptoSuite:   "v1",
		Envelope:      sampleEnvelope(),
	}
	if err := repo.Create(ctx, meta); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	info, err := repo.Inspect(ctx)
	if err != nil {
		t.Fatalf("Inspect failed: %v", err)
	}

	if info.VaultID != meta.VaultID || info.CryptoSuite != "v1" || info.SchemaVersion != 1 {
		t.Fatalf("inspected info mismatch: %+v", info)
	}
}
