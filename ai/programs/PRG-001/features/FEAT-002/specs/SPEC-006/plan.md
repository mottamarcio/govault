---
type: plan
for: SPEC-006
status: ready
---

# Implementation Plan: Encrypted Record, History & Tag Repository

## Summary

Design and implement the SQLite repositories for encrypted secret records (`RecordRepository`), audit history revisions (`HistoryRepository`), and tag categorizations (`TagRepository`) in `internal/storage/sqlite`, ensuring transactional safety, soft deletion/purging, historical archiving on update, and tag association management.

## Repository Context

`SPEC-004` (SQLite DB & Migration v1) and `SPEC-005` (Metadata Repository) are implemented in `internal/storage/sqlite`. The database tables `records`, `record_history`, `tags`, and `record_tags` already exist with foreign key cascades and indexes. This plan implements the repositories managing these entities.

## Requirement Coverage

- **R1 → Encrypted Record CRUD:**
  - Define `Record` entity struct:
    - `ID` (string UUID)
    - `VaultID` (string)
    - `RecordType` (string)
    - `Version` (uint32)
    - `Payload` ([]byte - packed ciphertext envelope)
    - `CreatedAt` (time.Time)
    - `UpdatedAt` (time.Time)
    - `DeletedAt` (*time.Time)
  - Implement `Create(ctx, record, tags)` inserting into `records` and `record_tags` in a single transaction.
  - Implement `GetByID(ctx, vaultID, id)` returning active record.
  - Implement `List(ctx, vaultID, includeDeleted)` returning records list.
- **R2 → Soft Deletion and Purge:**
  - Implement `SoftDelete(ctx, vaultID, id)` updating `deleted_at = unix_now`.
  - Implement `Restore(ctx, vaultID, id)` setting `deleted_at = NULL`.
  - Implement `PurgeDeleted(ctx, vaultID)` executing `DELETE FROM records WHERE vault_id = ? AND deleted_at IS NOT NULL` (cascading deletes to `record_tags` and `record_history`).
- **R3 → Historical Revisions:**
  - Define `HistoryEntry` struct: `ID`, `RecordID`, `VaultID`, `Version`, `Payload`, `ArchivedAt`.
  - In `Update(ctx, record, newTags)`:
    - Fetch current record within transaction; verify `record.Version == current.Version` (optimistic concurrency).
    - Insert current version into `record_history`.
    - Update `records` incrementing version to `current.Version + 1` with new `Payload` and `updated_at`.
    - Replace tag associations in `record_tags`.
  - Implement `ListHistory(ctx, recordID)` and `GetHistoryVersion(ctx, recordID, version)`.
- **R4 → Tag Association:**
  - Define `Tag` struct: `ID`, `VaultID`, `Name`, `CreatedAt`.
  - Implement `TagRepository`:
    - `CreateTag(ctx, vaultID, name)`
    - `ListTags(ctx, vaultID)`
    - `DeleteTag(ctx, vaultID, tagID)`
    - `GetRecordTags(ctx, recordID)`
    - `SetRecordTags(ctx, recordID, tagIDs)`

## Architecture

- Repository layer in `internal/storage/sqlite`.
- Atomic multi-table transactions for all write operations.
- Clean separation of concerns between records, history, and tags.

## Components Affected

- `internal/storage/sqlite/record_repo.go`: Record CRUD, soft-delete, and update with history archiving.
- `internal/storage/sqlite/tag_repo.go`: Tag CRUD and record-tag mapping.
- `internal/storage/sqlite/history_repo.go`: History retrieval methods.
- `internal/storage/sqlite/records_test.go`: Unit and integration tests for CRUD, history, tags, soft delete, and purge.

## Data Changes

- Interacts directly with `records`, `record_history`, `tags`, and `record_tags`.

## API Changes

```go
package sqlite

type Record struct {
    ID         string
    VaultID    string
    RecordType string
    Version    uint32
    Payload    []byte
    CreatedAt  time.Time
    UpdatedAt  time.Time
    DeletedAt  *time.Time
}

type HistoryEntry struct {
    ID         string
    RecordID   string
    VaultID    string
    Version    uint32
    Payload    []byte
    ArchivedAt time.Time
}

type Tag struct {
    ID        string
    VaultID   string
    Name      string
    CreatedAt time.Time
}

type RecordRepository struct {
    db *DB
}

type TagRepository struct {
    db *DB
}
```

## Integration Changes

- Ingested by `internal/service/record_service.go` in `FEAT-003`.

## Implementation Sequence

1. Implement `tag_repo.go` (tags and record_tags).
2. Implement `history_repo.go` (historical version queries).
3. Implement `record_repo.go` (record CRUD, soft delete, restore, purge, and transactional history archiving on update).
4. Implement comprehensive unit tests in `records_test.go`.

## Test Strategy

- **Record CRUD Test:** Insert encrypted record, retrieve by ID, assert fields match.
- **Soft Delete & Purge Test:** Soft-delete record; verify excluded from normal list; verify listed in trash; purge and verify permanently deleted.
- **History Archiving Test:** Update record across 3 revisions; verify `record_history` contains versions 1 and 2, while `records` contains version 3.
- **Tag Associations Test:** Associate tags; query record tags; delete tag and verify cascade.

## Risks

- None.

## Assumptions

- Foreign keys are active (`PRAGMA foreign_keys = ON;`) ensuring cascade deletes work as expected.
