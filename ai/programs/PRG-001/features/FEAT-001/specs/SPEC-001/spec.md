---
id: SPEC-001
type: spec
status: ready
parent: FEAT-001
depends_on: []
---

# SPEC-001: Password Key Derivation (Argon2id) & Domain Separation (HKDF-SHA-256)

## Intent

Specify the cryptographic key derivation pipeline that converts a user's master password into a Master Key using Argon2id, and subsequently derives separate, purpose-specific keys (Key Encryption Key and Master Confirmation Key) using HKDF-SHA-256 with strict domain separation tags.

## Requirements

- **R1: Argon2id Parameter Compliance:** The KDF must execute Argon2id adhering to RFC 9106 baseline settings: 3 iterations (`time=3`), 64 MB memory (`memory=65536` KiB), 4 threads (`parallelism=4`), producing a 32-byte (256-bit) Master Key using a 32-byte CSPRNG salt.
- **R2: Domain-Separated Key Derivation (HKDF):** The system must derive sub-keys from the Master Key using HKDF-SHA-256 with unique context info strings:
  - `govault/v1/kek`: Derives the 32-byte Key Encryption Key (KEK) used to wrap/unwrap the Vault Key.
  - `govault/v1/mck`: Derives the 32-byte Master Confirmation Key (MCK) used to verify master password correctness without persisting master passwords or exposing the KEK.
- **R3: Constant-Time Verification:** Password confirmation checks comparing derived MCK or authentication tags must execute using constant-time comparison (`crypto/subtle.ConstantTimeCompare`).
- **R4: Parameter Serialization:** KDF parameters (salt, time, memory, threads) must be serializable and deserializable for persistence and backup headers.

## Acceptance Scenarios

- **Scenario 1: Deterministic Key Derivation**
  - *Given* a fixed master password, fixed 32-byte salt, and fixed Argon2id parameters
  - *When* key derivation is executed
  - *Then* it produces the exact same 32-byte Master Key, KEK, and MCK matching pre-calculated KAT (Known Answer Test) vectors.
- **Scenario 2: Distinct Keys per Purpose**
  - *Given* a valid derived Master Key
  - *When* KEK and MCK are derived via HKDF
  - *Then* KEK and MCK are cryptographically distinct from each other and from the Master Key.
- **Scenario 3: Password Rejection**
  - *Given* an initialized vault with an existing MCK
  - *When* an incorrect master password is provided
  - *Then* the derived MCK does not match the stored confirmation tag and verification returns an authorization error.

## Edge Cases

- Empty or single-character master password inputs must be handled securely (either rejected or processed cleanly by KDF without crashing).
- Unicode master passwords must be normalized consistently (e.g. UTF-8 byte representation).

## Constraints

- Strictly offline: relies solely on Go standard library and `golang.org/x/crypto/argon2` + `golang.org/x/crypto/hkdf`.
- Master password byte slices should be zeroized/cleared from memory as soon as derivation completes.

## Non-Goals

- Password complexity policy enforcement (handled at the UI/CLI layer).
- Alternative KDF algorithms (e.g. PBKDF2 or scrypt).
