---
id: FEAT-002
type: feature
status: active
parent: PRG-001
---

# FEAT-002: SQLite Vault Storage & Persistence

## Capability

Provides local durable storage, database lifecycle management, schema migrations, WAL journaling configuration, transaction handling, encrypted record envelope persistence, and audit/history tracking.

## User Value

Ensures fast, ACID-compliant local database operations with crash-safety and durability, while guaranteeing that no sensitive plaintext payload ever hits disk unencrypted.

## Scope

- SQLite database initialization, connection pooling, and configuration (`PRAGMA journal_mode = WAL`, `PRAGMA synchronous = NORMAL`, `PRAGMA foreign_keys = ON`, `PRAGMA busy_timeout = 5000`).
- Schema versioning and migration engine.
- `vault_metadata` storage (Vault UUID, schema version, crypto suite ID, KDF salt/parameters, Wrapped Vault Key).
- `records` repository: CRUD operations, soft deletion / tombstones, and timestamps.
- `record_history` repository: retaining encrypted previous versions of records.
- `tags` and `record_tags` repository for secret categorization.
- In-memory search index population upon vault unlock.

## Non-Goals

- Storing unencrypted secret titles, passwords, or notes in SQLite tables.
- Remote database connections or SQLite replication over network.

## Constraints

- File permissions on `vault.db`, `vault.db-wal`, `vault.db-shm` must be restricted to owner-only (`0600`).
- All database write operations must execute within transactions.
- Zero network dependencies in SQLite driver setup.

## Relevant Knowledge

- [KNOW-004](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-004-vault-format.md) — Vault Database Storage Format and Schema
- [KNOW-006](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-006-architecture.md) — GoVault Software Architecture and Package Design

## Open Questions

- Final SQLite driver choice: pure-Go `modernc.org/sqlite` vs CGO `mattn/go-sqlite3`.
