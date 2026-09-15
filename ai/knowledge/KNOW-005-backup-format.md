---
id: KNOW-005
type: knowledge
status: active
sources:
  - path: ai/raw/backup-format.md
    fingerprint: sha256:9a4479e7d17b7c5d6544187ff776146e2b9d17192d5aabe80e0142a9c9812b48
---

# KNOW-005: Portable Backup File Format (.gvault)

## Summary

Defines the portable, self-contained, SQLite-independent encrypted archive format (`.gvault`) for backing up and restoring GoVault data across machines.

## Known Facts

- Independence: The backup format represents logical vault state (manifest and serialized secret records) rather than raw SQLite database files (`vault.db`).
- File Structure:
  - Magic Bytes: 8-byte file signature (e.g. `GVAULT01` or `\x89GVAULT\n`).
  - Header: Unencrypted header specifying format version, backup UUID, created timestamp, KDF parameters (Argon2id salt, time, memory, threads), and Backup Key wrapping envelope.
  - Payload: Encrypted archive stream (e.g. tar/gzip or CBOR/Protobuf stream) containing the logical manifest, records, tags, and histories, encrypted with a derived Backup Key using XChaCha20-Poly1305.
  - AAD Binding: Header fields are included as Associated Authenticated Data in payload encryption to prevent header tampering.
- Backup Options:
  - User can protect the backup using their existing vault master password or specify a distinct one-time backup passphrase during export.
- Restore Workflow:
  - Supports dry-run inspection of backup metadata (created date, number of entries, version) without executing a full database overwrite.
  - Supports full restore (replacing active vault) or merge/import (importing non-conflicting entries).

## Constraints

- Backups must be 100% self-contained and restorable without depending on the original SQLite database or machine-specific state.
- Plaintext secrets must never appear in unencrypted header segments or temporary backup artifacts.
- Atomic creation: backups must be written to temporary files and atomically renamed upon completion to avoid corrupt partial backup files.

## Unknowns

- Backup payload container format choice: CBOR vs canonical JSON vs Protobuf vs tar archive.

## Conflicts

- None.

## Provenance

- File layout, magic headers, encryption structure, and restore logic derived directly from `ai/raw/backup-format.md`.

## Related Topics

- [KNOW-003](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-003-cryptography.md) — Cryptographic Architecture
- [KNOW-004](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-004-vault-format.md) — Vault Storage Format
- [KNOW-006](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-006-architecture.md) — Software Architecture
