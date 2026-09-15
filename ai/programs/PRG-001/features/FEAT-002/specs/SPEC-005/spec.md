---
id: SPEC-005
type: spec
status: ready
parent: FEAT-002
depends_on:
  - SPEC-004
---

# SPEC-005: Vault Metadata Repository & Lifecycle Persistence

## Intent

Specify the storage repository responsible for reading, writing, updating, and verifying the `vault_metadata` table, which holds the Vault UUID, schema version, cryptographic suite identifier, KDF parameters, and the Wrapped Vault Key.

## Requirements

- **R1: Metadata Record Schema:** The `vault_metadata` table must store `vault_id` (UUID text primary key), `schema_version` (integer), `crypto_suite` (text, e.g. `v1`), `kdf_params_json` (text), `wrapped_key` (blob), `mck` (blob), `created_at` (integer Unix timestamp), and `updated_at` (integer Unix timestamp).
- **R2: Single-Vault Constraint:** The repository must enforce that a database file contains exactly one active vault metadata entry.
- **R3: Atomic Metadata Updates on Rotation:** When a master password rotation occurs, the repository must update `kdf_params_json`, `wrapped_key`, `mck`, and `updated_at` atomically in a single transaction.
- **R4: Metadata Inspection:** Provide repository methods to inspect non-secret vault attributes (`vault_id`, `crypto_suite`, `created_at`, `updated_at`) without requiring master password unwrapping.

## Acceptance Scenarios

- **Scenario 1: Store & Retrieve Metadata**
  - *Given* an initialized `WrappedKeyEnvelope` and a new Vault UUID
  - *When* metadata is saved to the database
  - *Then* it can be queried back with identical envelope parameters and timestamps.
- **Scenario 2: Single-Vault Invariant**
  - *Given* a database with existing metadata
  - *When* attempting to insert a second metadata row with a different vault_id
  - *Then* the insert is rejected.
- **Scenario 3: Atomic Rotation Persistence**
  - *Given* an existing vault metadata row
  - *When* saving an updated `WrappedKeyEnvelope` from a password rotation
  - *Then* the metadata row is updated atomically and `updated_at` reflects the update time.

## Edge Cases

- Reading metadata from an uninitialized or empty database must return a specific `ErrVaultNotInitialized` error.

## Constraints

- Secrets (raw Master Password, unwrapped Vault Key) must never be passed to or stored in `vault_metadata`.

## Non-Goals

- Multi-vault hosting in a single SQLite file (each vault file is an independent database).
