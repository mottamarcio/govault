---
type: plan
for: SPEC-011
status: ready
---

# Implementation Plan: Portable Backup File Format & Envelope Encoding (.gvault)

## Summary

Design and implement the low-level, self-contained `.gvault` binary backup container under `internal/backup/format`. This includes the fixed binary header parsing and serialization (8-byte magic `GOVAULTB`, versioning, KDF parameters, wrapped key envelope), canonical CBOR/binary logical payload data model (`BackupPayloadV1`), `BackupKey` derivation (HKDF-SHA256 from `VaultKey`), XChaCha20-Poly1305 payload encryption with `BackupAAD` binding, and non-secret structural header inspection without password inputs.

## Repository Context

- Cryptography: `internal/crypto/cipher` (XChaCha20-Poly1305), `internal/crypto/kdf` (Argon2id, HKDF-SHA256), `internal/crypto/keys` (Key wrapping and unwrapping).
- Domain: `internal/domain` (`Record`, `Tag`, `HistoryEntry`, payloads and serialization).
- Storage: Independent from SQLite storage layer; represents logical vault contents.
- Deterministic encoding: Big-endian fixed layout for unencrypted header and deterministic JSON/CBOR encoding for payload structures.

## Requirement Coverage

- **R1 → Backup File Magic & Header Binary Layout:**
  - Define constants in `internal/backup/format/header.go`:
    - `Magic = "GOVAULTB"` (8 bytes: `0x47, 0x4F, 0x56, 0x41, 0x55, 0x4C, 0x54, 0x42`).
    - `BackupFormatVersion = 1`, `CryptoSuiteVersion = 1`.
    - `KDFAlgArgon2id = 1`.
  - Define `Header` struct with fields: `BackupID [16]byte`, `VaultID string (16-byte UUID / string)`, `CreatedAtUnix int64`, `KDFParams *kdf.Argon2Params`, `WrappedVaultKey []byte`, `MCK []byte`, `BackupNonce [24]byte`, `CiphertextLength uint64`.
  - Implement `Header.MarshalBinary() ([]byte, error)` and `ParseHeader(r io.Reader) (*Header, int, error)` with defensive size limit checks.
- **R2 → Logical Backup Payload Data Model & Serialization:**
  - Define structures in `internal/backup/format/payload.go`:
    - `BackupManifest`: `SourceVaultVersion uint16`, `SourceSchemaVersion uint16`, `AppVersion string`, `EntryCount int`, `HistoryCount int`, `TagCount int`, `CreatedAtUnix int64`.
    - `BackupEntry`: `ID string`, `Type domain.RecordType`, `Title string`, `Payload []byte`, `Version int`, `CreatedAtUnix int64`, `UpdatedAtUnix int64`, `Tags []string`, `IsDeleted bool`.
    - `BackupHistory`: `HistoryID string`, `RecordID string`, `Version int`, `Payload []byte`, `CreatedAtUnix int64`.
    - `BackupTag`: `ID string`, `Name string`, `CreatedAtUnix int64`.
    - `BackupPayload`: `Version uint16`, `Manifest BackupManifest`, `Entries []BackupEntry`, `Histories []BackupHistory`, `Tags []BackupTag`.
  - Implement `BackupPayload.Marshal()` and `UnmarshalPayload(data []byte) (*BackupPayload, error)` with deterministic sorting (entries by ID, histories by RecordID+Version, tags by Name).
- **R3 → Cryptographic Backup Envelope & AAD Binding:**
  - Implement `DeriveBackupKey(vaultKey []byte, vaultID string) ([]byte, error)` using `kdf.DeriveSubkey(vaultKey, []byte(vaultID), "govault:v1:backup")`.
  - Implement `BackupAAD(header *Header) []byte`:
    - Canonical binding format: `govault/v1/backup|format:1|suite:1|bid:<backup_id>|vid:<vault_id>|created:<created_at_unix>`.
  - Implement `EncryptBackup(w io.Writer, payload *BackupPayload, vaultKey []byte, headerOpts HeaderOptions) error`:
    - Marshal payload to bytes.
    - Derive `BackupKey` and generate random 24-byte `backup_nonce`.
    - Construct `BackupAAD` from header metadata.
    - Encrypt payload bytes using `cipher.EncryptWithNonce(backupKey, backupNonce, payloadBytes, aad)`.
    - Write serialized Header followed by ciphertext stream to writer.
  - Implement `DecryptBackup(r io.Reader, vaultKey []byte) (*Header, *BackupPayload, error)`:
    - Parse Header and authenticate ciphertext using derived `BackupKey` and `BackupAAD`.
- **R4 → Non-Secret Structural Inspection & Parsing:**
  - Implement `InspectHeader(r io.Reader) (*HeaderInspection, error)`:
    - Reads and verifies magic, header length, versions, KDF parameter boundaries (memory >= 64MB, iterations >= 1, threads >= 1), wrapped key length, and ciphertext bounds without attempting any key derivation or decryption.
    - Returns non-secret inspection info (`BackupID`, `VaultID`, `CreatedAt`, `FormatVersion`, `CryptoSuite`, `KDFMemory`, `KDFIterations`, `FileSize`).

## Architecture

```
internal/backup/format/
├── header.go          # Binary header encoding/decoding, constants, validation
├── payload.go         # Logical BackupPayload structs and canonical serialization
├── crypto.go          # BackupKey derivation, BackupAAD construction, encryption/decryption
├── inspect.go         # Non-secret structural inspection
├── errors.go          # Domain format errors (ErrInvalidMagic, ErrUnsupportedVersion, etc.)
├── header_test.go     # Header serialization and boundary tests
├── payload_test.go    # Payload marshaling and ordering tests
└── crypto_test.go     # End-to-end backup envelope encryption and tamper resistance tests
```

## Components Affected

- `internal/backup/format/header.go`
- `internal/backup/format/payload.go`
- `internal/backup/format/crypto.go`
- `internal/backup/format/inspect.go`
- `internal/backup/format/errors.go`
- Test files under `internal/backup/format/`

## Data Changes

- Introduction of the `.gvault` file container format with binary header and encrypted payload stream.

## API Changes

```go
package format

type Header struct {
    FormatVersion      uint16
    CryptoSuiteVersion uint16
    BackupID           [16]byte
    VaultID            string
    CreatedAtUnix      int64
    KDFParams          *kdf.Argon2Params
    WrappedVaultKey    []byte
    MCK                []byte
    BackupNonce        [24]byte
    CiphertextLength   uint64
}

type HeaderInspection struct {
    BackupID      string
    VaultID       string
    CreatedAt     time.Time
    FormatVersion uint16
    CryptoSuite   uint16
    KDFMemoryMB   uint32
    KDFIterations uint32
    KDFThreads    uint8
    PayloadSize   uint64
}

type BackupPayload struct {
    Version   uint16          `json:"version"`
    Manifest  BackupManifest  `json:"manifest"`
    Entries   []BackupEntry   `json:"entries"`
    Histories []BackupHistory `json:"histories"`
    Tags      []BackupTag     `json:"tags"`
}

func (h *Header) MarshalBinary() ([]byte, error)
func ParseHeader(r io.Reader) (*Header, error)
func InspectHeader(r io.Reader) (*HeaderInspection, error)
func DeriveBackupKey(vaultKey []byte, vaultID string) ([]byte, error)
func EncryptBackup(w io.Writer, payload *BackupPayload, vaultKey []byte, header *Header) error
func DecryptBackup(r io.Reader, vaultKey []byte) (*Header, *BackupPayload, error)
```

## Integration Changes

- Consumed by `internal/service/backup_service.go` in `SPEC-012` and CLI backup commands in `FEAT-005`.

## Implementation Sequence

1. Implement errors and header binary serialization/parsing in `header.go` and `errors.go`.
2. Implement logical payload structs and canonical serialization in `payload.go`.
3. Implement `BackupKey` derivation, `BackupAAD`, `EncryptBackup`, and `DecryptBackup` in `crypto.go`.
4. Implement `InspectHeader` in `inspect.go`.
5. Implement comprehensive unit tests:
   - Header round-trip and corrupted header validation.
   - Payload serialization and deterministic ordering.
   - End-to-end backup encrypt/decrypt round-trip with correct/incorrect keys.
   - Header tampering detection via AAD mismatch.

## Test Strategy

- **Header Binary Tests:** Table-driven tests validating magic bytes, big-endian integer fields, and rejection of invalid versions or truncated headers.
- **Payload Round-Trip Tests:** Assert complete preservation of Unicode, multiline payloads, tags, and histories.
- **Cryptographic Envelope Tests:** Test encryption/decryption with valid `VaultKey`, verify failure with incorrect keys or corrupted ciphertext.
- **Tamper Resistance Tests:** Deliberately mutate individual header fields (`vault_id`, `created_at`, `backup_id`) and assert that decryption fails due to AAD authentication failure.
- **Structural Inspection Tests:** Assert that `InspectHeader` extracts correct metadata without requiring a password or secret key.

## Risks

- Integer overflow or unbounded allocation on reading corrupted `ciphertext_length` or `header_length`; mitigated by enforcing strict upper limits (MaxHeaderSize = 64KB, MaxCiphertextSize = 8GB) before allocating buffers.

## Assumptions

- Master key derivation and key wrapping implementations in `internal/crypto` are robust and reusable for backup envelopes.
