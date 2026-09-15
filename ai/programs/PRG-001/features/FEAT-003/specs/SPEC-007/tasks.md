---
type: tasks
for: SPEC-007
---

# Tasks

## TASK-025 — Domain Types, Value Objects & Field Structures

- [x] Completed
- **Serves:** SPEC-007:R1, SPEC-007:R2
- **Depends on:** none
- **Files/Components:** `internal/domain/errors.go`, `internal/domain/record_type.go`, `internal/domain/field.go`, `internal/domain/payloads.go`, `internal/domain/record.go`, `internal/domain/tag.go`, `internal/domain/history.go`
- **Verification:** Ran `go test -v ./internal/domain` ensuring types and structs compile cleanly without external dependencies. Evidence: compiled with zero errors.

## TASK-026 — Deterministic Payload Serialization & Deserialization

- [x] Completed
- **Serves:** SPEC-007:R3
- **Depends on:** TASK-025
- **Files/Components:** `internal/domain/serialization.go`, `internal/domain/serialization_test.go`
- **Verification:** Ran `go test -v -run TestPayloadSerialization ./internal/domain` validating exact round-trip serialization and deserialization for Login, Note, APIKey, and Custom payloads with Unicode, emojis, multiline strings, and custom fields. Evidence: `PASS TestPayloadSerialization`.

## TASK-027 — Domain Validation Rules and Model Constraints

- [x] Completed
- **Serves:** SPEC-007:R4
- **Depends on:** TASK-025
- **Files/Components:** `internal/domain/validation.go`, `internal/domain/validation_test.go`
- **Verification:** Ran `go test -v -race ./internal/domain/...` asserting validation error rules on missing IDs, invalid record types, blank titles, duplicate custom field keys, and excessive string lengths. Evidence: `PASS TestRecordValidation`, `PASS TestPayloadValidation`, `PASS TestTagValidation`.
