---
type: plan
for: SPEC-003
status: ready
---

# Implementation Plan: Key Hierarchy Management, Vault Key Wrapping & Master Password Rotation

## Summary

Design and implement the key hierarchy manager (`internal/crypto/keys`) responsible for generating the master Vault Key (VK), wrapping and unwrapping the Vault Key using the Key Encryption Key (KEK) and XChaCha20-Poly1305 with context AAD, executing master password rotations atomically, and performing memory scrubbing of sensitive cryptographic keys.

## Repository Context

`SPEC-001` (Argon2id KDF & HKDF domain separation in `internal/crypto/kdf`) and `SPEC-002` (XChaCha20-Poly1305 AEAD in `internal/crypto/cipher`) are implemented, tested, and validated. The key hierarchy package will reside at `internal/crypto/keys`, orchestrating `kdf` and `cipher` to provide the complete key lifecycle for GoVault.

## Requirement Coverage

- **R1 → Vault Key Generation:**
  - Implement `GenerateVaultKey() ([]byte, error)` producing a 32-byte (256-bit) cryptographically random Vault Key using `crypto/rand.Reader`.
- **R2 → Key Wrapping Envelope:**
  - Define `WrappedKeyEnvelope` struct: `KDFParams *kdf.Argon2Params`, `WrappedKey []byte` (packed XChaCha20-Poly1305 ciphertext including nonce & tag), and `MCK []byte` (Master Confirmation Key tag).
  - Implement `WrapVaultKey(vaultKey []byte, password []byte, vaultID string, params *kdf.Argon2Params) (*WrappedKeyEnvelope, error)`:
    - Derives Master Key from password + params.
    - Derives KEK and MCK via HKDF.
    - Encrypts `vaultKey` with KEK using AAD `govault/v1/vault_key:<vault_id>`.
    - Returns `WrappedKeyEnvelope`.
  - Implement `UnwrapVaultKey(env *WrappedKeyEnvelope, password []byte, vaultID string) ([]byte, error)`:
    - Derives candidate Master Key and candidate MCK.
    - Verives candidate MCK against `env.MCK` using `kdf.VerifyMCK`.
    - If valid, derives KEK and decrypts `env.WrappedKey` with AAD.
- **R3 → Master Password Rotation:**
  - Implement `RotateMasterPassword(env *WrappedKeyEnvelope, oldPassword, newPassword []byte, vaultID string, newParams *kdf.Argon2Params) (*WrappedKeyEnvelope, error)`:
    - Unwraps existing `vaultKey` using `oldPassword`.
    - Generates new KDF salt/params if `newParams` is nil.
    - Re-wraps `vaultKey` with `newPassword`.
    - Zeroizes intermediate plaintext keys.
    - Validates that the underlying `vaultKey` remains identical while `WrappedKey`, salt, and `MCK` are updated.
- **R4 → Key Material Scrubbing:**
  - Integrate `kdf.Zeroize` across all key derivation buffers, intermediate KEK/MasterKey slices, and unwrapped VaultKey copies upon session cleanup or rotation completion.

## Architecture

- Pure Go package at `internal/crypto/keys`.
- Integrates `internal/crypto/kdf` and `internal/crypto/cipher`.
- Exposes clean, high-level structs for downstream storage entities (`vault_metadata`) and application session services.

## Components Affected

- `internal/crypto/keys/keys.go`: Vault Key generation, wrapping, unwrapping, and rotation.
- `internal/crypto/keys/envelope.go`: `WrappedKeyEnvelope` JSON/binary serialization helpers.
- `internal/crypto/keys/*_test.go`: Unit, integration, rotation, and tamper tests.

## Data Changes

- Standard `WrappedKeyEnvelope` structure for database storage and backup packaging.

## API Changes

```go
package keys

type WrappedKeyEnvelope struct {
    KDFParams  *kdf.Argon2Params `json:"kdf_params"`
    WrappedKey []byte            `json:"wrapped_key"`
    MCK        []byte            `json:"mck"`
}

func GenerateVaultKey() ([]byte, error)
func WrapVaultKey(vaultKey []byte, password []byte, vaultID string, params *kdf.Argon2Params) (*WrappedKeyEnvelope, error)
func UnwrapVaultKey(env *WrappedKeyEnvelope, password []byte, vaultID string) ([]byte, error)
func RotateMasterPassword(env *WrappedKeyEnvelope, oldPassword, newPassword []byte, vaultID string, newParams *kdf.Argon2Params) (*WrappedKeyEnvelope, error)
```

## Integration Changes

- Directly ingested by `FEAT-002` (`internal/storage/sqlite` for storing and reading `vault_metadata`) and `FEAT-003` (`internal/service/vault_service.go` for unlocking, locking, and password changes).

## Implementation Sequence

1. Implement `keys.go` with `GenerateVaultKey`, `WrapVaultKey`, `UnwrapVaultKey`, and `RotateMasterPassword`.
2. Implement `envelope.go` for `WrappedKeyEnvelope` marshaling/unmarshaling.
3. Implement `keys_test.go` verifying generation, wrap/unwrap lifecycle, incorrect password rejection, and rotation preservation of the Vault Key.

## Test Strategy

- **Lifecycle Test:** Generate VK -> Wrap -> Unwrap -> assert VK equality.
- **Password Rotation Test:** Encrypt record with VK; Rotate password; Unwrap VK with new password; assert record still decrypts cleanly.
- **Invalid Password / Tampering:** Assert that incorrect password returns explicit auth error; tampered AAD (wrong vaultID) returns auth error.
- **Memory Scrubbing:** Verify intermediate keys are zeroized upon completion.

## Risks

- None.

## Assumptions

- Vault UUIDs are immutable for the lifetime of the vault instance.
