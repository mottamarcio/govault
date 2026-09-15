---
type: plan
for: SPEC-001
status: ready
---

# Implementation Plan: Password Key Derivation & Domain Separation

## Summary

Design and implement the foundational cryptographic key derivation package (`internal/crypto/kdf`) responsible for deriving high-entropy Master Keys from user passwords via Argon2id (RFC 9106 baseline parameters) and deriving purpose-specific sub-keys (KEK and MCK) via HKDF-SHA-256 with explicit domain separation.

## Repository Context

The repository is currently initialized with specifications and raw architecture documentation. No Go module or source code exists yet. This implementation will establish the Go module `github.com/mottamarcio/govault`, add standard cryptographic dependencies (`golang.org/x/crypto`), and build the pure `internal/crypto/kdf` package.

## Requirement Coverage

- **R1 → Argon2id Parameter Compliance:**
  - Define `Argon2Params` struct holding `Time` (uint32 = 3), `Memory` (uint32 = 64 * 1024 KiB), `Threads` (uint8 = 4), `KeyLen` (uint32 = 32), and `Salt` ([]byte = 32 bytes).
  - Implement `DeriveMasterKey(password []byte, params Argon2Params) ([]byte, error)` invoking `argon2.IDKey(password, salt, time, memory, threads, keyLen)`.
  - Provide helper `DefaultArgon2Params()` generating fresh 32-byte CSPRNG salt via `crypto/rand`.
- **R2 → Domain-Separated Key Derivation (HKDF):**
  - Implement `DeriveSubKey(masterKey []byte, info string, length int) ([]byte, error)` using `hkdf.New(sha256.New, masterKey, nil, []byte(info))`.
  - Define constants for context info strings: `InfoKEK = "govault/v1/kek"` and `InfoMCK = "govault/v1/mck"`.
  - Provide typed convenience helpers `DeriveKEK(masterKey []byte)` and `DeriveMCK(masterKey []byte)` returning 32-byte keys.
- **R3 → Constant-Time Verification:**
  - Implement `VerifyMCK(derivedMCK, storedMCK []byte) bool` utilizing `subtle.ConstantTimeCompare(derivedMCK, storedMCK) == 1`.
- **R4 → Parameter Serialization:**
  - Provide JSON and binary marshal/unmarshal methods for `Argon2Params` ensuring safe serialization for SQLite storage and backup headers.

## Architecture

- Pure, inward Go package located at `internal/crypto/kdf`.
- Zero network dependencies, relying only on Go standard library (`crypto/rand`, `crypto/subtle`, `crypto/sha256`, `io`) and `golang.org/x/crypto/argon2` + `golang.org/x/crypto/hkdf`.
- Defensive memory handling: input master passwords and sensitive intermediate key slices will provide scrubbing utilities (e.g. `Memclr` / `Zeroize`).

## Components Affected

- `go.mod` / `go.sum`: Initialize module and track `golang.org/x/crypto`.
- `internal/crypto/kdf/params.go`: Structs, default constants, validation, and serialization.
- `internal/crypto/kdf/argon2.go`: Argon2id key derivation logic.
- `internal/crypto/kdf/hkdf.go`: HKDF sub-key derivation and domain separation constants.
- `internal/crypto/kdf/verify.go`: Constant-time verification functions.
- `internal/crypto/kdf/mem.go`: Secure byte zeroization helpers.
- `internal/crypto/kdf/*_test.go`: Unit tests and KAT test suite.

## Data Changes

- Defines standard in-memory and serialized representations for `Argon2Params` used in downstream storage tables (`vault_metadata`) and backup headers.

## API Changes

```go
package kdf

type Argon2Params struct {
    Time        uint32 `json:"time"`
    Memory      uint32 `json:"memory"`
    Threads     uint8  `json:"threads"`
    KeyLen      uint32 `json:"key_len"`
    Salt        []byte `json:"salt"`
}

func DefaultArgon2Params() (*Argon2Params, error)
func (p *Argon2Params) Validate() error
func DeriveMasterKey(password []byte, params *Argon2Params) ([]byte, error)
func DeriveKEK(masterKey []byte) ([]byte, error)
func DeriveMCK(masterKey []byte) ([]byte, error)
func VerifyMCK(candidateMCK, expectedMCK []byte) bool
func Zeroize(b []byte)
```

## Integration Changes

- Serves as the direct foundation for `SPEC-002` (ciphers) and `SPEC-003` (Vault Key wrapping).

## Implementation Sequence

1. Initialize Go module `github.com/mottamarcio/govault` and install `golang.org/x/crypto`.
2. Implement memory zeroization helper `mem.go` and its unit tests.
3. Implement `params.go` with parameter struct, defaults, validation rules, and JSON serialization tests.
4. Implement `argon2.go` with `DeriveMasterKey` and parameter validation.
5. Implement `hkdf.go` with domain separation strings and sub-key derivation.
6. Implement `verify.go` with constant-time comparison.
7. Implement comprehensive unit tests and RFC KAT test vectors in `kdf_test.go`.

## Test Strategy

- **Known Answer Tests (KAT):** Validate Argon2id and HKDF derivation outputs against deterministic test vectors.
- **RFC Parameter Validation:** Verify that invalid or out-of-range parameters (e.g., empty salt, 0 iterations, insufficient memory) return explicit errors.
- **Domain Separation Independence:** Assert that `DeriveKEK` and `DeriveMCK` for identical Master Keys produce completely distinct, uncorrelated 32-byte arrays.
- **Timing & Verification Tests:** Verify constant-time comparison against matching and tampered byte slices.
- **Zeroization Tests:** Verify that `Zeroize` overwrites byte buffers with zeros.

## Risks

- *Argon2id Memory/CPU Impact in Tests:* Running 64MB Argon2id in unit tests can slow down test suites. Mitigation: Use standard baseline parameters for integration/production tests, and support configurable lightweight parameters specifically for fast unit testing.

## Assumptions

- Target environments have at least 64MB of available RAM for key derivation during vault unlock operations.
