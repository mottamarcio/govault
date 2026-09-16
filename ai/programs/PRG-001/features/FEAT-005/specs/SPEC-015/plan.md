# Implementation Plan: SPEC-015 (CLI Generators, Clipboard Integration, Diagnostics & Backup Commands)

## Summary

This plan specifies the technical design, architectural structure, and implementation sequence for the standalone utility and maintenance CLI commands of GoVault:
1. `govault generate`: Cryptographic password and Diceware passphrase generator with Shannon entropy calculation.
2. `govault copy` / `govault clip`: Secure clipboard copy with configurable timed auto-clear background worker.
3. `govault backup`, `govault restore`, and `govault inspect`: Encrypted portable `.gvault` backup management, atomic restoration with pre-validation snapshots, and header/manifest inspection.
4. `govault doctor`: 100% offline local environment diagnostic tool validating file permissions, SQLite DB integrity, crypto suites, and zero-network security invariants.

## Repository Context

- `internal/service/password_gen.go` & `passphrase.go`: Core generation algorithms with entropy and strength calculations (`service.GeneratePassword`, `service.GeneratePassphrase`, `service.CalculateEntropy`, `service.EvaluateStrength`).
- `internal/service/clipboard.go`: `ClipboardService` and `ClipboardDriver` abstractions with timed auto-clear routines.
- `internal/service/backup_service.go`: Encrypted `.gvault` export (`CreateBackup`), structural/manifest inspection (`VerifyBackup`), and atomic restore engine (`RestoreBackup`).
- `internal/backup/format/`: Binary header parsing, payload serialization, and inspection logic (`InspectHeader`).
- `internal/cli/`: `root.go`, `output.go`, `prompt.go`, `session_helper.go` providing command execution, formatting, and authentication.

## Requirement Coverage

- **R1 -> Password & Passphrase Generator (`govault generate`):**
  - Implemented in `internal/cli/generate.go` as `newGenerateCmd(appCtx)`.
  - Supports password flags: `--length` (default 20), `--no-upper`, `--no-digits`, `--no-symbols`, `--no-ambiguous`.
  - Supports passphrase flags: `--passphrase`, `--words` (default 5), `--delimiter` (default `-`), `--capitalize`.
  - Calculates Shannon entropy and strength grade.
  - When `--raw` / `-r` is provided, outputs only the raw password/passphrase string to stdout with zero decoration (for shell pipelines).
  - Emits JSON with length, entropy, and strength when `--json` is supplied.
- **R2 -> Secure Clipboard Copy (`govault copy` / `govault clip`):**
  - Implemented in `internal/cli/copy.go` as `newCopyCmd(appCtx)` (with alias `clip`).
  - Fetches target secret record by ID or title; extracts password (or specified `--field <key>`).
  - Copies to system clipboard driver and initiates auto-clear timer (default 45s, configurable with `--clear-after <duration>`).
  - Outputs `Copied to clipboard. Will auto-clear in 45s.` without echoing plaintext secret to stdout or logs.
  - Gracefully falls back on headless systems without X11/Wayland with an informative message to use `govault get <id> --raw`.
- **R3 -> Backup & Restore Management (`govault backup` / `govault restore` / `govault inspect`):**
  - Implemented in `internal/cli/backup.go` as `newBackupCmd(appCtx)`, `newRestoreCmd(appCtx)`, and `newInspectCmd(appCtx)`.
  - `govault backup <output.gvault>`: Prompts master password, supports `--passphrase` for custom export password, supports `--overwrite`, executes `backupSvc.CreateBackup`.
  - `govault restore <input.gvault>`: Prompts backup password, performs pre-validation, creates recovery snapshot (`.recovery-<timestamp>.gvault`), and atomically replaces active vault DB via `backupSvc.RestoreBackup`.
  - `govault inspect <backup.gvault>`: Uses `format.InspectHeader` or `backupSvc.VerifyBackup` to render unauthenticated structural header metadata or authenticated manifest stats in tabular/JSON format.
- **R4 -> GoVault Doctor Diagnostic (`govault doctor`):**
  - Implemented in `internal/cli/doctor.go` as `newDoctorCmd(appCtx)`.
  - Executes comprehensive offline checks:
    1. Vault DB existence & path resolution.
    2. File (`0600`) and directory (`0700`) POSIX permission compliance.
    3. SQLite database integrity (`PRAGMA integrity_check`).
    4. Crypto suite compatibility (`argon2id-hkdf-xchacha20poly1305`).
    5. Clipboard driver status.
    6. Offline invariant verification (0 network listener sockets).
  - Emits colorized/formatted check results (or JSON) and exits with code 0 if all critical checks pass, or 1 if errors are detected.

## Architecture

```
                          [User / Shell Pipeline]
                                    │
                                    ▼
                          [Cobra Command Tree]
      ┌─────────────┬─────────────┬─────────────┬─────────────┬─────────────┐
      │ generateCmd │   copyCmd   │  backupCmd  │ restoreCmd  │  doctorCmd  │
      └──────┬──────┴──────┬──────┴──────┬──────┴──────┬──────┴──────┬──────┘
             │             │             │             │             │
             ▼             ▼             ▼             ▼             ▼
       [Password &   [Clipboard    [BackupService [BackupService  [Doctor
       Passphrase     Service &     & Format]      & SQLite DB]   Integrity
       Generators]    Driver]                                     Engine]
```

## Components Affected

- `internal/cli/root.go`: Register `generate`, `copy` (alias `clip`), `backup`, `restore`, `inspect`, `doctor` subcommands.
- `internal/cli/generate.go`: Implementation of `govault generate`.
- `internal/cli/copy.go`: Implementation of `govault copy` / `govault clip`.
- `internal/cli/backup.go`: Implementation of `govault backup`, `govault restore`, and `govault inspect`.
- `internal/cli/doctor.go`: Implementation of `govault doctor`.
- `internal/cli/utility_test.go`: End-to-end integration and unit tests covering all utility commands.

## Data Changes

None. Uses existing SQLite tables and `.gvault` binary serialization format.

## API Changes

- Exported CLI command constructors:
  - `newGenerateCmd(appCtx *AppContext) *cobra.Command`
  - `newCopyCmd(appCtx *AppContext) *cobra.Command`
  - `newBackupCmd(appCtx *AppContext) *cobra.Command`
  - `newRestoreCmd(appCtx *AppContext) *cobra.Command`
  - `newInspectCmd(appCtx *AppContext) *cobra.Command`
  - `newDoctorCmd(appCtx *AppContext) *cobra.Command`
- AppContext addition: optional custom `ClipboardDriver` field in `AppContext` to facilitate testing clipboard copy routines in headless unit test environments.

## Integration Changes

- Seamless POSIX pipe compatibility with `--raw` on `govault generate`.
- Headless fallbacks for clipboard integration.

## Implementation Sequence

1. **AppContext Extension:** Support optional `ClipboardDriver` on `AppContext` for headless testing.
2. **Password & Passphrase Generator (`govault generate`):** Implement `internal/cli/generate.go` with character flags, passphrase mode, entropy calculations, raw output, and JSON mode.
3. **Secure Clipboard Copy (`govault copy` / `govault clip`):** Implement `internal/cli/copy.go` with secret resolution, field extraction, clipboard copy, and auto-clear timer initiation.
4. **Backup, Restore & Inspection Commands (`govault backup`, `govault restore`, `govault inspect`):** Implement `internal/cli/backup.go` with passphrase prompting, file overwrite checks, recovery snapshots, and header inspection.
5. **Doctor Diagnostics (`govault doctor`):** Implement `internal/cli/doctor.go` inspecting permissions, SQLite PRAGMA integrity, crypto suite, and offline status.
6. **Root Wiring & Comprehensive Test Suite:** Wire commands to `rootCmd` in `root.go` and add unit/race integration tests in `internal/cli/utility_test.go`.

## Test Strategy

- **Generator Tests:** Validate password lengths (8 to 128), character set exclusions, passphrase word counts (3 to 12), `--raw` piping, and JSON structure.
- **Clipboard Tests:** Use `InMemoryClipboardDriver` to test password copying, field extraction, custom clear timeout, and non-leak confirmation message.
- **Backup/Restore Lifecycle Tests:** Test `backup` -> `inspect` (header & manifest) -> `restore` -> verify active DB content and recovery snapshot creation.
- **Doctor Diagnostic Tests:** Test healthy vault reporting all green checks, as well as detecting non-existent DB or corrupted permissions.
- **Race Condition Verification:** Execute `go test -v -race ./internal/cli/...`.

## Risks

- *Risk:* Clipboard interaction on headless Linux CI or test servers failing due to missing DISPLAY/WAYLAND.
  *Mitigation:* Use `InMemoryClipboardDriver` in tests and provide clear graceful fallback in CLI when no system driver is detected.
- *Risk:* Accidental database corruption during restore.
  *Mitigation:* `BackupService.RestoreBackup` creates an automatic recovery snapshot and restores into a temporary SQLite DB first before atomic rename.

## Assumptions

- Backups utilize the existing `.gvault` format (v1).
- Shannon entropy estimation reflects the standard information theory formulas implemented in `service.CalculateEntropy`.
