---
id: SPEC-011
type: spec
status: ready
parent: FEAT-004
depends_on:
  - SPEC-001
  - SPEC-002
  - SPEC-003
  - SPEC-007
supersedes: []
---

# SPEC-011: Portable Backup File Format & Envelope Encoding (.gvault)

## Intent

Specify the canonical binary backup file format (`.gvault`), fixed header structure, KDF and wrapped vault key metadata serialization, XChaCha20-Poly1305 encrypted logical payload stream, Backup AAD construction, and structural validation parser.

## Requirements

- **R1: Backup File Magic & Header Binary Layout:** Define and implement the canonical binary encoding and parsing for the unencrypted backup header (8-byte magic `GOVAULTB`, 2-byte backup format version `1`, 2-byte crypto suite version `1`, 4-byte header length, 16-byte `backup_id`, 16-byte `vault_id`, 8-byte `created_at_unix`, KDF Argon2id metadata, 24-byte `key_wrap_nonce`, wrapped vault key, 24-byte `backup_nonce`, and 8-byte `ciphertext_length`) using big-endian integer encoding.
- **R2: Logical Backup Payload Data Model & Serialization:** Define and serialize the logical backup payload container (`BackupPayloadV1`) containing `BackupManifestV1` (entry count, history count, tag count, created timestamp, format versions), `BackupEntryV1` list (entry ID, record kind, version, created/updated timestamps, raw serialized domain payload), `BackupHistoryV1` list (history ID, entry ID, created timestamp, payload), and `BackupTagV1` list in deterministic sorted order using canonical CBOR/binary encoding.
- **R3: Cryptographic Backup Envelope & AAD Binding:** Encrypt and decrypt the logical payload using `BackupKey` (derived via HKDF-SHA256 from `VaultKey` with `vault_id` salt and info `govault:v1:backup`) and XChaCha20-Poly1305, binding the header fields as Associated Authenticated Data (`BackupAAD`) to guarantee complete header integrity and tamper resistance.
- **R4: Non-Secret Structural Inspection & Parsing:** Provide structural verification and non-secret header inspection without requiring master password input (verifying magic bytes, format/crypto suite versions, KDF parameter boundaries, wrapped key length, ciphertext bounds, and file truncation).

## Acceptance Scenarios

- **Scenario 1: Header Serialization and Structural Inspection**
  - *Given* valid backup header metadata and wrapped key parameters
  - *When* serialized to binary bytes and inspected structurally
  - *Then* magic bytes `GOVAULTB`, version 1, timestamps, KDF parameters, and IDs parse correctly without requiring password decryption.
- **Scenario 2: Full Backup Envelope Round-Trip**
  - *Given* an unlocked vault with logical records, tags, and revision histories
  - *When* serialized into `BackupPayloadV1`, encrypted with derived `BackupKey` and `BackupAAD`, and written to `.gvault` format
  - *Then* opening with the correct master password verifies the AAD, unwraps the Vault Key, decrypts the payload, and restores identical manifest counts and logical records.
- **Scenario 3: Tampered Header Rejection**
  - *Given* an encrypted `.gvault` file where a byte in the unencrypted header (e.g. `vault_id` or `created_at_unix`) is modified
  - *When* decryption is attempted with the correct master password
  - *Then* payload authentication fails due to Backup AAD mismatch.

## Edge Cases

- Backup file smaller than minimum header size or truncated mid-ciphertext fails immediately during structural check.
- Corrupted or invalid KDF parameters (zero iterations, memory < 64MB) are rejected defensively during parsing.
- Empty vault with zero entries still creates a valid `.gvault` backup with entry_count = 0 in manifest.

## Constraints

- Backup format must be completely self-contained and independent from SQLite database layout or files.
- Binary encoding must use deterministic big-endian integers and canonical CBOR/binary encoding.
- Zero plaintext secrets in backup headers or unencrypted streams.

## Non-Goals

- Database restoration execution or SQLite table writes (handled in SPEC-012).
