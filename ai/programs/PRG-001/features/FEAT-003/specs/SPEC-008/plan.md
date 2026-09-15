---
type: plan
for: SPEC-008
status: ready
---

# Implementation Plan: Vault Lifecycle Application Service & Session Management

## Summary

Design and implement `VaultService` and thread-safe in-memory `Session` manager under `internal/service`, orchestrating initialization, unlock, lock with key zeroization, master password rotation, and vault database wiping/purging.

## Repository Context

- Cryptography: `internal/crypto/kdf` (Argon2id KDF, MCK verification, zeroization), `internal/crypto/keys` (VaultKey generation, wrapping, unwrapping, rotation), `internal/crypto/cipher` (XChaCha20-Poly1305).
- Storage: `internal/storage/sqlite` (`DB`, `MetadataRepository`, migrations).
- Domain: `internal/domain` (Entities, errors, types).

## Requirement Coverage

- **R1 → Vault Initialization:**
  - Implement `VaultService.Init(ctx, masterPassword string) (*domain.Vault, error)`:
    - Generate random UUID for `VaultID`.
    - Generate random 32-byte `VaultKey` via `keys.GenerateVaultKey()`.
    - Generate default Argon2id parameters with random 32-byte salt (`kdf.DefaultArgon2Params()`).
    - Derive `MasterKey` and subkeys via `kdf.DeriveMasterKey(password, params)` and `kdf.DeriveSubKeys(masterKey)`.
    - Wrap `VaultKey` with `K_wrap` producing `keys.WrappedKeyEnvelope`.
    - Persist `sqlite.VaultMetadata` via `MetadataRepository.Create(ctx, meta)`.
    - Automatically open session or return initialized status.
- **R2 → Unlock & Active Session State:**
  - Implement `Session` struct holding `vaultID`, `vaultKey` ([]byte), `isUnlocked` (bool), protected by `sync.RWMutex`.
  - Implement `VaultService.Unlock(ctx, masterPassword string) (*Session, error)`:
    - Fetch `VaultMetadata` from `MetadataRepository.Get(ctx)`.
    - Derive `MasterKey` using stored salt and KDF params.
    - Compute `MCK` and verify against stored `envelope.MCK` via `kdf.VerifyMCK()`; return `domain.ErrInvalidPassword` on failure.
    - Unwrap `VaultKey` via `keys.UnwrapVaultKey(derivedSubkeys.WrapKey, envelope)`.
    - Store `VaultKey` in active `Session` and mark unlocked.
- **R3 → Lock & Secure Teardown:**
  - Implement `Session.Lock()` and `VaultService.Lock()`:
    - Acquire write lock on `Session`.
    - Zeroize in-memory `VaultKey` bytes via `kdf.Zeroize(session.vaultKey)`.
    - Reset `session.isUnlocked = false` and nil key slice.
- **R4 → Master Password Rotation:**
  - Implement `VaultService.ChangeMasterPassword(ctx, oldPassword, newPassword string) error`:
    - Fetch `VaultMetadata`.
    - Rotate master password via `keys.RotateMasterPassword(oldPassword, newPassword, meta.Envelope, newParams)`.
    - Atomically update metadata via `MetadataRepository.UpdateEnvelope(ctx, meta.VaultID, rotatedEnvelope)`.
    - If active session exists, update internal keys or require re-unlock.
- **R5 → Vault Purge:**
  - Implement `VaultService.Purge(ctx) error`:
    - Lock active session if unlocked.
    - Close DB connection and remove database files (including WAL/SHM).

## Architecture

```
internal/service/
├── session.go         # In-memory thread-safe Session manager with key zeroization
├── vault_service.go   # Vault lifecycle use-case orchestration
└── errors.go          # Service-level errors (e.g. ErrVaultLocked, ErrInvalidPassword)
```

- Inward dependency flow: `internal/service` orchestrates domain models, crypto packages, and storage repositories.

## Components Affected

- `internal/service/errors.go`
- `internal/service/session.go`
- `internal/service/vault_service.go`
- `internal/service/vault_service_test.go`

## Data Changes

- None (uses existing `vault_metadata` table).

## API Changes

```go
package service

type Session struct {
    sync.RWMutex
    // unexported fields: vaultID, vaultKey, unlocked
}

func (s *Session) IsUnlocked() bool
func (s *Session) VaultID() (string, error)
func (s *Session) VaultKey() ([]byte, error)
func (s *Session) Lock()

type VaultService struct {
    db       *sqlite.DB
    metaRepo *sqlite.MetadataRepository
    session  *Session
    dbPath   string
}

func NewVaultService(db *sqlite.DB, metaRepo *sqlite.MetadataRepository, dbPath string) *VaultService
func (s *VaultService) Init(ctx context.Context, masterPassword string) error
func (s *VaultService) Unlock(ctx context.Context, masterPassword string) (*Session, error)
func (s *VaultService) Lock() error
func (s *VaultService) IsUnlocked() bool
func (s *VaultService) Session() *Session
func (s *VaultService) ChangeMasterPassword(ctx context.Context, oldPassword, newPassword string) error
func (s *VaultService) Inspect(ctx context.Context) (*sqlite.VaultHeaderInfo, error)
func (s *VaultService) Purge(ctx context.Context) error
```

## Integration Changes

- Ingested by CLI commands (`govault init`, `unlock`, `lock`, `status`, `passwd`, `purge`) and TUI state machines.

## Implementation Sequence

1. Define service errors (`ErrVaultLocked`, `ErrInvalidPassword`, `ErrAlreadyUnlocked`, etc.).
2. Implement thread-safe `Session` with mutex protection, status queries, and zeroization on `Lock()`.
3. Implement `VaultService` methods: `Init`, `Unlock`, `Lock`, `ChangeMasterPassword`, `Inspect`, `Purge`.
4. Write comprehensive unit and lifecycle integration tests covering init, unlock, invalid password rejection, password rotation, lock zeroization, and purge.

## Test Strategy

- **Init & Unlock Flow:** Initialize new vault, unlock with password, verify `Session.VaultKey()` is valid 32-byte key.
- **Wrong Password Rejection:** Attempt unlock with wrong password, assert `ErrInvalidPassword` and session remains locked.
- **Lock & Zeroize Test:** Unlock session, call `Lock()`, assert session is locked and access to `VaultKey()` returns `ErrVaultLocked`.
- **Master Password Change Test:** Change password from "pass1" to "pass2", verify unlock with "pass1" fails, and unlock with "pass2" succeeds with identical underlying decrypted secrets.
- **Purge Test:** Initialize and purge vault, verify database files are deleted.

## Risks

- File locking on Windows during Purge; mitigated by closing DB handle before deleting file.

## Assumptions

- Single active session per application instance.
