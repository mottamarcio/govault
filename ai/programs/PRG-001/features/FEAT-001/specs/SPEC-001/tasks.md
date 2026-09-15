---
type: tasks
for: SPEC-001
---

# Tasks

## TASK-001 — Go Module Initialization and Secure Memory Helpers

- [x] Completed
- **Serves:** SPEC-001:R1
- **Depends on:** none
- **Files/Components:** `go.mod`, `internal/crypto/kdf/mem.go`, `internal/crypto/kdf/mem_test.go`
- **Verification:** Ran `go test -v ./internal/crypto/kdf` verifying `Zeroize` overwrites byte buffers cleanly. Evidence: All unit tests passed cleanly.

## TASK-002 — Argon2id Parameter Definition, Validation, and Serialization

- [x] Completed
- **Serves:** SPEC-001:R1, SPEC-001:R4
- **Depends on:** TASK-001
- **Files/Components:** `internal/crypto/kdf/params.go`, `internal/crypto/kdf/params_test.go`
- **Verification:** Ran `go test -v -run TestArgon2Params ./internal/crypto/kdf` verifying default RFC 9106 values, bounds validation, and JSON round-trip serialization. Evidence: `PASS TestArgon2ParamsValidation` and `PASS TestArgon2ParamsSerialization`.

## TASK-003 — Argon2id Master Key Derivation

- [x] Completed
- **Serves:** SPEC-001:R1
- **Depends on:** TASK-002
- **Files/Components:** `internal/crypto/kdf/argon2.go`, `internal/crypto/kdf/argon2_test.go`
- **Verification:** Ran `go test -v -run TestDeriveMasterKey ./internal/crypto/kdf` verifying deterministic Argon2id output and known test vectors. Evidence: `PASS TestDeriveMasterKey`.

## TASK-004 — HKDF Sub-Key Derivation and Domain Separation

- [x] Completed
- **Serves:** SPEC-001:R2
- **Depends on:** TASK-003
- **Files/Components:** `internal/crypto/kdf/hkdf.go`, `internal/crypto/kdf/hkdf_test.go`
- **Verification:** Ran `go test -v -run TestDeriveSubKeys ./internal/crypto/kdf` verifying distinct KEK (`govault/v1/kek`) and MCK (`govault/v1/mck`) outputs. Evidence: `PASS TestDeriveSubKeys`.

## TASK-005 — Constant-Time MCK Verification and KAT Test Suite

- [x] Completed
- **Serves:** SPEC-001:R3
- **Depends on:** TASK-004
- **Files/Components:** `internal/crypto/kdf/verify.go`, `internal/crypto/kdf/kdf_test.go`
- **Verification:** Ran `go test -v -race ./internal/crypto/kdf` verifying constant-time comparison, password rejection on tampered MCK, and full KAT test suite. Evidence: `PASS TestVerifyMCK` and `PASS TestKDFEndToEndPipeline` with 0 race warnings.
