---
id: KNOW-004
type: knowledge
status: active
sources:
  - path: ai/raw/vault-format.md
    fingerprint: sha256:bee7cbfd122eb7c712638cedff9b58e49e64581e76b5645979adf040ded064c4
---

# KNOW-004: Vault Database Storage Format and Schema

## Summary

Specifies the on-disk SQLite representation for GoVault, detailing table structures, encrypted record envelopes, metadata storage, WAL configuration, transactions, and migration strategies.

## Known Facts

- Storage Engine: SQLite (using modern Go driver such as `modernc.org/sqlite` pure Go or `mattn/go-sqlite3`).
- Journal Mode & Pragmas: WAL (Write-Ahead Logging) mode (`PRAGMA journal_mode = WAL`), `PRAGMA synchronous = NORMAL`, `PRAGMA foreign_keys = ON`, `PRAGMA busy_timeout = 5000`.
- Schema Layout:
  - `vault_metadata`: Stores vault UUID, schema version, crypto suite version, KDF parameters (salt, iterations, memory, parallelism), and the Wrapped Vault Key payload.
  - `records`: Primary table for secrets (UUID primary key, created_at, updated_at, deleted_at for soft-delete/tombstone, record_type, version, nonce, ciphertext, auth_tag, and unencrypted blind routing metadata if configured).
  - `record_history`: Previous versions of encrypted records to support revision inspection and rollback.
  - `tags` / `record_tags`: Many-to-many tag relationships.
- Plaintext vs. Encrypted Boundary:
  - Plaintext in SQLite: Record UUID, timestamps, record type identifier, soft-delete flags, schema version numbers.
  - Encrypted in Payload: Title/name, username, password, URLs, notes, custom key-value attributes, TOTP seeds, history contents.
- Search Indexing: An in-memory search index is constructed upon vault unlock; search across secret fields is never persisted unencrypted in SQLite FTS tables.

## Constraints

- Sensitive data must always be authenticated and encrypted before touching SQLite.
- File permissions on `vault.db`, `vault.db-wal`, and `vault.db-shm` must be restricted (`0600`).
- Schema migrations must be atomic within transactions and backwards-compatible with rollback mechanisms.

## Unknowns

- Blind indexing / deterministic HMAC searchable encryption for fast title search without loading full payloads vs. in-memory searching of all decrypted titles upon unlock (specification prefers in-memory search index for v1).

## Conflicts

- None.

## Provenance

- SQLite schema, envelopes, PRAGMA settings, and indexing model derived directly from `ai/raw/vault-format.md`.

## Related Topics

- [KNOW-003](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-003-cryptography.md) — Cryptographic Architecture
- [KNOW-005](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-005-backup-format.md) — Backup Format
- [KNOW-006](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-006-architecture.md) — Software Architecture
