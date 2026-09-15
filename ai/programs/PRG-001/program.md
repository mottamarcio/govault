---
id: PRG-001
type: program
status: active
---

# PRG-001: GoVault Offline Password & Secrets Manager

## Problem

Developers, system administrators, and security-conscious users require a fast, local, and reliable password and secrets manager that operates entirely inside the terminal. Existing solutions often depend on cloud synchronization, mandatory user accounts, remote telemetry, or complex GUI dependencies, which introduce unwanted external attack surfaces, network requirements, and friction in automated command-line workflows.

## Users and Stakeholders

- **Developers & DevOps Engineers:** Need scriptable CLI commands to generate, retrieve, and store secrets seamlessly in terminal workflows.
- **System Administrators:** Require secure, offline local credential management without remote sync risks or daemon overhead.
- **Security-Conscious Terminal Users:** Want an interactive TUI for browsing, searching, and managing secret records with zero telemetry or network calls.

## Desired Outcome

Deliver GoVault as a single, self-contained, offline Go binary providing:
1. A rich interactive Terminal User Interface (TUI) powered by Bubble Tea.
2. A scriptable, composable Command Line Interface (CLI) powered by Cobra.
3. Robust local storage with SQLite using envelope-encrypted records (XChaCha20-Poly1305 and Argon2id).
4. Portable, encrypted `.gvault` backup archives for manual data transfer.
5. Absolute offline operation with zero telemetry, remote sync, or networking dependencies.

## Scope

- **Core Cryptography (Crypto Suite v1):** Argon2id KDF, HKDF-SHA-256 key derivation, XChaCha20-Poly1305 AEAD, secure random generation, key hierarchy (Master Key, KEK, MCK, Vault Key, Record Keys), and password rotation without re-encrypting records.
- **Local Storage Engine:** SQLite persistence with WAL mode, schema migrations, metadata tracking, encrypted record envelopes, soft deletion, and revision history.
- **Secrets Management Domain:** Secret entries (titles, usernames, passwords, URLs, custom key-value pairs, notes, tags), CRUD operations, and in-memory search index.
- **Password & Passphrase Generator:** Configurable character sets, lengths, Diceware passphrases, and entropy estimation.
- **Portable Backups:** Creation, inspection, verification, and restoration of `.gvault` encrypted backup files.
- **Dual Presentation Interfaces:**
  - Interactive TUI (dashboard, navigation, forms, search, clipboard actions, auto-lock timeouts).
  - Scriptable CLI (commands, subcommands, stdin/stdout flags, exit codes, tabular/JSON formatting).
- **Security & OS Ergonomics:** Timed clipboard clearing, standard OS configuration paths (XDG on Linux/Unix), and signal handling (`SIGINT`, `SIGTERM`).

## Non-Goals

- Cloud synchronization, automated remote backups, or remote server APIs.
- Account systems, user registration, or authentication servers.
- Telemetry, usage analytics, crash reporting over network, or automatic update checks.
- Remote breach checking or remote website favicon/metadata scraping.
- Hardware key token management (e.g. FIDO2/YubiKey) for v1.
- Multi-user access control or team credential sharing.

## Constraints

- Strictly offline: No network imports or networking libraries in production builds.
- Single Core / Dual Interface: CLI and TUI must use identical domain and application service layers.
- Sensitive data must be encrypted before reaching SQLite storage.
- Nonce uniqueness and AAD binding must be strictly guaranteed.
- Secrets must never be accepted via command-line arguments (only via stdin, prompts, or secure environment variables).
- Adhere to the Constitution invariants ([`ai/memory/constitution.md`](file:///home/marciovcm/workspace/golang/govault/ai/memory/constitution.md)).

## Success Criteria

- Clean compilation of a single standalone `govault` binary with zero network packages.
- 100% test pass rate across domain, cryptographic, storage, CLI, and backup modules.
- End-to-end functionality of all CLI commands (init, lock/unlock, create, list, get, update, delete, generate, backup, restore).
- Interactive TUI responsive, ergonomic, and fully feature-equivalent with CLI capabilities.
- Verification that `.gvault` backups can be exported, inspected, and restored across fresh installations.

## Relevant Knowledge

- [KNOW-001](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-001-product-requirements.md) — Product Requirements and CLI/TUI Capabilities
- [KNOW-002](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-002-threat-model.md) — Security Threat Model and Trust Boundaries
- [KNOW-003](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-003-cryptography.md) — Cryptographic Architecture and Key Hierarchy
- [KNOW-004](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-004-vault-format.md) — Vault Database Storage Format and Schema
- [KNOW-005](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-005-backup-format.md) — Portable Backup File Format (.gvault)
- [KNOW-006](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-006-architecture.md) — GoVault Software Architecture and Package Design

## Open Questions

- Selection of SQLite driver for Go (`modernc.org/sqlite` pure Go vs `mattn/go-sqlite3` CGO).
- Serialization format for `.gvault` backup payload (CBOR vs JSON).
