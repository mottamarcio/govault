---
id: FEAT-003
type: feature
status: active
parent: PRG-001
---

# FEAT-003: Core Secrets Domain & Application Services

## Capability

Implements the domain entities (Vault, Record, Tag, History, Value Objects) and application use-case services orchestrating cryptography, storage, search, generation, and session state.

## User Value

Provides clean, business-logic-rich application services that both CLI and TUI consume identically, ensuring consistent behavior, validation rules, and error handling across all user interactions.

## Scope

- Pure Domain Models: Record (Login, SecureNote, APIKey, Custom), Tag, Category, HistoryEntry, VaultSession.
- Application Services:
  - `VaultService`: Init, open/unlock, lock, status, master password change, purge/wipe.
  - `RecordService`: Create, get, list, search, update, delete, restore, history inspection.
  - `PasswordGenService`: Password generation, Diceware passphrase generation, entropy calculation.
  - `ClipboardService`: Timed clipboard copying and automatic clearing.
- In-memory search and fuzzy filtering across decrypted record titles and tags.

## Non-Goals

- Direct terminal rendering or Cobra command flag parsing (delegated to presentation layers).

## Constraints

- Inward dependency flow: Domain has 0 third-party dependencies; Application Services depend only on Domain, Crypto interfaces, and Storage interfaces.
- Secrets must be sanitized and cleared from transient memory structures where possible.

## Relevant Knowledge

- [KNOW-001](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-001-product-requirements.md) — Product Requirements and CLI/TUI Capabilities
- [KNOW-006](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-006-architecture.md) — GoVault Software Architecture and Package Design

## Open Questions

- None.
