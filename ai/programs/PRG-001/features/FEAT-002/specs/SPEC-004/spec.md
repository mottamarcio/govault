---
id: SPEC-004
type: spec
status: ready
parent: FEAT-002
depends_on:
  - SPEC-001
  - SPEC-002
  - SPEC-003
---

# SPEC-004: SQLite Database Connection Management, Pragmas & Schema Migrations

## Intent

Specify the SQLite database engine initialization, connection pooling, required security and performance PRAGMA configurations (WAL mode, foreign keys, busy timeouts), and migration runner.

## Requirements

- **R1: SQLite Engine Configuration & Pragmas:** The database manager must establish local SQLite connections configured with:
  - `PRAGMA journal_mode = WAL`
  - `PRAGMA synchronous = NORMAL`
  - `PRAGMA foreign_keys = ON`
  - `PRAGMA busy_timeout = 5000`
- **R2: File Permissions Enforcement:** When creating a new database file or opening an existing one on POSIX systems, file permissions for `vault.db`, `vault.db-wal`, and `vault.db-shm` must be restricted to owner-only (`0600`), and parent directories restricted to `0700`.
- **R3: Schema Migration Runner:** The storage layer must provide a deterministic migration engine executing versioned SQL migration steps inside atomic database transactions and tracking the current schema version in `schema_version`.
- **R4: Baseline Schema Provisioning:** Migration v1 must construct the baseline schema tables: `vault_metadata`, `records`, `record_history`, `tags`, and `record_tags`.

## Acceptance Scenarios

- **Scenario 1: Fresh Database Initialization**
  - *Given* an empty database path
  - *When* database connection and migration v1 are executed
  - *Then* tables `vault_metadata`, `records`, `record_history`, `tags`, `record_tags`, and `schema_version` exist, and PRAGMAs (WAL mode, foreign keys) are active.
- **Scenario 2: POSIX Permission Check**
  - *Given* a created SQLite database on a Unix filesystem
  - *When* inspecting file mode bits
  - *Then* the file mode is `0600` (read/write by owner only).
- **Scenario 3: Atomic Schema Migration**
  - *Given* a running migration step with invalid SQL
  - *When* the migration fails
  - *Then* the transaction rolls back cleanly without leaving partial schema changes.

## Edge Cases

- Database path in non-existent directory tree must create parent directories with `0700` permissions.
- Concurrently opened connections must honor `busy_timeout` rather than failing immediately with SQLITE_BUSY.

## Constraints

- SQLite driver must operate completely offline with no network features.
- All structural schema updates must execute in transactions.

## Non-Goals

- Remote SQLite database connections or network replication.
