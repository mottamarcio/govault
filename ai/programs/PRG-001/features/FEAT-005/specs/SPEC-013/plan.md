---
type: plan
for: SPEC-013
status: ready
---

# Implementation Plan: CLI Architecture, Root Command Tree & Vault Lifecycle Commands

## Summary

Design and implement the CLI application entrypoint and root command structure using Cobra under `cmd/govault` and `internal/cli`. This includes root global flags (`--vault`, `--json`, `--quiet`, `--version`), masked interactive password prompting via `golang.org/x/term` with fallback to piped stdin and environment variables, POSIX exit codes, error handling, and vault lifecycle subcommands (`init`, `status`, `unlock`, `lock`, `passwd`).

## Repository Context

- Application Services: `internal/service/vault_service.go`, `internal/service/session.go`.
- Database & Storage: `internal/storage/sqlite/db.go`, `internal/storage/sqlite/metadata_repo.go`.
- Cryptography: `internal/crypto/kdf`, `internal/crypto/keys`.
- Dependencies: `github.com/spf13/cobra`, `golang.org/x/term`.

## Requirement Coverage

- **R1 → Root Command & Global Flags:**
  - Define root command `govault` in `internal/cli/root.go`:
    - Global persistent flags:
      - `--vault <path>` (defaults to `~/.govault/vault.db` or `GOVAULT_PATH`).
      - `--json` (boolean, triggers machine-readable structured JSON output).
      - `--quiet` / `-q` (boolean, suppresses informational logs on stdout).
    - `--version` flag prints GoVault version, build date, and crypto suite version.
    - Global error handler printing friendly message to `stderr` and returning exit code 1.
- **R2 → Secure Terminal Input & Ingestion:**
  - Implement `internal/cli/prompt.go`:
    - `PromptPassword(prompt string) (string, error)` using `term.ReadPassword(int(syscall.Stdin))` without echo.
    - `PromptPasswordConfirm(prompt string) (string, error)` prompting twice and ensuring matching inputs.
    - Fallback: check `GOVAULT_PASSWORD` environment variable or stdin if terminal is not a TTY.
    - Strict prohibition: no `--password` flag accepted anywhere in the CLI.
- **R3 → Vault Initialization & Status:**
  - Implement `govault init` in `internal/cli/init.go`:
    - Prompts master password twice with confirmation.
    - Creates database directories with `0700` permissions.
    - Opens SQLite database, applies migrations, and calls `VaultService.Init(ctx, password)`.
    - Outputs confirmation or JSON formatted status.
  - Implement `govault status` in `internal/cli/status.go`:
    - Inspects vault metadata without password requirement via `VaultService.Inspect(ctx)`.
    - Outputs tabular info (Vault ID, Status, Schema Version, Created At) or JSON payload.
- **R4 → Session Lifecycle & Password Rotation:**
  - Implement `govault unlock` in `internal/cli/unlock.go`:
    - Prompts password, calls `VaultService.Unlock(ctx, password)`.
  - Implement `govault lock` in `internal/cli/lock.go`:
    - Wipes cached keys and outputs locked status.
  - Implement `govault passwd` in `internal/cli/passwd.go`:
    - Prompts for current master password and new master password (confirmed twice).
    - Calls `VaultService.ChangePassword(ctx, oldPassword, newPassword)`.

## Architecture

```
cmd/govault/
└── main.go              # CLI main entrypoint

internal/cli/
├── root.go              # Root command, global flags, version, context setup
├── prompt.go            # Secure masked terminal password prompt & stdin ingestion
├── output.go            # JSON and tabular formatted printer helpers
├── init.go              # govault init command
├── status.go            # govault status command
├── unlock.go            # govault unlock command
├── lock.go              # govault lock command
├── passwd.go            # govault passwd command
├── root_test.go         # Tests for root flags and global behavior
├── lifecycle_test.go    # Tests for init, status, unlock, lock, passwd commands
└── prompt_test.go       # Tests for password ingestion and validation
```

## Components Affected

- `cmd/govault/main.go`
- `internal/cli/root.go`
- `internal/cli/prompt.go`
- `internal/cli/output.go`
- `internal/cli/init.go`
- `internal/cli/status.go`
- `internal/cli/unlock.go`
- `internal/cli/lock.go`
- `internal/cli/passwd.go`
- Tests under `internal/cli/`

## Data Changes

- Default vault location created at `~/.govault/vault.db` with directory permissions `0700` and file permissions `0600`.

## API Changes

```go
package cli

type AppContext struct {
    VaultPath string
    JSON      bool
    Quiet     bool
    DB        *sqlite.DB
    VaultSvc  *service.VaultService
    In        io.Reader
    Out       io.Writer
    Err       io.Writer
}

func NewRootCommand(appCtx *AppContext) *cobra.Command
func Execute() error
```

## Integration Changes

- Integrates with Application Services (`VaultService`).

## Implementation Sequence

1. Implement `AppContext`, `output.go`, and `prompt.go` in `internal/cli`.
2. Implement root command tree and global flags in `internal/cli/root.go` and `cmd/govault/main.go`.
3. Implement `init.go`, `status.go`, `unlock.go`, `lock.go`, and `passwd.go`.
4. Implement unit and CLI end-to-end integration tests using in-memory streams (`bytes.Buffer`) and temporary directory databases.

## Test Strategy

- **CLI Flag Tests:** Assert `--vault`, `--json`, `--quiet`, `--version` behave properly.
- **Prompt Tests:** Validate password ingestion from stdin stream and confirmation validation.
- **Lifecycle Integration Tests:** Execute `init` → `status` → `unlock` → `passwd` → `status` using custom `AppContext` writers.

## Risks

- Terminal raw mode handling on non-standard terminals; mitigated by fallback to piped stdin reader if stdin is not a TTY.

## Assumptions

- Standard POSIX exit code 0 indicates success; code 1 indicates operational error.
