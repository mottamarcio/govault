---
type: validation
for: SPEC-011
result: pass
---

# Validation: Portable Backup File Format & Envelope Encoding (.gvault)

## Summary

All 4 requirements in SPEC-011 have been implemented, tested, and verified against the real codebase in `internal/backup/format`. The test suite passed with 100% success rate under race detection, confirming the canonical big-endian binary header layout, the logical backup payload model with deterministic sorting, `BackupKey` derivation (HKDF-SHA256 from `VaultKey`), XChaCha20-Poly1305 encryption with `BackupAAD` tamper protection, and password-free non-secret structural header inspection.

## Requirement Validation

### R1: Backup File Magic & Header Binary Layout
- **Plan coverage:** Mapped in `plan.md` under R1 (`Header` struct, big-endian binary layout, `GOVAULTB` magic bytes, format version `1`, crypto suite `1`, KDF Argon2id metadata, wrapped vault key, MCK, nonces, and bounded length checks).
- **Task coverage:** Covered and completed in `TASK-037`.
- **Code evidence:** Implemented in `internal/backup/format/header.go` (methods `MarshalBinary`, `ParseHeader`).
- **Test evidence:** `TestHeaderMarshalParseRoundTrip` and `TestHeaderValidationErrors` in `internal/backup/format/header_test.go` verified exact field round-tripping and defensive rejection of invalid magics, unsupported versions, and truncated headers.
- **Result:** pass

### R2: Logical Backup Payload Data Model & Serialization
- **Plan coverage:** Mapped in `plan.md` under R2 (`BackupPayload`, `BackupManifest`, `BackupEntry`, `BackupHistory`, `BackupTag`, canonical JSON serialization, deterministic sorting, and structural integrity checks).
- **Task coverage:** Covered and completed in `TASK-038`.
- **Code evidence:** Implemented in `internal/backup/format/payload.go` (methods `Marshal`, `UnmarshalPayload`, `SortNormalizes`).
- **Test evidence:** `TestPayloadSerializationRoundTrip` and `TestPayloadValidationErrors` in `internal/backup/format/payload_test.go` verified deterministic sorting, full preservation of typed records, and rejection of duplicate IDs and orphaned history records.
- **Result:** pass

### R3: Cryptographic Backup Envelope & AAD Binding
- **Plan coverage:** Mapped in `plan.md` under R3 (`DeriveBackupKey` with `vault_id` salt and info `govault:v1:backup`, `ConstructBackupAAD`, `EncryptBackup`, and `DecryptBackup` with XChaCha20-Poly1305).
- **Task coverage:** Covered and completed in `TASK-039`.
- **Code evidence:** Implemented in `internal/backup/format/crypto.go` (functions `DeriveBackupKey`, `ConstructBackupAAD`, `EncryptBackup`, `DecryptBackup`).
- **Test evidence:** `TestBackupCryptoRoundTrip` in `internal/backup/format/crypto_test.go` verified round-trip encryption/decryption with valid `VaultKey`, failure with incorrect key, and immediate rejection of tampered header fields due to `BackupAAD` mismatch.
- **Result:** pass

### R4: Non-Secret Structural Inspection & Parsing
- **Plan coverage:** Mapped in `plan.md` under R4 (`InspectHeader` parsing header metadata without secret key or password input).
- **Task coverage:** Covered and completed in `TASK-037`.
- **Code evidence:** Implemented in `internal/backup/format/inspect.go` (function `InspectHeader`).
- **Test evidence:** `TestInspectHeader` in `internal/backup/format/header_test.go` verified extraction of `BackupID`, `VaultID`, timestamps, format versions, KDF parameters, and ciphertext sizes without password prompts or decryption.
- **Result:** pass

## Unplanned Implementation

- None.

## Findings

- Clear separation of low-level `.gvault` binary container concerns (`internal/backup/format`) from SQLite storage or service orchestrations (`SPEC-012`).
- Strict enforcement of bounded allocations (`MaxHeaderSize = 64KB`, `MaxCiphertext = 8GB`) protecting against resource exhaustion.
- Cryptographic binding between header and ciphertext via `BackupAAD` prevents unauthorized metadata modification.

## Recommended Corrections

- None.
