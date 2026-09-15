---
id: SPEC-008
type: spec
status: ready
parent: FEAT-003
depends_on:
  - SPEC-001
  - SPEC-002
  - SPEC-003
  - SPEC-005
  - SPEC-007
---

# SPEC-008: Vault Lifecycle Application Service & Session Management

## Intent

Specify the `VaultService` orchestrating vault lifecycle use cases: initialization, opening/unlocking with master password, active in-memory session holding the decrypted `VaultKey`, locking/closing with memory zeroization, master password rotation, and vault wipe/purge.

## Requirements

- **R1: Vault Initialization:** Orchestrate vault initialization use case: generate new random salt and vault ID, derive `MasterKey` via Argon2id, compute MCK, generate 32-byte random `VaultKey`, wrap `VaultKey` with `MasterKey`, and persist metadata through `MetadataRepository`.
- **R2: Unlock & Active Session State:** Provide `Unlock(password)` verifying MCK and unwrapping `VaultKey`. Store unlocked state in an in-memory `Session` object providing thread-safe access to `VaultKey` while unlocked.
- **R3: Lock & Secure Teardown:** Provide `Lock()` method clearing and zeroizing the active `VaultKey` from memory and resetting session state to locked.
- **R4: Master Password Rotation:** Provide `ChangeMasterPassword(oldPassword, newPassword)` re-encrypting the existing `VaultKey` under a newly derived master key with fresh salt and updated MCK atomically in storage without re-encrypting all secrets.
- **R5: Vault Purge:** Provide `Purge()` completely wiping the vault database file and resetting state.

## Acceptance Scenarios

- **Scenario 1: Vault Initialization and Unlock**
  - *Given* a new vault configuration with a master password
  - *When* `Init` is executed followed by `Unlock`
  - *Then* the vault metadata is stored and an active unlocked session with valid `VaultKey` is established.
- **Scenario 2: Vault Lock Memory Zeroization**
  - *Given* an unlocked vault session
  - *When* `Lock` is called
  - *Then* the session is marked locked and previous key memory is wiped.
- **Scenario 3: Master Password Change**
  - *Given* an initialized vault with master password "oldPass"
  - *When* `ChangeMasterPassword("oldPass", "newPass")` is called
  - *Then* unlocking with "oldPass" fails with invalid password error, while unlocking with "newPass" succeeds and unwraps the same `VaultKey`.

## Edge Cases

- Attempting to perform operations requiring an active session (e.g. locking an already locked vault, or reading keys from a locked session) must return `ErrVaultLocked`.
- Supplying an incorrect master password during unlock or password change must return `ErrInvalidPassword` without exposing internal error traces.

## Constraints

- Active master passwords and derived master keys must never be stored persistently.
- Unlocked `VaultKey` in session memory must be wiped upon `Lock()`.

## Non-Goals

- OS keyring storage or daemon background IPC (handled in session/CLI integration if needed).
