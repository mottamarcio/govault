---
id: FEAT-004
type: feature
status: active
parent: PRG-001
---

# FEAT-004: Portable Backup & Restore Engine (.gvault)

## Capability

Provides creation, inspection, verification, and restoration of portable encrypted backup files (`.gvault`) containing the complete logical vault state.

## User Value

Allows users to safely export their encrypted secrets, verify archive integrity, inspect metadata without full decryption, and restore or merge backups onto any machine without SQLite lock-in.

## Scope

- Serialization and deserialization of logical vault manifests, records, tags, and histories.
- Cryptographic backup envelope: unencrypted header with Argon2id KDF metadata and Wrapped Backup Key, authenticated payload encrypted with XChaCha20-Poly1305.
- AAD cryptographic binding between header and encrypted archive stream.
- Support for exporting with current master password or custom one-time export passphrase.
- Dry-run backup inspection (version, item count, creation date) without executing a database write.
- Atomic backup file writing (temp file write followed by atomic rename).
- Full restore and selective merge import workflows.

## Non-Goals

- Cloud / automated network backup streaming.
- Incremental differential network backups.

## Constraints

- Backups must be strictly self-contained and independent from the SQLite database file structure.
- No unencrypted secrets in backup headers or temporary filesystem locations.
- Restrict backup file permissions to `0600`.

## Relevant Knowledge

- [KNOW-003](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-003-cryptography.md) — Cryptographic Architecture and Key Hierarchy
- [KNOW-005](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-005-backup-format.md) — Portable Backup File Format (.gvault)

## Open Questions

- Final backup archive stream container format (CBOR vs JSON vs tar stream).
