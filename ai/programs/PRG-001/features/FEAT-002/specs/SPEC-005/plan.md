---
type: plan
for: SPEC-005
status: ready
---

# Implementation Plan: Vault Metadata Repository & Lifecycle Persistence

## Summary

Design and implement the `MetadataRepository` in `internal/storage/sqlite` responsible for persisting, querying, and updating the single-vault metadata entry in SQLite (`vault_metadata`), supporting inspection without password unwrapping and atomic metadata updates during master password rotation.

## Repository Context

`SPEC-004` (SQLite database manager, PRAGMAs, and Schema Migration v1) is implemented and validated in `internal/storage/sqlite`. Cryptographic types (`WrappedKeyEnvelope`, `Argon2Params`) are available in `internal/crypto/keys` and `internal/crypto/kdf`. This implementation adds the metadata repository interacting directly with the `vault_metadata` table created by Migration v1.

## Requirement Coverage

- **R1 → Metadata Record Schema:**
  - Define `VaultMetadata` entity struct:
    - `VaultID` (string)
    - `SchemaVersion` (int)
    - `CryptoSuite` (string = "v1")
    - `Envelope` (*keys.WrappedKeyEnvelope)
    - `CreatedAt` (time.Time)
    - `UpdatedAt` (time.Time)
  - Implement SQL mapping serializing `Envelope.KDFParams` to JSON, `Envelope.WrappedKey` as BLOB, `Envelope.MCK` as BLOB, and timestamps as Unix integers.
- **R2 → Single-Vault Constraint:**
  - In `Create(ctx context.Context, meta *VaultMetadata) error`:
    - Enforce via transaction that `vault_metadata` is currently empty before insertion.
    - If a row already exists, return `ErrVaultAlreadyExists`.
- **R3 → Atomic Metadata Updates on Rotation:**
  - In `UpdateEnvelope(ctx context.Context, vaultID string, env *keys.WrappedKeyEnvelope) error`:
    - Update `kdf_params_json`, `wrapped_key`, `mck`, and `updated_at` in a single SQL UPDATE.
    - Ensure update affects exactly 1 row matching `vaultID`.
- **R4 → Metadata Inspection:**
  - Implement `Get(ctx context.Context) (*VaultMetadata, error)` returning the full metadata record (or `ErrVaultNotInitialized` if no rows exist).
  - Implement `Inspect(ctx context.Context) (*VaultHeaderInfo, error)` returning non-secret summary (`VaultID`, `SchemaVersion`, `CryptoSuite`, `CreatedAt`, `UpdatedAt`) without unpacking sensitive envelope components.

## Architecture

- Repository layer in `internal/storage/sqlite`.
- Consumes `internal/crypto/keys` for envelope structuring.
- Exposes clear error sentinel values: `ErrVaultNotInitialized`, `ErrVaultAlreadyExists`.

## Components Affected

- `internal/storage/sqlite/metadata_repo.go`: `MetadataRepository` implementation and entity types.
- `internal/storage/sqlite/metadata_repo_test.go`: Unit and integration tests for CRUD, single-vault enforcement, inspection, and atomic rotation updates.

## Data Changes

- Directly interacts with `vault_metadata` table.

## API Changes

```go
package sqlite

type VaultMetadata struct {
    VaultID       string
    SchemaVersion int
    CryptoSuite   string
    Envelope      *keys.WrappedKeyEnvelope
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type VaultHeaderInfo struct {
    VaultID       string
    SchemaVersion int
    CryptoSuite   string
    CreatedAt     time.Time
    UpdatedAt     time.Time
}

type MetadataRepository struct {
    db *DB
}

func NewMetadataRepository(db *DB) *MetadataRepository
func (r *MetadataRepository) Create(ctx context.Context, meta *VaultMetadata) error
func (r *MetadataRepository) Get(ctx context.Context) (*VaultMetadata, error)
func (r *MetadataRepository) Inspect(ctx context.Context) (*VaultHeaderInfo, error)
func (r *MetadataRepository) UpdateEnvelope(ctx context.Context, vaultID string, env *keys.WrappedKeyEnvelope) error
```

## Integration Changes

- Ingested by `internal/service/vault_service.go` in `FEAT-003`.

## Implementation Sequence

1. Implement `metadata_repo.go` with domain structs, errors, and repository methods.
2. Implement `metadata_repo_test.go` verifying create, get, inspect, single-vault constraint, and rotation envelope update.

## Test Strategy

- **Creation & Retrieval Test:** Create metadata; query back via `Get`; assert fields and envelope match.
- **Single-Vault Invariant Test:** Attempt to insert a second metadata row; assert `ErrVaultAlreadyExists`.
- **Inspection Test:** Call `Inspect`; assert correct timestamps and ID returned.
- **Envelope Update Test:** Execute `UpdateEnvelope`; verify updated timestamp and new envelope in database.

## Risks

- None.

## Assumptions

- Database has Migration v1 applied prior to repository calls.
