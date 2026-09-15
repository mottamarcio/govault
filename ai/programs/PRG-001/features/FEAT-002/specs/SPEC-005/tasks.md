---
type: tasks
for: SPEC-005
---

# Tasks

## TASK-018 — Vault Metadata Entity Definition and Repository Scaffolding

- [x] Completed
- **Serves:** SPEC-005:R1
- **Depends on:** none
- **Files/Components:** `internal/storage/sqlite/metadata_repo.go`, `internal/storage/sqlite/metadata_repo_test.go`
- **Verification:** Ran `go test -v -run TestMetadataCreateGet ./internal/storage/sqlite` verifying metadata entity mapping and persistence. Evidence: `PASS TestMetadataCreateGet`.

## TASK-019 — Single-Vault Constraint Enforcement

- [x] Completed
- **Serves:** SPEC-005:R2
- **Depends on:** TASK-018
- **Files/Components:** `internal/storage/sqlite/metadata_repo.go`, `internal/storage/sqlite/metadata_repo_test.go`
- **Verification:** Ran `go test -v -run TestSingleVaultConstraint ./internal/storage/sqlite` verifying rejection of duplicate vault initialization. Evidence: `PASS TestSingleVaultConstraint`.

## TASK-020 — Atomic Metadata Update during Master Password Rotation

- [x] Completed
- **Serves:** SPEC-005:R3
- **Depends on:** TASK-018
- **Files/Components:** `internal/storage/sqlite/metadata_repo.go`, `internal/storage/sqlite/metadata_repo_test.go`
- **Verification:** Ran `go test -v -run TestMetadataUpdateEnvelope ./internal/storage/sqlite` verifying atomic update of KDF params, wrapped key, MCK, and updated_at. Evidence: `PASS TestMetadataUpdateEnvelope`.

## TASK-021 — Non-Secret Metadata Inspection Method

- [x] Completed
- **Serves:** SPEC-005:R4
- **Depends on:** TASK-018
- **Files/Components:** `internal/storage/sqlite/metadata_repo.go`, `internal/storage/sqlite/metadata_repo_test.go`
- **Verification:** Ran `go test -v -race ./internal/storage/sqlite` verifying `Inspect` returns header summary info without errors. Evidence: `PASS TestMetadataInspect`.
