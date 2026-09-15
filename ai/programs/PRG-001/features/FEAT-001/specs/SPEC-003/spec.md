---
id: SPEC-003
type: spec
status: ready
parent: FEAT-001
depends_on:
  - SPEC-001
  - SPEC-002
---

# SPEC-003: Key Hierarchy Management, Vault Key Wrapping & Master Password Rotation

## Intent

Specify the key lifecycle, Vault Key (VK) generation, key wrapping via Key Encryption Key (KEK), and master password rotation workflows that allow changing master credentials without re-encrypting existing vault records.

## Requirements

- **R1: Vault Key Generation:** Upon vault creation, a cryptographically random 32-byte (256-bit) Vault Key (VK) must be generated using `crypto/rand`.
- **R2: Key Wrapping Envelope:** The Vault Key must be wrapped (encrypted) with the KEK derived from the master password (SPEC-001) using XChaCha20-Poly1305 (SPEC-002), bound to the vault's unique UUID via AAD (`govault/v1/vault_key:<vault_id>`).
- **R3: Master Password Rotation:** When the user changes their master password:
  - Derive a new Master Key and new KEK from the new password using a freshly generated salt.
  - Re-encrypt the existing un-wrapped Vault Key with the new KEK.
  - Update the stored Wrapped Vault Key and KDF parameters atomically.
  - Existing individual record ciphertexts encrypted with the Vault Key remain untouched and valid.
- **R4: Key Material Scrubbing:** Vault Key and intermediate KEK byte slices in memory must be explicitly overwritten/zeroized when locking the vault or completing rotation.

## Acceptance Scenarios

- **Scenario 1: Vault Initialization & Key Unwrapping**
  - *Given* a new vault master password
  - *When* initialization generates a Vault Key and wraps it with KEK
  - *Then* unlocking the vault with the same master password successfully unwraps and recovers the identical Vault Key.
- **Scenario 2: Master Password Rotation**
  - *Given* an initialized vault with several records encrypted under Vault Key `VK`
  - *When* the master password is changed from `PasswordOld` to `PasswordNew`
  - *Then* the Wrapped Vault Key is updated; unlocking with `PasswordNew` successfully unwraps `VK` and all existing records decrypt successfully without re-encryption.
- **Scenario 3: Old Password Invalidation**
  - *Given* a successfully rotated master password
  - *When* attempting to unlock using `PasswordOld`
  - *Then* unlocking fails and the Vault Key cannot be unwrapped.

## Edge Cases

- Concurrent rotation attempts or system interruptions during rotation must not leave the vault in an unrecoverable state (rotation updates must be atomic).
- Ensuring memory buffers holding the unencrypted Vault Key are wiped upon failure.

## Constraints

- The raw un-wrapped Vault Key must never be written to non-volatile storage.
- Key wrapping must always use authenticated encryption with distinct AAD.

## Non-Goals

- Multi-party key recovery or escrow systems.
