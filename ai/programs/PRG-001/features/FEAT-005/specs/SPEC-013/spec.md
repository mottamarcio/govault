---
id: SPEC-013
type: spec
status: ready
parent: FEAT-005
depends_on:
  - SPEC-004
  - SPEC-008
supersedes: []
---

# SPEC-013: CLI Architecture, Root Command Tree & Vault Lifecycle Commands

## Intent

Specify the Cobra command structure, global flags (`--vault`, `--json`, `--quiet`), masked terminal password prompting, POSIX standard exit codes, and vault lifecycle commands (`init`, `status`, `unlock`, `lock`, `passwd`).

## Requirements

- **R1: Root Command & Global Flags:** Implement root `govault` Cobra command with global flags:
  - `--vault <path>` (custom database file path, defaults to `~/.govault/vault.db` or `GOVAULT_PATH` env var).
  - `--json` (structured JSON output for machine consumption and scripting).
  - `--quiet` / `-q` (suppress informational output on stdout, errors still emitted to stderr).
  - `--version` (print application version and crypto suite version).
- **R2: Secure Terminal Input & Ingestion:** Provide masked interactive password prompting without echo via `golang.org/x/term` with fallback to piped stdin (e.g. `echo "secret" | govault unlock --stdin` or `GOVAULT_PASSWORD` env var). Command-line argument password flags (e.g. `--password secret`) are strictly forbidden to prevent leaking secrets in process tables (`ps`).
- **R3: Vault Initialization & Status:** Implement `govault init` (initialize new vault with master password confirmation and Argon2id KDF parameter setup) and `govault status` (inspect unencrypted vault header metadata, initialization state, schema version, and lock state).
- **R4: Session Lifecycle & Password Rotation:** Implement `govault unlock` (prompt master password, verify envelope, start session / check status), `govault lock` (explicitly terminate active session and wipe in-memory cached keys), and `govault passwd` (re-wrap Vault Key under new master password with fresh Argon2id salt and atomic metadata update).

## Acceptance Scenarios

- **Scenario 1: Vault Initialization**
  - *Given* an empty directory with no vault database
  - *When* `govault init` is executed with matching master passwords
  - *Then* a new `vault.db` is created with permissions `0600` and `vault_metadata` initialized with Argon2id parameters.
- **Scenario 2: JSON Output Mode for Status**
  - *Given* an initialized vault
  - *When* `govault status --json` is executed
  - *Then* output on stdout is valid JSON containing `{"vault_id":"...", "status":"locked", "schema_version":1}` with exit code 0.
- **Scenario 3: Password Rotation**
  - *Given* an initialized vault with master password "OldSecret123!"
  - *When* `govault passwd` is executed with valid old password and new password "NewSecret456!"
  - *Then* the vault metadata envelope is updated atomically and subsequent unlocks require the new password.

## Edge Cases

- Running `govault init` on an already-initialized vault returns a non-zero exit code (code 1) with an informative error.
- Providing mismatched password confirmation during `init` or `passwd` prompts aborts immediately without modifying any files.
- Invoking commands against a non-existent vault displays `vault not initialized: run govault init` to stderr with exit code 1.

## Constraints

- Pure Go and zero network dependencies.
- Passwords must never appear in `argv` / command-line flags.
- Output formatting must strictly direct data to stdout and errors to stderr.

## Non-Goals

- Secret record CRUD operations (specified in SPEC-014).
- Backup and generator tools (specified in SPEC-015).
