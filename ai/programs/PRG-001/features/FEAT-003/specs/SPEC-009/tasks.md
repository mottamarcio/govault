---
type: tasks
for: SPEC-009
---

# Tasks

## TASK-031 — In-Memory Record Filter and Relevance Ranking

- [x] Completed
- **Serves:** SPEC-009:R5
- **Depends on:** none
- **Files/Components:** `internal/service/search.go`, `internal/service/search_test.go`
- **Verification:** Ran `go test -v -run TestSearchFilter ./internal/service` asserting substring matching, fuzzy tag matching, case-insensitivity, and relevance sorting across decrypted attributes. Evidence: `PASS TestSearchFilter`.

## TASK-032 — Record Encryption, Decryption and CRUD Orchestration

- [x] Completed
- **Serves:** SPEC-009:R1, SPEC-009:R2
- **Depends on:** TASK-031
- **Files/Components:** `internal/service/record_service.go`, `internal/service/record_service_test.go`
- **Verification:** Ran `go test -v -run TestRecordServiceCRUD ./internal/service` asserting end-to-end creation, subkey derivation, AAD encapsulation, retrieval, and decryption for Login, Note, APIKey, and Custom secrets. Evidence: `PASS TestRecordServiceCRUD`.

## TASK-033 — Revision History Traversal, Soft Delete, and Trash Purge

- [x] Completed
- **Serves:** SPEC-009:R3, SPEC-009:R4
- **Depends on:** TASK-032
- **Files/Components:** `internal/service/record_service.go`, `internal/service/record_service_test.go`
- **Verification:** Ran `go test -v -race -run TestRecordServiceHistoryAndTrash ./internal/service` asserting revision archiving on update, historical revision retrieval, soft deletion, trash listing, restore, and permanent purge. Evidence: `PASS TestRecordServiceHistoryAndTrash`.
