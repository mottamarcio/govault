---
type: tasks
for: SPEC-012
---

# Tasks

## TASK-040 — Atomic Backup Export & Verification Application Service

- [x] Completed
- **Serves:** SPEC-012:R1, SPEC-012:R2
- **Depends on:** none
- **Files/Components:** `internal/service/backup_service.go`, `internal/service/backup_service_test.go`
- **Verification:** Ran `go test -v -race ./internal/service -run 'TestBackupExportAndVerify'` verifying atomic `.gvault` file creation (`0600` permissions), export with active master password vs custom passphrase re-wrapping, structural inspection, and cryptographic verification with manifest counting. Evidence: `PASS TestBackupExportAndVerify`.

## TASK-041 — Full Vault Restoration with Safety Snapshot & Atomic Swap

- [x] Completed
- **Serves:** SPEC-012:R3
- **Depends on:** TASK-040
- **Files/Components:** `internal/service/backup_service.go`, `internal/service/backup_service_test.go`
- **Verification:** Ran `go test -v -race ./internal/service -run 'TestBackupRestore'` verifying pre-validation safety checks, creation of timestamped recovery snapshots, temporary SQLite database construction, and atomic replacement of active vault files without data corruption. Evidence: `PASS TestBackupRestore`.

## TASK-042 — Selective Merge Import Strategy

- [x] Completed
- **Serves:** SPEC-012:R4
- **Depends on:** TASK-040
- **Files/Components:** `internal/service/backup_service.go`, `internal/service/backup_service_test.go`
- **Verification:** Ran `go test -v -race ./internal/service -run 'TestBackupMerge'` verifying merge imports into an unlocked active vault using `skip`, `overwrite`, and `rename` conflict strategies. Evidence: `PASS TestBackupMerge`.
