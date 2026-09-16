---
type: tasks
for: SPEC-015
---

# Tasks

## TASK-050 — Password & Diceware Passphrase Generator Command (`govault generate`)

- [x] Completed
- **Serves:** SPEC-015:R1
- **Depends on:** none
- **Files/Components:** `internal/cli/generate.go`, `internal/cli/root.go`, `internal/cli/utility_test.go`
- **Verification:** Ran `go test -v -race ./internal/cli -run 'TestGenerateCommand'` verifying password generation flags (`--length`, `--no-upper`, `--no-digits`, `--no-symbols`, `--no-ambiguous`), passphrase generation flags (`--passphrase`, `--words`, `--delimiter`, `--capitalize`), Shannon entropy bit calculations and strength grading, `--raw` raw stdout piping, `--json` structured formatting, and boundary error validation (length < 8, words < 3). Evidence: `PASS TestGenerateCommand`.

## TASK-051 — Secure Clipboard Copy & Background Auto-Clear (`govault copy` / `govault clip`)

- [x] Completed
- **Serves:** SPEC-015:R2
- **Depends on:** none
- **Files/Components:** `internal/cli/copy.go`, `internal/cli/root.go`, `internal/cli/utility_test.go`
- **Verification:** Ran `go test -v -race ./internal/cli -run 'TestCopyCommand'` verifying secret retrieval and extraction by record ID/title, copying target secret to clipboard using `ClipboardService`, launching timed auto-clear background worker (default 45s, configurable via `--clear-after`), displaying non-leaking confirmation message `Copied to clipboard. Will auto-clear in 45s.`, and fallback error messaging for headless environments. Evidence: `PASS TestCopyCommand`.

## TASK-052 — Portable Backup, Restore & Inspection Commands (`govault backup`, `govault restore`, `govault inspect`)

- [x] Completed
- **Serves:** SPEC-015:R3
- **Depends on:** none
- **Files/Components:** `internal/cli/backup.go`, `internal/cli/root.go`, `internal/cli/utility_test.go`
- **Verification:** Ran `go test -v -race ./internal/cli -run 'TestBackupRestoreInspectCommands'` verifying `govault backup` exporting encrypted `.gvault` files (with `--passphrase` and `--overwrite`), `govault inspect` rendering structural header metadata and authenticated manifest stats, and `govault restore` performing pre-validation, generating recovery snapshots, and atomically updating the active vault. Evidence: `PASS TestBackupRestoreInspectCommands`.

## TASK-053 — Offline Environment Diagnostic Command (`govault doctor`)

- [x] Completed
- **Serves:** SPEC-015:R4
- **Depends on:** none
- **Files/Components:** `internal/cli/doctor.go`, `internal/cli/root.go`, `internal/cli/utility_test.go`
- **Verification:** Ran `go test -v -race ./internal/cli -run 'TestDoctorCommand'` verifying 100% offline system health checks: vault file and directory existence, POSIX permissions (`0600` / `0700`), SQLite `PRAGMA integrity_check`, cryptographic suite compatibility (`argon2id-hkdf-xchacha20poly1305`), clipboard driver availability, zero-network security invariants, and appropriate POSIX exit codes (0 for healthy, 1 for issues). Evidence: `PASS TestDoctorCommand`.
