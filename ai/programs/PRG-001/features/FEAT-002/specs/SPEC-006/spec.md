---
id: SPEC-006
type: spec
status: ready
parent: FEAT-002
depends_on:
  - SPEC-004
  - SPEC-005
---

# SPEC-006: Encrypted Record, History & Tag Repository

## Intent

Specify the SQLite repositories for persisting, querying, updating, soft-deleting, and versioning encrypted secrets (`records`), audit history (`record_history`), and categories/tags (`tags`, `record_tags`).

## Requirements

- **R1: Encrypted Record CRUD:** Implement CRUD repository operations on `records` table storing non-secret routing metadata (UUID `id`, `vault_id`, `record_type`, `version`, timestamps, `deleted_at`) as plaintext columns and sensitive data strictly as packed ciphertext blobs (`payload` blob containing nonce, ciphertext, and tag).
- **R2: Soft Deletion and Purge:** Deleting a record must set `deleted_at` timestamp (soft delete) by default. A dedicated `PurgeDeleted` operation must permanently remove tombstoned records and their associated history entries.
- **R3: Historical Revisions:** When updating an existing record, the previous encrypted envelope and version must be archived into `record_history` with an `archived_at` timestamp. Provide methods to list and retrieve historical revisions.
- **R4: Tag Association:** Implement repository operations for adding, removing, and listing tags for records in `tags` and `record_tags` tables with foreign key cascade support.

## Acceptance Scenarios

- **Scenario 1: Encrypted Record Persistence and Retrieval**
  - *Given* an encrypted record payload
  - *When* inserted into `records`
  - *Then* it can be fetched by UUID with identical ciphertext payload and metadata.
- **Scenario 2: History Retention on Update**
  - *Given* an existing record at version 1
  - *When* updated to version 2 with a new encrypted payload
  - *Then* version 1 is preserved in `record_history` and `records` contains version 2.
- **Scenario 3: Soft Delete and Filter**
  - *Given* an active record
  - *When* soft-deleted
  - *Then* standard queries exclude it, while trash/deleted queries list it until purged.

## Edge Cases

- Attempting to update a record with an outdated version (optimistic locking conflict) must return a concurrency error.
- Deleting a tag must cleanly remove `record_tags` associations without deleting the underlying secret records.

## Constraints

- Plaintext secrets (passwords, usernames, secret notes, custom attributes) must never appear in SQLite columns or unencrypted index tables.
- All write operations mutating multiple tables (e.g. record update + history archive + tag associations) must execute within an atomic transaction.

## Non-Goals

- Storing unencrypted full-text search indexes in SQLite FTS tables (in-memory search is used instead).
