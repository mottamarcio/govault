---
type: tasks
for: SPEC-004
---

# Tasks

## TASK-014 — SQLite Driver Installation and Connection Manager with PRAGMAs

- [x] Completed
- **Serves:** SPEC-004:R1
- **Depends on:** none
- **Files/Components:** `go.mod`, `internal/storage/sqlite/db.go`, `internal/storage/sqlite/db_test.go`
- **Verification:** Ran `go test -v -run TestDBOpenPragmas ./internal/storage/sqlite` verifying WAL mode, foreign keys, synchronous normal, and busy timeout settings. Evidence: `PASS TestDBOpenPragmas`.

## TASK-015 — POSIX Database File and Directory Permissions Enforcement

- [x] Completed
- **Serves:** SPEC-004:R2
- **Depends on:** TASK-014
- **Files/Components:** `internal/storage/sqlite/permissions.go`, `internal/storage/sqlite/permissions_test.go`
- **Verification:** Ran `go test -v -run TestFilePermissions ./internal/storage/sqlite` verifying `0600` file permissions and `0700` parent directory permissions. Evidence: `PASS TestFilePermissions`.

## TASK-016 — Transactional Schema Migration Engine

- [x] Completed
- **Serves:** SPEC-004:R3
- **Depends on:** TASK-014
- **Files/Components:** `internal/storage/sqlite/migrations.go`, `internal/storage/sqlite/migrations_test.go`
- **Verification:** Ran `go test -v -run TestMigrations ./internal/storage/sqlite` verifying atomic execution, version tracking in `schema_version`, and rollback on error. Evidence: `PASS TestMigrations` and `PASS TestMigrationRollback`.

## TASK-017 — Migration v1 Baseline Schema Provisioning

- [x] Completed
- **Serves:** SPEC-004:R4
- **Depends on:** TASK-016
- **Files/Components:** `internal/storage/sqlite/schema_v1.go`, `internal/storage/sqlite/schema_v1_test.go`
- **Verification:** Ran `go test -v -race ./internal/storage/sqlite` verifying that all tables (`vault_metadata`, `records`, `record_history`, `tags`, `record_tags`) and indices are created cleanly. Evidence: `PASS TestSchemaV1TablesExist`.
