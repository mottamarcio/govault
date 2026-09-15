---
id: SPEC-007
type: spec
status: ready
parent: FEAT-003
depends_on:
  - SPEC-001
  - SPEC-002
  - SPEC-003
  - SPEC-004
  - SPEC-005
  - SPEC-006
---

# SPEC-007: Domain Models, Value Objects & Secure Serialization

## Intent

Specify the core domain entities, value objects, secret payload structures (Login, SecureNote, APIKey, Custom, Field), validation rules, and deterministic binary/JSON serialization for encrypted payload envelopes.

## Requirements

- **R1: Core Domain Entities & Value Objects:** Define domain structs for `Vault`, `Record`, `Tag`, `HistoryEntry`, and custom field key-value pairs (`Field`), with zero external third-party dependencies.
- **R2: Typed Secret Payloads:** Implement typed secret payloads for standard secret types:
  - `LoginPayload`: username, password, URI, notes, custom fields.
  - `NotePayload`: title, content, custom fields.
  - `APIKeyPayload`: service, key, secret, endpoint, notes, custom fields.
  - `CustomPayload`: fields map, notes.
- **R3: Serialization & Envelope Marshaling:** Provide deterministic serialization/deserialization of secret payloads to/from raw bytes for payload encryption and decryption with AAD context verification.
- **R4: Domain Validation Rules:** Enforce validation rules on entity creation and modification (non-empty IDs, valid record types, max lengths, valid field names).

## Acceptance Scenarios

- **Scenario 1: Typed Payload Serialization Round-trip**
  - *Given* a `LoginPayload` with username, password, URI, and custom fields
  - *When* serialized to binary/JSON and deserialized back
  - *Then* all fields, notes, and custom key-values match the original exactly.
- **Scenario 2: Validation Failure on Invalid Payload**
  - *Given* a record with an invalid or unsupported record type or blank title
  - *When* validated
  - *Then* a domain validation error is returned indicating the invalid field.

## Edge Cases

- Payloads containing Unicode characters, emojis, multiline strings, or empty optional fields must serialize and deserialize without loss.
- Custom field names must be unique within a single secret payload.

## Constraints

- Domain models must reside in `internal/domain` with 0 external dependencies.
- Passwords and secret values within domain objects should provide zeroization or memory safety utilities where applicable.

## Non-Goals

- Database persistence logic or SQL schema definitions (handled in storage repositories).
