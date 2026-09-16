---
id: SPEC-015
type: spec
status: ready
parent: FEAT-005
depends_on:
  - SPEC-010
  - SPEC-011
  - SPEC-012
  - SPEC-013
supersedes: []
---

# SPEC-015: CLI Generators, Clipboard Integration, Diagnostics & Backup Commands

## Intent

Specify the standalone utility commands of the CLI: password & Diceware passphrase generator (`govault generate`), secure clipboard copy with auto-clear timer (`govault copy` / `govault clip`), offline diagnostics (`govault doctor`), and backup commands (`govault backup`, `govault restore`, `govault inspect`).

## Requirements

- **R1: Password & Passphrase Generator (`govault generate`):**
  - Implement `govault generate` with flags for length (`--length`, default 20), character sets (`--no-upper`, `--no-digits`, `--no-symbols`), and ambiguous character exclusions (`--no-ambiguous`).
  - Implement passphrase mode via `--passphrase` (word count `--words`, delimiter `--delimiter`, `--capitalize`).
  - Calculate and display Shannon entropy in bits and strength estimation (unless `--raw` is specified).
- **R2: Secure Clipboard Copy (`govault copy` / `govault clip`):**
  - Implement `govault copy <id>` to copy password (or specified field via `--field`) to system clipboard.
  - Automatically launch timed background clearer (default 45s, configurable via `--clear-after <duration>`).
  - Print confirmation to stdout: `Copied to clipboard. Will auto-clear in 45s.` without leaking the secret value.
- **R3: Backup & Restore Management (`govault backup` / `govault restore` / `govault inspect`):**
  - `govault backup <output.gvault>`: Export vault to encrypted `.gvault` file (supporting `--passphrase` for custom export passphrase and `--overwrite`).
  - `govault restore <input.gvault>`: Restore backup with pre-validation, automatic recovery snapshot, and atomic replacement.
  - `govault inspect <backup.gvault>`: Display non-secret structural header information (created date, versions, KDF parameters) or full authenticated manifest stats if password is provided.
- **R4: GoVault Doctor Diagnostic (`govault doctor`):**
  - Implement `govault doctor` to verify local system health: database file existence, directory and file permissions (`0600`/`0700`), SQLite integrity check, cryptographic suite compatibility, clipboard driver availability, and offline guarantees (0 network packages).

## Acceptance Scenarios

- **Scenario 1: Password Generation for Scripting**
  - *When* `govault generate --length 32 --raw` is executed
  - *Then* exactly 32 random characters are written to stdout with high entropy and exit code 0.
- **Scenario 2: Secure Copy with Auto-Clear**
  - *Given* a record with password "MySecretPass"
  - *When* `govault copy <id>` is executed
  - *Then* the password is in clipboard and a background timer is launched to clear it after the configured timeout.
- **Scenario 3: Vault Doctor Diagnostics**
  - *Given* a healthy GoVault installation
  - *When* `govault doctor` is executed
  - *Then* all checks pass with green checkmarks and return exit code 0.

## Edge Cases

- Generating password with length smaller than 8 returns an error with exit code 1.
- Executing `govault copy` on a headless environment without X11/Wayland falls back gracefully with an informative error instructing use of `--raw` or stdout.
- Backup creation fails if target file exists and `--overwrite` flag is not provided.

## Constraints

- Generator must strictly use `crypto/rand` only.
- `doctor` must execute 100% offline without attempting network calls.
- Clipboard operations must never echo secrets to stdout or terminal logs.

## Non-Goals

- GUI or web browser extension integration.
