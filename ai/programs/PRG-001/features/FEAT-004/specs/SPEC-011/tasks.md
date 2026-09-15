---
type: tasks
for: SPEC-011
---

# Tasks

## TASK-037 — Backup File Binary Header Layout & Structural Inspection

- [x] Completed
- **Serves:** SPEC-011:R1, SPEC-011:R4
- **Depends on:** none
- **Files/Components:** `internal/backup/format/header.go`, `internal/backup/format/inspect.go`, `internal/backup/format/errors.go`, `internal/backup/format/header_test.go`
- **Verification:** Ran `go test -v -race ./internal/backup/format -run 'TestHeader|TestInspect'` verifying big-endian binary header serialization, round-trip parsing, magic byte validation (`GOVAULTB`), version checking, KDF boundary limits, and password-free structural inspection. Evidence: `PASS TestHeaderMarshalParseRoundTrip`, `PASS TestInspectHeader`, `PASS TestHeaderValidationErrors`.

## TASK-038 — Logical Backup Payload Data Model & Canonical Serialization

- [x] Completed
- **Serves:** SPEC-011:R2
- **Depends on:** none
- **Files/Components:** `internal/backup/format/payload.go`, `internal/backup/format/payload_test.go`
- **Verification:** Ran `go test -v -race ./internal/backup/format -run 'TestPayload'` verifying `BackupPayload`, `BackupManifest`, `BackupEntry`, `BackupHistory`, and `BackupTag` deterministic sorting and round-trip marshaling without data loss. Evidence: `PASS TestPayloadSerializationRoundTrip`, `PASS TestPayloadValidationErrors`.

## TASK-039 — Cryptographic Backup Envelope, Key Derivation & AAD Authentication

- [x] Completed
- **Serves:** SPEC-011:R3
- **Depends on:** TASK-037, TASK-038
- **Files/Components:** `internal/backup/format/crypto.go`, `internal/backup/format/crypto_test.go`
- **Verification:** Ran `go test -v -race ./internal/backup/format -run 'TestBackupCrypto'` verifying `BackupKey` derivation via HKDF-SHA256, full envelope encryption with XChaCha20-Poly1305, successful decryption with valid key, and strict rejection of tampered headers via Backup AAD verification. Evidence: `PASS TestBackupCryptoRoundTrip`.
