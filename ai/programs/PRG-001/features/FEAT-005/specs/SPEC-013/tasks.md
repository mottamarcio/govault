---
type: tasks
for: SPEC-013
---

# Tasks

## TASK-043 — Root Cobra Command Tree, Global Flags & Output Formatters

- [x] Completed
- **Serves:** SPEC-013:R1
- **Depends on:** none
- **Files/Components:** `internal/cli/root.go`, `internal/cli/output.go`, `cmd/govault/main.go`, `internal/cli/root_test.go`
- **Verification:** Ran `go test -v -race ./internal/cli -run 'TestRootCommand|TestGlobalFlags'` verifying persistent flags (`--vault`, `--json`, `--quiet`, `--version`), default database path resolution (`~/.govault/vault.db` or `GOVAULT_PATH`), JSON/tabular formatting, and error output to stderr with POSIX exit codes. Evidence: `PASS TestRootCommandVersionAndFlags`.

## TASK-044 — Secure Masked Password Prompting & Stdin Ingestion

- [x] Completed
- **Serves:** SPEC-013:R2
- **Depends on:** none
- **Files/Components:** `internal/cli/prompt.go`, `internal/cli/prompt_test.go`
- **Verification:** Ran `go test -v -race ./internal/cli -run 'TestPrompt'` verifying terminal masked input via `golang.org/x/term`, password confirmation matching, non-TTY stdin pipe ingestion, and `GOVAULT_PASSWORD` environment fallback. Evidence: `PASS TestPromptPassword`.

## TASK-045 — Vault Lifecycle Commands (init, status, unlock, lock, passwd)

- [x] Completed
- **Serves:** SPEC-013:R3, SPEC-013:R4
- **Depends on:** TASK-043, TASK-044
- **Files/Components:** `internal/cli/init.go`, `internal/cli/status.go`, `internal/cli/unlock.go`, `internal/cli/lock.go`, `internal/cli/passwd.go`, `internal/cli/lifecycle_test.go`
- **Verification:** Ran `go test -v -race ./internal/cli -run 'TestVaultLifecycle'` verifying end-to-end command execution: `govault init` (creating 0600 vault with Argon2id), `govault status` (non-secret inspection), `govault unlock` / `govault lock` session transitions, and `govault passwd` master password rotation. Evidence: `PASS TestVaultLifecycle`.
