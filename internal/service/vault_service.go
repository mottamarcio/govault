package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mottamarcio/govault/internal/crypto/kdf"
	"github.com/mottamarcio/govault/internal/crypto/keys"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
)

// VaultService orchestrates vault lifecycle management, encryption envelope updates, and sessions.
type VaultService struct {
	db       *sqlite.DB
	metaRepo *sqlite.MetadataRepository
	session  *Session
	dbPath   string
}

// NewVaultService creates a new VaultService instance.
func NewVaultService(db *sqlite.DB, metaRepo *sqlite.MetadataRepository, dbPath string) *VaultService {
	return &VaultService{
		db:       db,
		metaRepo: metaRepo,
		session:  NewSession(),
		dbPath:   dbPath,
	}
}

// Session returns the active session manager.
func (s *VaultService) Session() *Session {
	return s.session
}

// IsUnlocked returns true if the vault is unlocked.
func (s *VaultService) IsUnlocked() bool {
	return s.session.IsUnlocked()
}

// Init initializes a fresh vault with master password.
func (s *VaultService) Init(ctx context.Context, masterPassword string) error {
	if strings.TrimSpace(masterPassword) == "" {
		return ErrEmptyPassword
	}

	// Check if already initialized
	_, err := s.metaRepo.Get(ctx)
	if err == nil {
		return ErrVaultAlreadyExists
	}
	if !errors.Is(err, sqlite.ErrVaultNotInitialized) {
		return fmt.Errorf("failed to check vault status: %w", err)
	}

	// 1. Generate VaultID (UUID)
	var uuidBuf [16]byte
	if _, err := rand.Read(uuidBuf[:]); err != nil {
		return fmt.Errorf("failed to generate random vault ID: %w", err)
	}
	vaultID := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuidBuf[0:4], uuidBuf[4:6], uuidBuf[6:8], uuidBuf[8:10], uuidBuf[10:16])

	// 2. Generate random 32-byte VaultKey
	vaultKey, err := keys.GenerateVaultKey()
	if err != nil {
		return fmt.Errorf("failed to generate vault key: %w", err)
	}
	defer kdf.Zeroize(vaultKey)

	// 3. Default Argon2id parameters
	params, err := kdf.DefaultArgon2Params()
	if err != nil {
		return fmt.Errorf("failed to generate argon2 params: %w", err)
	}

	// 4. Wrap vault key
	envelope, err := keys.WrapVaultKey(vaultKey, []byte(masterPassword), vaultID, params)
	if err != nil {
		return fmt.Errorf("failed to wrap vault key: %w", err)
	}

	// 5. Persist metadata
	meta := &sqlite.VaultMetadata{
		VaultID:       vaultID,
		SchemaVersion: 1,
		CryptoSuite:   "Argon2id+XChaCha20-Poly1305+HKDF",
		Envelope:      envelope,
	}

	if err := s.metaRepo.Create(ctx, meta); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

// Unlock authenticates against the vault envelope and unlocks the session.
func (s *VaultService) Unlock(ctx context.Context, masterPassword string) (*Session, error) {
	if strings.TrimSpace(masterPassword) == "" {
		return nil, ErrEmptyPassword
	}

	meta, err := s.metaRepo.Get(ctx)
	if err != nil {
		if errors.Is(err, sqlite.ErrVaultNotInitialized) {
			return nil, ErrVaultNotInitialized
		}
		return nil, fmt.Errorf("failed to fetch vault metadata: %w", err)
	}

	// Unwrap VaultKey
	vaultKey, err := keys.UnwrapVaultKey(meta.Envelope, []byte(masterPassword), meta.VaultID)
	if err != nil {
		if errors.Is(err, keys.ErrAuthenticationFailed) {
			return nil, ErrInvalidPassword
		}
		return nil, ErrInvalidPassword
	}
	defer kdf.Zeroize(vaultKey)

	s.session.Unlock(meta.VaultID, vaultKey)
	return s.session, nil
}

// Lock locks the active session and zeroizes the key.
func (s *VaultService) Lock() error {
	s.session.Lock()
	return nil
}

// Inspect returns public vault header information without unlocking.
func (s *VaultService) Inspect(ctx context.Context) (*sqlite.VaultHeaderInfo, error) {
	header, err := s.metaRepo.Inspect(ctx)
	if err != nil {
		if errors.Is(err, sqlite.ErrVaultNotInitialized) {
			return nil, ErrVaultNotInitialized
		}
		return nil, fmt.Errorf("failed to inspect metadata: %w", err)
	}
	return header, nil
}

// ChangeMasterPassword rotates the master password envelope without re-encrypting records.
func (s *VaultService) ChangeMasterPassword(ctx context.Context, oldPassword, newPassword string) error {
	if strings.TrimSpace(oldPassword) == "" || strings.TrimSpace(newPassword) == "" {
		return ErrEmptyPassword
	}

	meta, err := s.metaRepo.Get(ctx)
	if err != nil {
		if errors.Is(err, sqlite.ErrVaultNotInitialized) {
			return ErrVaultNotInitialized
		}
		return fmt.Errorf("failed to fetch vault metadata: %w", err)
	}

	newParams, err := kdf.DefaultArgon2Params()
	if err != nil {
		return fmt.Errorf("failed to generate new argon2 params: %w", err)
	}

	rotatedEnvelope, err := keys.RotateMasterPassword(meta.Envelope, []byte(oldPassword), []byte(newPassword), meta.VaultID, newParams)
	if err != nil {
		return ErrInvalidPassword
	}

	if err := s.metaRepo.UpdateEnvelope(ctx, meta.VaultID, rotatedEnvelope); err != nil {
		return fmt.Errorf("failed to update metadata envelope: %w", err)
	}

	return nil
}

// Purge closes database connection and removes the vault file and WAL/SHM files.
func (s *VaultService) Purge(ctx context.Context) error {
	s.session.Lock()

	if s.db != nil {
		_ = s.db.Close()
	}

	if s.dbPath != "" && s.dbPath != ":memory:" {
		_ = os.Remove(s.dbPath)
		_ = os.Remove(s.dbPath + "-wal")
		_ = os.Remove(s.dbPath + "-shm")
	}

	return nil
}
