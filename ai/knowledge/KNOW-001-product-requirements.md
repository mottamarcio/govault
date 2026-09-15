---
id: KNOW-001
type: knowledge
status: active
sources:
  - path: ai/raw/product-requirements-document.md
    fingerprint: sha256:30f76d6cd490aa7eb4092a621ed21184f881eae4e67c473a4ca1be9ee83676a3
---

# KNOW-001: Product Requirements and CLI/TUI Capabilities

## Summary

GoVault is a local-first, terminal-native password and secrets manager written in Go. It operates strictly offline with zero network connectivity, providing dual interfaces (a rich Bubble Tea interactive TUI and a scriptable Cobra CLI) backed by a shared SQLite local storage and cryptographic core.

## Known Facts

- GoVault targets developers, system administrators, and security-conscious users who need local secrets management without accounts, telemetry, or remote services.
- The dual-interface model provides parity: the TUI and CLI operate on the same application/domain core.
- Core capabilities include:
  - Vault management: initialization, master password changes, locking/unlocking, status checks, and self-destruction / purging.
  - Secret entries: create, read, update, delete, search, copy to clipboard, organize by tags/categories, custom key-value fields, and secret revision history.
  - Password generation: configurable length, character sets (upper, lower, digits, symbols), passphrases (diceware wordlists), and entropy calculation.
  - Backup & restore: manual export to and import from portable encrypted `.gvault` backup archives.
  - Secure clipboard integration: automatic clipboard clearing after a configurable timeout (default 45s) and stdout suppression by default.
  - Session and auto-lock: inactivity timeout auto-locking and secure memory handling during unlocked sessions.

## Constraints

- Strictly offline: No telemetry, analytics, cloud sync, remote API calls, update checks, remote breach lookups, or remote favicon/metadata fetching.
- Core code must not import or use networking packages (e.g. `net/http`).
- Passwords and secrets must be masked/hidden by default; secrets are never printed to stdout unless explicitly passed a flag (e.g. `--show` or `--stdout`).
- CLI commands must support non-interactive / scriptable execution (structured output, exit codes, reading passwords from stdin/environment variables where appropriate).

## Unknowns

- Final choice of Diceware / wordlist data embedding (standard EFF long wordlist vs curated embedded lists).
- Exact OS-level background daemon / agent model for cross-process clipboard clearing and auto-lock timeouts across short-lived CLI invocations versus long-running TUI sessions.

## Conflicts

- None identified in the source requirements.

## Provenance

- All requirements, product principles, CLI commands, TUI workflows, and functional capabilities derived from `ai/raw/product-requirements-document.md`.

## Related Topics

- [KNOW-002](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-002-threat-model.md) — Security Threat Model
- [KNOW-006](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-006-architecture.md) — Software Architecture
