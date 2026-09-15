---
type: plan
for: SPEC-009
status: ready
---

# Implementation Plan: Record Application Service & In-Memory Search

## Summary

Design and implement `RecordService` under `internal/service`, orchestrating encrypted secret creation, retrieval, decryption, modification with history archiving, soft deletion, restore, permanent purging, tag management, and in-memory search and fuzzy filtering across decrypted record attributes.

## Repository Context

- Cryptography: `internal/crypto/cipher` (XChaCha20-Poly1305, AADContext), `internal/crypto/kdf` (`DeriveSubKey`), `internal/crypto/keys`.
- Storage: `internal/storage/sqlite` (`RecordRepository`, `TagRepository`, `HistoryRepository`).
- Domain: `internal/domain` (Entities, value objects, typed payloads, serialization, validation).
- Service Session: `internal/service/session.go` (`Session`).

## Requirement Coverage

- **R1 → Record Encryption & Creation:**
  - Implement `RecordService.Create(ctx, input RecordInput) (*domain.Record, error)`:
    - Validate session is unlocked (`session.VaultKey()`, `session.VaultID()`).
    - Validate input payload and metadata.
    - Serialize typed payload to JSON via `domain.SerializePayload()`.
    - Derive record encryption subkey `K_record` via HKDF: `kdf.DeriveSubKey(vaultKey, fmt.Sprintf("govault/v1/record/%s", recordID), 32)`.
    - Construct authenticated `cipher.AADContext{VaultID, RecordID, RecordType, Version: 1}`.
    - Encrypt payload via `cipher.Encrypt(kRecord, plaintext, aadContext.Bytes())`.
    - Resolve tag IDs from names (creating new tags if necessary via `TagRepository`).
    - Save encrypted record and tag associations via `RecordRepository.Create(ctx, record, tagIDs)`.
- **R2 → Record Retrieval & Decryption:**
  - Implement `RecordService.GetByID(ctx, id string) (*domain.Record, error)`:
    - Fetch encrypted `sqlite.Record` from `RecordRepository.GetByID(ctx, vaultID, id)`.
    - Derive record subkey `K_record`.
    - Decrypt payload via `cipher.Decrypt(kRecord, record.Payload, aadContext.Bytes())`.
    - Deserialize into typed domain payload via `domain.DeserializePayload()`.
    - Fetch tag names via `TagRepository.GetRecordTags(ctx, id)` and attach to `domain.Record`.
  - Implement `RecordService.List(ctx, includeDeleted bool) ([]*domain.Record, error)` (fetching active/trash records and decrypting in-memory).
- **R3 → Record Modification & History Inspection:**
  - Implement `RecordService.Update(ctx, id string, input RecordInput) (*domain.Record, error)`:
    - Fetch existing record; verify version match.
    - Increment version (`current.Version + 1`).
    - Encrypt new payload with new AADContext (version `current.Version + 1`).
    - Call `RecordRepository.Update(ctx, updatedRecord, newTagIDs)`.
  - Implement `RecordService.ListHistory(ctx, recordID string) ([]*domain.HistoryEntry, error)` and `GetHistoryRevision(ctx, recordID string, version uint32) (*domain.Record, error)`.
- **R4 → Soft Deletion, Restore & Trash Management:**
  - Implement `RecordService.Delete(ctx, id string) error` (`RecordRepository.SoftDelete`).
  - Implement `RecordService.Restore(ctx, id string) error` (`RecordRepository.Restore`).
  - Implement `RecordService.ListTrash(ctx) ([]*domain.Record, error)`.
  - Implement `RecordService.PurgeTrash(ctx) error` (`RecordRepository.PurgeDeleted`).
- **R5 → In-Memory Search & Filtering:**
  - Implement `RecordService.Search(ctx, filter SearchFilter) ([]*domain.Record, error)`:
    - Retrieve and decrypt all active records into memory.
    - Filter records by:
      - `Query`: case-insensitive substring match against `Title`, `Username`, `URI`, `Service`, `Content`, `Notes`, or custom field keys/values.
      - `Type`: record type filter (if specified).
      - `Tags`: tag list inclusion (if specified).
    - Sort results by relevance (exact title match > prefix match > content match).

## Architecture

```
internal/service/
├── record_service.go        # Record CRUD, history, and trash use cases
├── search.go                # In-memory search, filtering, and scoring logic
├── search_test.go           # In-memory search tests
└── record_service_test.go   # End-to-end service tests
```

## Components Affected

- `internal/service/record_service.go`
- `internal/service/search.go`
- `internal/service/record_service_test.go`
- `internal/service/search_test.go`

## Data Changes

- None.

## API Changes

```go
package service

type RecordInput struct {
    ID      string
    Title   string
    Type    domain.RecordType
    Tags    []string
    Payload any
    Version uint32
}

type SearchFilter struct {
    Query      string
    Type       domain.RecordType
    Tags       []string
    IncludeTrash bool
}

type RecordService struct {
    session     *Session
    recordRepo  *sqlite.RecordRepository
    tagRepo     *sqlite.TagRepository
    historyRepo *sqlite.HistoryRepository
}

func NewRecordService(
    session *Session,
    recordRepo *sqlite.RecordRepository,
    tagRepo *sqlite.TagRepository,
    historyRepo *sqlite.HistoryRepository,
) *RecordService

func (s *RecordService) Create(ctx context.Context, input RecordInput) (*domain.Record, error)
func (s *RecordService) GetByID(ctx context.Context, id string) (*domain.Record, error)
func (s *RecordService) List(ctx context.Context, includeDeleted bool) ([]*domain.Record, error)
func (s *RecordService) Update(ctx context.Context, id string, input RecordInput) (*domain.Record, error)
func (s *RecordService) Delete(ctx context.Context, id string) error
func (s *RecordService) Restore(ctx context.Context, id string) error
func (s *RecordService) ListTrash(ctx context.Context) ([]*domain.Record, error)
func (s *RecordService) PurgeTrash(ctx context.Context) error
func (s *RecordService) ListHistory(ctx context.Context, recordID string) ([]*domain.HistoryEntry, error)
func (s *RecordService) GetHistoryRevision(ctx context.Context, recordID string, version uint32) (*domain.Record, error)
func (s *RecordService) Search(ctx context.Context, filter SearchFilter) ([]*domain.Record, error)
```

## Integration Changes

- Ingested by CLI commands (`govault add`, `get`, `list`, `edit`, `rm`, `restore`, `trash`, `history`, `search`) and Bubble Tea interactive TUI models.

## Implementation Sequence

1. Implement in-memory search helper (`search.go`).
2. Implement `RecordService` with encrypted CRUD, tag mapping, subkey derivation, and AAD construction (`record_service.go`).
3. Implement soft-deletion, restore, purge trash, and history inspection methods.
4. Implement comprehensive unit and integration tests (`record_service_test.go`, `search_test.go`).

## Test Strategy

- **Creation & Decryption Test:** Create records of each type (Login, Note, APIKey, Custom), retrieve and decrypt, verify exact field matches.
- **Subkey & AAD Isolation Test:** Verify that tampering with record ID, version, or ciphertext fails decryption with authentication errors.
- **Revision History Test:** Update record multiple times, list history, fetch and decrypt historical revisions.
- **Trash & Purge Test:** Soft delete, verify excluded from normal search, list in trash, restore, delete again and purge.
- **In-Memory Search Test:** Search by title, tag, partial username, notes, and custom field values.

## Risks

- Performance on large vaults; mitigated by keeping SQLite I/O sequential and fast in-memory filtering.

## Assumptions

- Vault is unlocked and session has active `VaultKey` before calling `RecordService` methods.
