---
type: plan
for: SPEC-012
status: ready
---

# Implementation Plan: Backup Creation, Restore & Merge Application Service

## Summary

Design and implement `BackupService` under `internal/service`. `BackupService` orchestrates high-level logical vault export (`.gvault`), custom passphrase re-wrapping, structural and cryptographic backup verification, automatic recovery snapshots, full vault restoration into a temporary SQLite database with atomic replacement, and non-conflicting merge import strategies.

## Repository Context

- Format & Envelope: `internal/backup/format` (`Header`, `BackupPayload`, `EncryptBackup`, `DecryptBackup`, `InspectHeader`).
- Storage: `internal/storage/sqlite` (`VaultMetadataRepository`, `RecordRepository`, `TagRepository`, DB transactions).
- Domain & Crypto: `internal/domain`, `internal/crypto/kdf`, `internal/crypto/keys`, `internal/crypto/cipher`.
- Services: `internal/service/vault_service.go`, `internal/service/record_service.go`, `internal/service/session.go`.

## Requirement Coverage

- **R1 → Atomic Backup Creation & Custom Passphrase Export:**
  - Implement `BackupService.CreateBackup(ctx, outputPath, opts)`:
    - Validate output path and check for existing file (error unless `opts.Overwrite` is true).
    - Query all active tags, records, and historical revisions from repositories using the active session.
    - Build `BackupPayload` with logical entries, histories, and tags.
    - If `opts.ExportPassphrase` is set:
      - Generate fresh Argon2id KDF params, wrap active `VaultKey` with export passphrase.
    - Else:
      - Use active master password / wrapped key envelope from `vault_meta`.
    - Write to temporary file `.gvault.tmp-<random>` with `0600` permissions.
    - Verify written file structurally.
    - Atomically rename temporary file to `outputPath`.
- **R2 → Structural & Cryptographic Backup Verification:**
  - Implement `BackupService.VerifyBackup(ctx, backupPath, password)`:
    - Open backup file and execute `format.InspectHeader` (checking magic, version, bounds, KDF parameters).
    - If password is provided:
      - Parse header, unwrap Vault Key from header using password.
      - Decrypt payload using derived `BackupKey` and `BackupAAD`.
      - Validate payload integrity (manifest counts vs actual items, no duplicate IDs, no orphaned histories).
      - Return verified manifest summary and structural statistics.
- **R3 → Full Vault Restore with Atomic Swap & Safety Snapshot:**
  - Implement `BackupService.RestoreBackup(ctx, backupPath, password, opts)`:
    - Pre-validate backup cryptographically before altering any storage state.
    - Create timestamped safety snapshot of existing vault (`<vaultPath>.recovery-<timestamp>.gvault`).
    - Initialize temporary SQLite database file (`<vaultPath>.tmp-<random>`).
    - Write `vault_meta` with restored `VaultKey` and envelope.
    - Insert all tags, records, and histories with fresh record encryption nonces.
    - Close temporary database, verify database integrity, and atomically rename over active vault DB.
- **R4 → Merge Import Strategy:**
  - Implement `BackupService.MergeBackup(ctx, backupPath, password, strategy)`:
    - Requires active unlocked session.
    - Decrypt backup payload using provided password.
    - Detect conflicts against existing records (matching ID or matching Title/Type).
    - Apply `strategy`:
      - `ConflictSkip`: skip importing conflicting records.
      - `ConflictOverwrite`: update existing record with backup version.
      - `ConflictRename`: append suffix (e.g. ` (Imported)`) and generate new record ID.
    - Create missing tags and insert non-conflicting records and histories into active vault within a single transaction.

## Architecture

```
internal/service/
├── backup_service.go       # BackupService interface and implementation
├── backup_service_test.go  # Unit and integration tests for export, verify, restore, and merge
```

## Components Affected

- `internal/service/backup_service.go`
- `internal/service/backup_service_test.go`
- `internal/service/errors.go`

## Data Changes

- None to existing SQLite schema; backup files created on disk as `.gvault`.

## API Changes

```go
package service

type ConflictStrategy string

const (
    ConflictSkip      ConflictStrategy = "skip"
    ConflictOverwrite ConflictStrategy = "overwrite"
    ConflictRename    ConflictStrategy = "rename"
)

type BackupCreateOptions struct {
    ExportPassphrase string
    Overwrite        bool
}

type BackupVerifyResult struct {
    Inspection *format.HeaderInspection
    Manifest   *format.BackupManifest
    IsVerified bool
}

type BackupRestoreOptions struct {
    SkipSnapshot bool
}

type BackupMergeResult struct {
    ImportedCount int
    SkippedCount  int
    RenamedCount  int
}

type BackupService struct {
    metaRepo   sqlite.VaultMetadataRepository
    recordRepo sqlite.RecordRepository
    tagRepo    sqlite.TagRepository
    dbPath     string
}

func NewBackupService(metaRepo sqlite.VaultMetadataRepository, recordRepo sqlite.RecordRepository, tagRepo sqlite.TagRepository, dbPath string) *BackupService
func (s *BackupService) CreateBackup(ctx context.Context, sess *Session, outputPath string, opts BackupCreateOptions) error
func (s *BackupService) VerifyBackup(ctx context.Context, backupPath string, password string) (*BackupVerifyResult, error)
func (s *BackupService) RestoreBackup(ctx context.Context, backupPath string, password string, opts BackupRestoreOptions) error
func (s *BackupService) MergeBackup(ctx context.Context, sess *Session, backupPath string, password string, strategy ConflictStrategy) (*BackupMergeResult, error)
```

## Integration Changes

- Consumed by CLI commands (`govault backup create`, `govault backup restore`, `govault backup verify`, `govault backup merge`) in `FEAT-005`.

## Implementation Sequence

1. Define backup options, error types, and `BackupService` struct in `internal/service/backup_service.go`.
2. Implement `CreateBackup` with custom passphrase wrapping and atomic file write.
3. Implement `VerifyBackup` for structural and full cryptographic verification.
4. Implement `RestoreBackup` with pre-validation, recovery snapshots, and atomic DB swap.
5. Implement `MergeBackup` with conflict resolution strategies (`skip`, `overwrite`, `rename`).
6. Write integration tests in `backup_service_test.go` covering all acceptance scenarios and edge cases.

## Test Strategy

- **Export & Restore Test:** Export backup with master password and restore into clean directory; assert all records, histories, and tags match.
- **Custom Passphrase Test:** Export with distinct passphrase, verify restoration succeeds with export passphrase and fails with original master password.
- **Pre-Validation Safety Test:** Attempt to restore corrupted backup; verify live database is completely untouched and error is returned.
- **Merge Strategy Tests:** Test `skip`, `overwrite`, and `rename` conflict strategies against colliding records.

## Risks

- Platform differences in atomic file rename; mitigated by placing temporary files in the same parent directory as destination path.

## Assumptions

- SQLite database file path is accessible to `BackupService` for snapshotting and atomic replacement.
