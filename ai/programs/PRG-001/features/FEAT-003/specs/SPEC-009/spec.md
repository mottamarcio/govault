---
id: SPEC-009
type: spec
status: ready
parent: FEAT-003
depends_on:
  - SPEC-001
  - SPEC-006
  - SPEC-007
  - SPEC-008
---

# SPEC-009: Record Application Service & In-Memory Search

## Intent

Specify the `RecordService` orchestrating record operations: creating encrypted secrets, retrieving and decrypting secrets, updating records with automatic revision archiving, soft-deleting/restoring/purging records, tag management, and in-memory search and fuzzy filtering across decrypted record attributes.

## Requirements

- **R1: Record Encryption & Creation:** Serialize domain secret payload, derive record subkey (`K_record`) using HKDF and record ID, encrypt ciphertext envelope with authenticated AAD context, and persist through `RecordRepository` within an unlocked session.
- **R2: Record Retrieval & Decryption:** Fetch encrypted record from storage, derive record subkey, decrypt ciphertext envelope, verify AAD context, deserialize into typed domain record, and attach tag names.
- **R3: Record Modification & History Inspection:** Provide `Update` method encrypting new payload, archiving old revision into history, and updating tag associations. Provide `ListHistory` and `GetHistoryRevision` to inspect and decrypt previous revisions.
- **R4: Soft Deletion, Restore & Trash Management:** Provide `Delete` (soft-delete), `Restore`, `ListTrash`, and `PurgeTrash` operations.
- **R5: In-Memory Search & Filtering:** Provide search capability filtering decrypted records in-memory by query string, record type, and tags without exposing plaintext indexes to SQLite.

## Acceptance Scenarios

- **Scenario 1: End-to-End Record Creation and Decryption**
  - *Given* an unlocked vault session
  - *When* a new Login secret is created
  - *Then* it is stored encrypted in SQLite and can be retrieved and decrypted back into the typed Login secret.
- **Scenario 2: In-Memory Search Across Decrypted Records**
  - *Given* multiple encrypted records in the vault
  - *When* `Search(query)` is executed with a matching title or tag
  - *Then* the matching records are returned sorted by relevance without creating database FTS indexes.
- **Scenario 3: Revision History Traversal**
  - *Given* a record modified three times
  - *When* `ListHistory` and `GetHistoryRevision` are called
  - *Then* previous revision payloads are retrievable and decryptable.

## Edge Cases

- Searching an empty vault or query matching zero records returns an empty slice without error.
- Decrypting a record whose ciphertext was corrupted or tampered with returns `ErrDecryptionFailed`.

## Constraints

- Decryption must only occur in-memory when requested within an unlocked session.
- Never write decrypted title, username, or secret fields to SQLite plaintext columns.

## Non-Goals

- Terminal presentation, tables, or interactive UI rendering.
