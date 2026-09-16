---
type: tasks
for: SPEC-014
---

# Tasks

## TASK-046 — CLI Session Helper & Record Creation Command (`govault add`)

- [x] Completed
- **Serves:** SPEC-014:R1
- **Depends on:** none
- **Files/Components:** `internal/cli/session_helper.go`, `internal/cli/record_add.go`, `internal/cli/root.go`, `internal/cli/record_test.go`
- **Verification:** Ran `go test -v -race ./internal/cli -run 'TestRecordAdd'` verifying interactive and flag-based secret creation across all supported types (`login`, `note`, `apikey`, `custom`), secure masked password prompt or auto-generation (`--generate` / `-g`), tag parsing (`--tag`), custom key-value field ingestion (`--field key=value`), JSON output, and database persistence. Evidence: `PASS TestRecordAdd`.

## TASK-047 — Secret Retrieval, Masking & Raw Pipeline Output (`govault get` / `govault show`)

- [x] Completed
- **Serves:** SPEC-014:R2
- **Depends on:** TASK-046
- **Files/Components:** `internal/cli/record_get.go`, `internal/cli/root.go`, `internal/cli/record_test.go`
- **Verification:** Ran `go test -v -race ./internal/cli -run 'TestRecordGet'` verifying secret retrieval by ID and title, default password masking (`••••••••••••`), plaintext display with explicit `--show` flag, single field extraction (`--field password`), raw unformatted output for shell pipelines (`--raw`), `--json` payload export, and error exit code 1 on missing records. Evidence: `PASS TestRecordGet`.

## TASK-048 — Tabular Overview & In-Memory Fuzzy Search (`govault list` & `govault search`)

- [x] Completed
- **Serves:** SPEC-014:R3
- **Depends on:** TASK-046
- **Files/Components:** `internal/cli/record_list.go`, `internal/cli/record_search.go`, `internal/cli/output.go`, `internal/cli/root.go`, `internal/cli/record_test.go`
- **Verification:** Ran `go test -v -race ./internal/cli -run 'TestRecordListAndSearch'` verifying POSIX tabular rendering of records (columns: `ID`, `Type`, `Title`, `Tags`, `Updated`), filtering by `--type` and `--tag`, in-memory fuzzy/prefix search relevance ranking across decrypted titles and tags via `govault search <query>`, and `--json` array outputs. Evidence: `PASS TestRecordListAndSearch`.

## TASK-049 — Record Modification, Deletion & Trash Management (`govault edit`, `govault delete`, `govault trash`)

- [x] Completed
- **Serves:** SPEC-014:R4
- **Depends on:** TASK-046, TASK-047
- **Files/Components:** `internal/cli/record_edit.go`, `internal/cli/record_delete.go`, `internal/cli/record_trash.go`, `internal/cli/root.go`, `internal/cli/record_test.go`
- **Verification:** Ran `go test -v -race ./internal/cli -run 'TestRecordEditDeleteTrash'` verifying updating record fields/tags (`govault edit`), soft deletion to trash (`govault delete`), permanent purge with `--permanent`, listing soft-deleted items (`govault trash list`), restoring items (`govault trash restore <id>`), and emptying the trash (`govault trash purge`). Evidence: `PASS TestRecordEditDeleteTrash`.
