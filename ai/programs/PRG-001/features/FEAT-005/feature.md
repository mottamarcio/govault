---
id: FEAT-005
type: feature
status: active
parent: PRG-001
---

# FEAT-005: Command-Line Interface (CLI)

## Capability

Provides a scriptable, ergonomic, and composable Unix command-line interface powered by Cobra, offering full feature parity with the GoVault core.

## User Value

Enables developers and sysadmins to interactively execute single commands, script password retrieval/generation in pipelines, export/import backups, and inspect vault status with standard Unix paradigms.

## Scope

- Cobra command tree:
  - `govault init`: Initialize new vault.
  - `govault lock` / `govault unlock` / `govault status`: Session and status management.
  - `govault add` / `govault get` / `govault list` / `govault edit` / `govault delete`: Secret record operations.
  - `govault search`: Fast title/tag searching.
  - `govault copy`: Copy username or password to clipboard with auto-clear.
  - `govault generate`: Password and passphrase generator with entropy display.
  - `govault backup` / `govault restore` / `govault inspect`: Backup management.
  - `govault passwd`: Master password rotation.
- Formatting options: interactive table format, raw output, and structured `--json` output.
- Input methods: interactive masked prompt, stdin pipe (`-`), or environment variable ingestion.
- Standard POSIX exit codes and error output to stderr.

## Non-Goals

- Accepting passwords directly as command-line argument flags (e.g. `--password secret` is forbidden for security).
- TUI rendering (delegated to FEAT-006).

## Constraints

- Output masking: Secrets are masked or copied to clipboard by default; only revealed with explicit `--show` / `--stdout` flags.
- CLI commands must invoke Application Services and contain zero direct database or crypto implementation logic.

## Relevant Knowledge

- [KNOW-001](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-001-product-requirements.md) — Product Requirements and CLI/TUI Capabilities
- [KNOW-002](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-002-threat-model.md) — Security Threat Model and Trust Boundaries
- [KNOW-006](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-006-architecture.md) — GoVault Software Architecture and Package Design

## Open Questions

- None.
