---
type: plan
for: SPEC-007
status: ready
---

# Implementation Plan: Domain Models, Value Objects & Secure Serialization

## Summary

Design and implement the core domain entities, value objects, typed secret payloads (`LoginPayload`, `NotePayload`, `APIKeyPayload`, `CustomPayload`), custom key-value field structures (`Field`), entity validation rules, and secure deterministic JSON/binary serialization under `internal/domain`.

## Repository Context

`FEAT-001` (crypto primitives) and `FEAT-002` (SQLite storage layer) are implemented and verified. The `internal/domain` package will serve as the innermost core of the architecture, depending strictly on the standard library with zero external dependencies. The storage repositories in `internal/storage/sqlite` and application services in `internal/service` will map to/from these domain models.

## Requirement Coverage

- **R1 → Core Domain Entities & Value Objects:**
  - Define domain entities and value objects in `internal/domain/`:
    - `RecordType` (enum/string: `login`, `note`, `api_key`, `custom`).
    - `Field` struct: `Key` (string), `Value` (string), `Masked` (bool).
    - `Record` struct: `ID`, `VaultID`, `Title`, `RecordType`, `Payload` (typed or byte slice), `Tags` ([]string), `Version`, `CreatedAt`, `UpdatedAt`, `DeletedAt`.
    - `Tag` struct: `ID`, `VaultID`, `Name`, `CreatedAt`.
    - `HistoryEntry` struct: `ID`, `RecordID`, `VaultID`, `Version`, `Payload`, `ArchivedAt`.
- **R2 → Typed Secret Payloads:**
  - Implement concrete typed payload models:
    - `LoginPayload`: `Username`, `Password`, `URI`, `Notes`, `CustomFields` ([]Field).
    - `NotePayload`: `Content`, `CustomFields` ([]Field).
    - `APIKeyPayload`: `Service`, `Key`, `Secret`, `Endpoint`, `Notes`, `CustomFields` ([]Field).
    - `CustomPayload`: `Notes`, `Fields` ([]Field).
- **R3 → Serialization & Envelope Marshaling:**
  - Implement deterministic serialization (`SerializePayload`, `DeserializePayload`) using canonical JSON encoding.
  - Implement helper methods to convert between typed payload structs and raw byte payloads expected by cipher envelopes.
  - Implement payload zeroization helper where memory scrubbing of sensitive plaintext fields is needed.
- **R4 → Domain Validation Rules:**
  - Implement `Validate()` methods on `Record`, `LoginPayload`, `NotePayload`, `APIKeyPayload`, `CustomPayload`, `Field`, `Tag`.
  - Validate non-empty required fields (e.g. `RecordType`, `Title`), max string lengths (e.g., Title <= 255 chars), uniqueness of custom field keys, and valid UUID identifiers.

## Architecture

```
internal/domain/
├── record_type.go     # RecordType enum and constants
├── field.go           # Custom Field key-value and masking
├── payloads.go        # Typed payloads (Login, Note, APIKey, Custom)
├── record.go          # Record entity and aggregate methods
├── tag.go             # Tag entity
├── history.go         # HistoryEntry entity
├── serialization.go   # Canonical payload serialization and deserialization
├── validation.go      # Domain validation logic and error types
└── errors.go          # Domain-specific sentinel errors
```

- Inward dependency flow: `internal/domain` has 0 third-party dependencies and does not import `internal/crypto` or `internal/storage`.

## Components Affected

- `internal/domain/errors.go`
- `internal/domain/record_type.go`
- `internal/domain/field.go`
- `internal/domain/payloads.go`
- `internal/domain/record.go`
- `internal/domain/tag.go`
- `internal/domain/history.go`
- `internal/domain/serialization.go`
- `internal/domain/validation.go`
- `internal/domain/domain_test.go`

## Data Changes

- None (database schema already defined in `FEAT-002`).

## API Changes

```go
package domain

type RecordType string

const (
    RecordTypeLogin  RecordType = "login"
    RecordTypeNote   RecordType = "note"
    RecordTypeAPIKey RecordType = "api_key"
    RecordTypeCustom RecordType = "custom"
)

type Field struct {
    Key    string `json:"key"`
    Value  string `json:"value"`
    Masked bool   `json:"masked"`
}

type LoginPayload struct {
    Username     string  `json:"username"`
    Password     string  `json:"password"`
    URI          string  `json:"uri,omitempty"`
    Notes        string  `json:"notes,omitempty"`
    CustomFields []Field `json:"custom_fields,omitempty"`
}

type NotePayload struct {
    Content      string  `json:"content"`
    CustomFields []Field `json:"custom_fields,omitempty"`
}

type APIKeyPayload struct {
    Service      string  `json:"service"`
    Key          string  `json:"key"`
    Secret       string  `json:"secret"`
    Endpoint     string  `json:"endpoint,omitempty"`
    Notes        string  `json:"notes,omitempty"`
    CustomFields []Field `json:"custom_fields,omitempty"`
}

type CustomPayload struct {
    Notes  string  `json:"notes,omitempty"`
    Fields []Field `json:"fields,omitempty"`
}

type Record struct {
    ID        string
    VaultID   string
    Title     string
    Type      RecordType
    Tags      []string
    Version   uint32
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt *time.Time
}

func SerializePayload(payload any) ([]byte, error)
func DeserializePayload(recordType RecordType, data []byte) (any, error)
```

## Integration Changes

- Ingested by `internal/service/record_service.go` in `SPEC-009` and mapper layers in CLI/TUI.

## Implementation Sequence

1. Define domain errors (`ErrInvalidRecordType`, `ErrEmptyTitle`, `ErrDuplicateFieldKey`, `ErrInvalidUUID`, etc.).
2. Define `RecordType`, `Field`, and typed payload structs (`LoginPayload`, `NotePayload`, `APIKeyPayload`, `CustomPayload`).
3. Define `Record`, `Tag`, and `HistoryEntry` entities.
4. Implement payload serialization and deserialization functions.
5. Implement validation methods across all domain models.
6. Write thorough unit tests covering round-trip serialization, edge cases (Unicode, multiline, emojis, empty optional fields), and validation rules.

## Test Strategy

- **Payload Serialization Tests:** Test JSON round-trip serialization/deserialization for each payload type.
- **Validation Tests:** Table-driven tests validating valid/invalid titles, UUIDs, field names, duplicate keys, and unsupported record types.
- **Edge Cases Tests:** Test handling of Unicode characters, multiline text, large custom fields, and empty slices.

## Risks

- None.

## Assumptions

- Standard Go `encoding/json` provides sufficient determinism and flexibility for typed secret payloads without external dependencies.
