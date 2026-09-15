---
type: plan
for: SPEC-004
status: ready
---

# Implementation Plan: SQLite Connection Management, Pragmas & Schema Migrations

## Summary

Design and implement the SQLite database manager and migration engine (`internal/storage/sqlite`) responsible for opening/configuring local SQLite databases with strict security and durability PRAGMAs (WAL mode, foreign keys, busy timeout), enforcing POSIX file permissions (`0600`/`0700`), and executing transactional, versioned SQL schema migrations.

## Repository Context

`SPEC-001`, `SPEC-002`, and `SPEC-003` are fully implemented and validated in `internal/crypto`. The storage layer package will reside in `internal/storage/sqlite`. For SQLite in Go without CGO dependencies and completely offline, `modernc.org/sqlite` is selected as the database driver, fitting the pure-Go, cross-platform architecture of GoVault.

## Requirement Coverage

- **R1 → SQLite Engine Configuration & Pragmas:**
  - Implement `Open(path string) (*DB, error)` initializing `sql.DB` with `modernc.org/sqlite`.
  - Apply required PRAGMAs on connection initialization:
    - `PRAGMA journal_mode = WAL;`
    - `PRAGMA synchronous = NORMAL;`
    - `PRAGMA foreign_keys = ON;`
    - `PRAGMA busy_timeout = 5000;`
  - Set connection pool limits: `db.SetMaxOpenConns(1)` (for write serialization safety) or appropriate read/write pool settings.
- **R2 → File Permissions Enforcement:**
  - When opening/creating a database at `path`:
    - Ensure parent directory exists with `0700` (`os.MkdirAll(dir, 0700)`).
    - If database file is created, ensure file mode is `0600` (`os.OpenFile` with `0600` and `os.Chmod(path, 0600)`).
    - Apply `0600` permissions defensively to auxiliary WAL files (`<path>-wal`, `<path>-shm`) when they appear.
- **R3 → Schema Migration Runner:**
  - Create table `schema_version (version INTEGER PRIMARY KEY, applied_at INTEGER NOT NULL)`.
  - Implement migration runner executing versioned SQL migration scripts in sequential order inside atomic `db.BeginTx()` database transactions.
  - Return migration errors and roll back if any migration step fails.
- **R4 → Baseline Schema Provisioning (Migration v1):**
  - Define Migration v1 creating:
    - `vault_metadata`: `(vault_id TEXT PRIMARY KEY, schema_version INTEGER NOT NULL, crypto_suite TEXT NOT NULL, kdf_params_json TEXT NOT NULL, wrapped_key BLOB NOT NULL, mck BLOB NOT NULL, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL)`
    - `records`: `(id TEXT PRIMARY KEY, vault_id TEXT NOT NULL, record_type TEXT NOT NULL, version INTEGER NOT NULL, payload BLOB NOT NULL, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL, deleted_at INTEGER)`
    - `record_history`: `(id TEXT PRIMARY KEY, record_id TEXT NOT NULL, vault_id TEXT NOT NULL, version INTEGER NOT NULL, payload BLOB NOT NULL, archived_at INTEGER NOT NULL, FOREIGN KEY(record_id) REFERENCES records(id) ON DELETE CASCADE)`
    - `tags`: `(id TEXT PRIMARY KEY, vault_id TEXT NOT NULL, name TEXT NOT NULL, created_at INTEGER NOT NULL, UNIQUE(vault_id, name))`
    - `record_tags`: `(record_id TEXT NOT NULL, tag_id TEXT NOT NULL, PRIMARY KEY(record_id, tag_id), FOREIGN KEY(record_id) REFERENCES records(id) ON DELETE CASCADE, FOREIGN KEY(tag_id) REFERENCES tags(id) ON DELETE CASCADE)`
    - Indexes on `records(vault_id, deleted_at)`, `record_history(record_id)`, `tags(vault_id)`.

## Architecture

- Layer: Infrastructure persistence layer (`internal/storage/sqlite`).
- Inward dependency flow: purely implements database connections and repositories, having zero presentation layer logic.
- Offline by architecture: Pure Go SQLite driver (`modernc.org/sqlite`), no external network or C toolchain requirements.

## Components Affected

- `go.mod` / `go.sum`: Add `modernc.org/sqlite`.
- `internal/storage/sqlite/db.go`: Connection manager, PRAGMA setup, and file permission checks.
- `internal/storage/sqlite/migrations.go`: Migration runner, schema version query, and migration registry.
- `internal/storage/sqlite/schema_v1.go`: Migration v1 SQL definitions.
- `internal/storage/sqlite/*_test.go`: Unit and integration tests for PRAGMAs, file permissions, migrations, and rollbacks.

## Data Changes

- Establishes the core SQLite physical schema for all GoVault tables.

## API Changes

```go
package sqlite

type DB struct {
    *sql.DB
    path string
}

func Open(path string) (*DB, error)
func (db *DB) Close() error
func (db *DB) Migrate() error
func (db *DB) CurrentVersion() (int, error)
```

## Integration Changes

- Ingested by `SPEC-005` (`vault_metadata` repository) and `SPEC-006` (`records`/`tags` repositories).

## Implementation Sequence

1. Add `modernc.org/sqlite` to `go.mod`.
2. Implement `db.go` with connection opening, file permission enforcement, and PRAGMA execution.
3. Implement `migrations.go` and `schema_v1.go` with version tracking and transactional execution.
4. Implement comprehensive unit and integration tests verifying PRAGMA settings, permission bits, table creations, and rollback handling.

## Test Strategy

- **PRAGMA Verification Test:** Query `PRAGMA journal_mode`, `PRAGMA foreign_keys`, `PRAGMA synchronous` and verify expected values.
- **Permission Verification Test:** On Linux/POSIX, create temp database and inspect `os.Stat(path).Mode().Perm()` asserting `0600`.
- **Migration & Rollback Test:** Run migrations to latest version; verify all tables and indices exist; verify that intentional SQL failure in a migration step rolls back cleanly.

## Risks

- *Pure Go SQLite Performance:* `modernc.org/sqlite` is slightly slower than CGO `mattn/go-sqlite3` on massive batch operations, but for a local password manager holding thousands of secrets, it offers superior portability, CGO-free compilation, and zero external C dependencies.

## Assumptions

- Operating filesystem supports standard file locking and POSIX permissions.
