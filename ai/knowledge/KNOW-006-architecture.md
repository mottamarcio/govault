---
id: KNOW-006
type: knowledge
status: active
sources:
  - path: ai/raw/architecture.md
    fingerprint: sha256:3eb45e560b7ed8bea17771ac03cdb96e8e1b7e283a073f16308b3eb0872271e2
---

# KNOW-006: GoVault Software Architecture and Package Design

## Summary

Defines the hexagonal/clean software architecture, package boundaries, dependency directions, domain models, and service interfaces for GoVault in Go.

## Known Facts

- Architecture Style: Inward-flowing layered architecture (Clean/Hexagonal Architecture).
- Core Package Structure:
  - `cmd/govault`: Entry point CLI and TUI launcher.
  - `internal/domain`: Core entities (Vault, Record, Tag, Metadata, History, Value Objects) free of external dependencies.
  - `internal/crypto`: Argon2id KDF, HKDF, XChaCha20-Poly1305 AEAD, secure random generators, key hierarchy management.
  - `internal/storage`: SQLite repository implementations, migrations, database connection pooling, WAL configuration.
  - `internal/backup`: Portable backup writer, reader, inspector, and restore logic.
  - `internal/service` (or `internal/app`): Application use cases (VaultService, RecordService, PasswordGenService, BackupService, ClipboardService).
  - `internal/cli`: Cobra command tree, flags, stdin parsing, non-interactive formatters (JSON, table, plain text).
  - `internal/tui`: Bubble Tea models, Bubbles UI components, Lip Gloss styling, keybindings, and interactive screens.
  - `internal/config`: TOML configuration loading and paths (XDG compliant: `~/.config/govault`, `~/.local/share/govault`).
- Dependency Rules:
  - CLI and TUI are peer presentation layers that invoke the Application Services.
  - Neither CLI nor TUI contains cryptographic or database logic directly.
  - Domain models have zero third-party framework dependencies.

## Constraints

- Offline enforcement: The production build and CI must strictly prohibit network packages (e.g. `net`, `net/http`, `net/url` network callers).
- Shared domain and services: All features available in the CLI must share business logic with the TUI, avoiding duplicate implementations.
- Graceful shutdown and signal handling (`SIGINT`, `SIGTERM`) must lock the vault and clear active memory keys/clipboards where applicable.

## Unknowns

- Exact selection of SQLite driver: `modernc.org/sqlite` (pure Go, CGO-free, easier cross-compilation) vs `mattn/go-sqlite3` (CGO required, standard SQLite C library performance).

## Conflicts

- None.

## Provenance

- Package structures, layer responsibilities, dependency flow, and domain designs derived directly from `ai/raw/architecture.md`.

## Related Topics

- [KNOW-001](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-001-product-requirements.md) — Product Requirements
- [KNOW-003](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-003-cryptography.md) — Cryptographic Architecture
- [KNOW-004](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-004-vault-format.md) — Vault Storage Format
