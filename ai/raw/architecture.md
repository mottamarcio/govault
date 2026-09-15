# GoVault Software Architecture

**Document:** `docs/architecture.md`
**Project:** GoVault
**Status:** Draft / Pre-v1 Architecture Specification
**Applies to:** GoVault v0.1+
**Last Updated:** September 2026

---

# 1. Purpose

This document defines the software architecture of GoVault.

It translates the requirements and security specifications defined in:

```text
docs/threat-model.md
docs/cryptography.md
docs/vault-format.md
docs/backup-format.md
```

into a concrete Go application architecture.

This specification defines:

* Repository structure.
* Package responsibilities.
* Dependency direction.
* Domain models.
* Application services.
* Cryptographic boundaries.
* Persistence boundaries.
* CLI architecture.
* TUI architecture.
* Backup architecture.
* Session lifecycle.
* Lock and unlock lifecycle.
* Clipboard handling.
* Configuration.
* Error handling.
* Dependency injection.
* Testing strategy.
* Offline enforcement.
* Startup and shutdown behavior.

The primary architectural objective is:

> Security-sensitive infrastructure should be isolated behind narrow interfaces, while the CLI and TUI remain presentation layers over the same application core.

---

# 2. Architectural Principles

GoVault follows several core principles.

## 2.1 Offline by Architecture

GoVault does not merely avoid network features.

The application should be designed so that its core has no reason to possess network capabilities.

The production dependency graph should avoid networking packages unless a future specification explicitly introduces them.

---

## 2.2 Dependency Direction

Dependencies flow inward toward domain and application logic.

Conceptually:

```text
CLI ───────────┐
               │
TUI ───────────┼──► Application Services
               │          │
               │          ▼
               │        Domain
               │
               ├──► Crypto
               │
               ├──► Storage
               │
               ├──► Backup
               │
               ├──► Clipboard
               │
               └──► Configuration
```

Presentation layers must not contain cryptographic or database business logic.

---

## 2.3 One Core, Multiple Interfaces

The Cobra CLI and Bubble Tea TUI use the same application services.

For example:

```text
govault copy github
```

and:

```text
TUI
→ GitHub
→ Copy Password
```

must ultimately invoke the same application operation.

---

# 3. Technology Stack

The initial GoVault stack is:

```text
Language
Go

CLI
Cobra

TUI
Bubble Tea
Lip Gloss

Persistence
SQLite

Cryptography
golang.org/x/crypto

Password KDF
Argon2id

Key Derivation
HKDF-SHA-256

Authenticated Encryption
XChaCha20-Poly1305

Randomness
crypto/rand
```

The exact SQLite Go driver must be selected separately.

Preference should be given to a driver that minimizes deployment complexity and unnecessary dependencies.

---

# 4. High-Level Architecture

GoVault is divided into six conceptual layers:

```text
┌─────────────────────────────────────┐
│          Presentation               │
│                                     │
│       Cobra CLI    Bubble Tea       │
└──────────────────┬──────────────────┘
                   │
┌──────────────────▼──────────────────┐
│           Application               │
│                                     │
│ Commands / Use Cases / Services     │
└──────────────────┬──────────────────┘
                   │
┌──────────────────▼──────────────────┐
│              Domain                 │
│                                     │
│ Entries / IDs / Policies / Models   │
└──────────────────┬──────────────────┘
                   │
        ┌──────────┼──────────┐
        ▼          ▼          ▼
     Crypto     Storage     Backup
        │          │          │
        └──────────┼──────────┘
                   │
              OS Adapters
```

---

# 5. Proposed Repository Structure

Recommended initial structure:

```text
govault/
├── cmd/
│   └── govault/
│       └── main.go
│
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   ├── session.go
│   │   ├── vault_service.go
│   │   ├── entry_service.go
│   │   ├── search_service.go
│   │   ├── generator_service.go
│   │   ├── totp_service.go
│   │   ├── audit_service.go
│   │   ├── backup_service.go
│   │   └── doctor_service.go
│   │
│   ├── domain/
│   │   ├── entry.go
│   │   ├── login.go
│   │   ├── note.go
│   │   ├── credential.go
│   │   ├── field.go
│   │   ├── history.go
│   │   ├── ids.go
│   │   └── errors.go
│   │
│   ├── crypto/
│   │   ├── crypto.go
│   │   ├── keys.go
│   │   ├── kdf.go
│   │   ├── hkdf.go
│   │   ├── aead.go
│   │   ├── aad.go
│   │   ├── wrap.go
│   │   ├── random.go
│   │   └── zero.go
│   │
│   ├── storage/
│   │   ├── storage.go
│   │   └── sqlite/
│   │       ├── store.go
│   │       ├── schema.go
│   │       ├── migrations.go
│   │       ├── entries.go
│   │       ├── history.go
│   │       └── metadata.go
│   │
│   ├── serialization/
│   │   ├── entry.go
│   │   ├── history.go
│   │   └── versions.go
│   │
│   ├── backup/
│   │   ├── format.go
│   │   ├── header.go
│   │   ├── payload.go
│   │   ├── reader.go
│   │   ├── writer.go
│   │   ├── verify.go
│   │   └── restore.go
│   │
│   ├── generator/
│   │   ├── password.go
│   │   ├── passphrase.go
│   │   ├── entropy.go
│   │   └── wordlist.go
│   │
│   ├── totp/
│   │   ├── totp.go
│   │   └── validate.go
│   │
│   ├── audit/
│   │   ├── audit.go
│   │   ├── weak.go
│   │   ├── reused.go
│   │   └── age.go
│   │
│   ├── clipboard/
│   │   ├── clipboard.go
│   │   └── platform_*.go
│   │
│   ├── config/
│   │   ├── config.go
│   │   ├── defaults.go
│   │   └── paths.go
│   │
│   ├── cli/
│   │   ├── root.go
│   │   ├── list.go
│   │   ├── search.go
│   │   ├── show.go
│   │   ├── copy.go
│   │   ├── add.go
│   │   ├── edit.go
│   │   ├── delete.go
│   │   ├── generate.go
│   │   ├── backup.go
│   │   ├── restore.go
│   │   ├── audit.go
│   │   └── doctor.go
│   │
│   └── tui/
│       ├── app.go
│       ├── model.go
│       ├── update.go
│       ├── view.go
│       ├── keys.go
│       ├── theme.go
│       ├── components/
│       └── screens/
│
├── migrations/
│   └── ...
│
├── docs/
│   ├── prd.md
│   ├── threat-model.md
│   ├── cryptography.md
│   ├── vault-format.md
│   ├── backup-format.md
│   └── architecture.md
│
├── testdata/
│   ├── vaults/
│   └── backups/
│
├── go.mod
├── go.sum
├── README.md
└── LICENSE
```

This structure is intentionally modular without turning every small concept into its own package.

---

# 6. Package Dependency Rules

The architecture should enforce a mostly acyclic dependency graph.

Recommended direction:

```text
domain
  ▲
  │
serialization
  ▲
  │
app
  ▲
  │
CLI / TUI
```

Infrastructure implements interfaces consumed by the application layer.

---

# 7. Domain Package

`internal/domain` contains logical GoVault concepts.

It must not depend on:

```text
SQLite
Cobra
Bubble Tea
Lip Gloss
filesystem layout
clipboard implementations
```

Ideally it also does not depend directly on concrete cryptographic implementations.

---

# 8. Domain Entry Model

A possible logical model:

```go
type Entry struct {
    ID        EntryID
    Kind      EntryKind

    Name      string
    Tags      []string
    Aliases   []string
    Favorite  bool

    Fields    []Field

    CreatedAt time.Time
    UpdatedAt time.Time
}
```

Entry-type-specific models may wrap or extend this concept.

---

# 9. Entry Kinds

Conceptually:

```go
type EntryKind uint16

const (
    EntryLogin EntryKind = iota + 1
    EntrySecureNote
    EntryAPICredential
    EntryDatabaseCredential
    EntryCustom
)
```

Future:

```text
SSH Key
```

---

# 10. Generic Fields

A generic encrypted field model allows GoVault to support flexible record types.

Conceptually:

```go
type Field struct {
    Name      string
    Value     string
    Sensitive bool
}
```

---

# 11. Sensitive Flag

The `Sensitive` property is primarily a presentation policy.

It determines whether a value should:

```text
be masked
require explicit reveal
be copied instead of displayed
```

All fields remain encrypted at rest regardless of this flag.

---

# 12. Domain IDs

Identifiers should use dedicated types.

Example:

```go
type EntryID [16]byte
type HistoryID [16]byte
type VaultID [16]byte
type BackupID [16]byte
```

Avoid passing arbitrary `[]byte` identifiers throughout application code.

---

# 13. Domain Errors

Domain errors should represent semantic failures.

Examples:

```text
entry not found
ambiguous alias
duplicate alias
invalid entry
unsupported entry type
```

They should not contain SQLite-specific details.

---

# 14. Application Layer

`internal/app` implements GoVault use cases.

Examples:

```text
UnlockVault
LockVault
CreateEntry
UpdateEntry
DeleteEntry
GetEntry
SearchEntries
CopySecret
GeneratePassword
GeneratePassphrase
GenerateTOTP
RunAudit
CreateBackup
VerifyBackup
RestoreBackup
RunDoctor
ChangeMasterPassword
```

---

# 15. Application Service Responsibilities

Application services coordinate:

```text
domain
crypto
storage
serialization
clipboard
backup
configuration
```

They should not directly implement low-level cryptographic primitives.

---

# 16. Session

The application session represents an unlocked vault.

Conceptually:

```go
type Session struct {
    vaultID VaultID

    keys SessionKeys

    index SearchIndex

    unlockedAt time.Time
    lastActivity time.Time
}
```

---

# 17. Session Keys

Conceptually:

```go
type SessionKeys struct {
    VaultKey   SecretKey
    EntryKey   SecretKey
    HistoryKey SecretKey
    BackupKey  SecretKey
}
```

Whether all derived keys are retained or derived on demand should be evaluated.

---

# 18. Session Ownership

There should be one clear owner of unlocked key material.

Presentation code must never receive raw keys.

---

# 19. Locked Application State

When locked:

```text
Vault Key absent
derived keys absent
search index absent
decrypted entry cache absent
```

Only non-secret configuration and vault header metadata remain available.

---

# 20. Unlock Flow

Conceptually:

```text
User
 │
 ▼
CLI/TUI password prompt
 │
 ▼
Application Unlock Service
 │
 ├──► Load vault metadata
 │
 ├──► Validate KDF bounds
 │
 ├──► Derive KEK
 │
 ├──► Authenticate + unwrap Vault Key
 │
 ├──► Derive subkeys
 │
 ├──► Build search index
 │
 ▼
Unlocked Session
```

---

# 21. Password Ownership

The presentation layer obtains the password from the terminal.

It passes it to the unlock service through a narrow API.

The password must not be:

```text
stored in config
logged
included in errors
retained after unlock
```

---

# 22. Unlock API

Conceptually:

```go
func (a *App) Unlock(
    ctx context.Context,
    password []byte,
) error
```

Using `[]byte` instead of immutable Go strings may allow somewhat better lifecycle control.

It does not guarantee complete memory erasure.

---

# 23. Unlock Failure

On failure:

```text
no session created
no partial search index retained
temporary key material discarded
```

---

# 24. Lock Flow

Conceptually:

```text
Lock requested
     │
     ▼
cancel secret operations
     │
     ▼
clear clipboard if owned
     │
     ▼
drop search index
     │
     ▼
discard cached plaintext
     │
     ▼
clear session keys where practical
     │
     ▼
clear terminal
     │
     ▼
LOCKED
```

---

# 25. Lock Triggers

Lock may occur because of:

```text
explicit user command
TUI shortcut
inactivity timeout
process shutdown
fatal security-sensitive error
```

---

# 26. Auto-Lock

Auto-lock belongs to the application/session layer rather than individual TUI screens.

The TUI reports activity.

The session manager tracks:

```text
lastActivity
autoLockDuration
```

---

# 27. Activity Events

Activity may include:

```text
keyboard input
entry navigation
copy operation
edit operation
search
```

Background timers should not count as user activity.

---

# 28. Auto-Lock Warning

The TUI may show:

```text
Vault will lock in 30 seconds due to inactivity.
```

Any user activity may reset the timer.

---

# 29. Crypto Package

`internal/crypto` owns low-level cryptographic operations.

It is one of the most security-sensitive packages in the project.

---

# 30. Crypto Responsibilities

The package handles:

```text
Argon2id
HKDF-SHA-256
XChaCha20-Poly1305
random generation
Vault Key wrapping
Vault Key unwrapping
AAD encoding
key derivation
nonce generation
```

---

# 31. Crypto Package Must Not Know UI

The crypto package must not import:

```text
Cobra
Bubble Tea
Lip Gloss
```

---

# 32. Crypto Package Must Not Know SQLite

It should operate on typed cryptographic structures.

Example:

```go
type EncryptedEnvelope struct {
    Nonce      Nonce
    Ciphertext []byte
}
```

---

# 33. Key Types

Use distinct key types where practical.

Conceptually:

```go
type VaultKey struct {
    bytes [32]byte
}

type EntryKey struct {
    bytes [32]byte
}

type HistoryKey struct {
    bytes [32]byte
}

type BackupKey struct {
    bytes [32]byte
}
```

This reduces accidental key misuse.

---

# 34. Key Access

Avoid APIs exposing key bytes casually.

Prefer operations such as:

```go
EncryptEntry(...)
DecryptEntry(...)
WrapVaultKey(...)
UnwrapVaultKey(...)
```

over:

```go
GetKeyBytes()
```

---

# 35. Randomness API

All cryptographic randomness comes from:

```text
crypto/rand
```

Centralize generation of:

```text
Vault IDs
Entry IDs
History IDs
Backup IDs
Vault Keys
KDF salts
AEAD nonces
```

---

# 36. Serialization Package

`internal/serialization` translates domain objects into canonical plaintext representations.

It must be independent from SQLite.

Conceptually:

```text
Domain Entry
    │
    ▼
Serializer
    │
    ▼
Canonical Bytes
    │
    ▼
Crypto
```

---

# 37. Serialization Responsibilities

It handles:

```text
entry serialization
entry deserialization
history serialization
version handling
field validation
canonical encoding
```

---

# 38. Serialization Is Security Sensitive

Authenticated ciphertext may contain malformed plaintext produced by:

```text
old buggy GoVault versions
test fixtures
future incompatible implementations
```

Therefore authenticated plaintext still requires strict parsing.

---

# 39. Storage Interface

The application should depend on an abstract storage interface.

Conceptually:

```go
type Store interface {
    Metadata(ctx context.Context) (VaultMetadata, error)

    ListEntries(ctx context.Context) ([]EntryEnvelope, error)

    GetEntry(
        ctx context.Context,
        id domain.EntryID,
    ) (EntryEnvelope, error)

    CreateEntry(
        ctx context.Context,
        entry EntryEnvelope,
    ) error

    UpdateEntry(
        ctx context.Context,
        current EntryEnvelope,
        history HistoryEnvelope,
    ) error

    DeleteEntry(
        ctx context.Context,
        id domain.EntryID,
    ) error
}
```

---

# 40. Storage Envelope

Storage receives encrypted objects.

Conceptually:

```go
type EntryEnvelope struct {
    ID                   domain.EntryID
    Kind                 domain.EntryKind
    CryptoSuiteVersion   uint16
    SerializationVersion uint16
    Nonce                []byte
    Ciphertext           []byte
    CreatedAt            time.Time
    UpdatedAt            time.Time
}
```

---

# 41. Storage Boundary Rule

The SQLite package should never receive:

```text
password field plaintext
TOTP secret plaintext
secure note body plaintext
API secret plaintext
```

---

# 42. SQLite Adapter

`internal/storage/sqlite` implements the storage interfaces.

Responsibilities:

```text
database creation
schema validation
queries
transactions
migrations
SQLite integrity checks
vault metadata persistence
encrypted envelope persistence
```

---

# 43. SQLite Adapter Does Not Encrypt

This is deliberately wrong:

```text
storage.SavePassword(...)
```

Preferred:

```text
crypto.EncryptEntry(...)
→ storage.SaveEnvelope(...)
```

---

# 44. Transaction Ownership

Storage owns database transaction mechanics.

Application services define logical atomic operations.

Example:

```text
update entry + create history
```

must map to one storage transaction.

---

# 45. Search Architecture

Search is performed using an in-memory index.

No plaintext searchable index is persisted.

---

# 46. Search Index Model

Conceptually:

```go
type SearchDocument struct {
    EntryID   domain.EntryID
    Kind      domain.EntryKind
    Name      string
    Username  string
    Website   string
    Tags      []string
    Aliases   []string
    Favorite  bool
}
```

---

# 47. Search Index Contents

Do not include:

```text
passwords
TOTP seeds
API secrets
SSH private keys
secure note bodies
```

by default.

---

# 48. Search Lifecycle

```text
unlock
  ↓
decrypt entries
  ↓
extract metadata
  ↓
build index

lock
  ↓
destroy index
```

---

# 49. Search Implementation

Initial search can use an in-memory fuzzy matcher.

Avoid introducing persistent indexing infrastructure until performance requires it.

---

# 50. Entry Read Flow

Conceptually:

```text
UI
 │
 ▼
EntryService.Get
 │
 ▼
Storage.GetEnvelope
 │
 ▼
Crypto.Decrypt
 │
 ▼
Serialization.Decode
 │
 ▼
Domain Entry
 │
 ▼
UI
```

---

# 51. Entry Write Flow

```text
UI
 │
 ▼
EntryService.Create
 │
 ▼
Validate Domain Entry
 │
 ▼
Serialize
 │
 ▼
Encrypt
 │
 ▼
Storage Insert
```

---

# 52. Entry Update Flow

```text
UI
 │
 ▼
EntryService.Update
 │
 ├──► Load current entry
 │
 ├──► Decrypt current entry
 │
 ├──► Serialize history snapshot
 │
 ├──► Encrypt history
 │
 ├──► Validate updated entry
 │
 ├──► Serialize updated entry
 │
 ├──► Encrypt updated entry
 │
 ▼
Atomic storage transaction
```

---

# 53. Entry Delete Flow

Delete should require explicit application intent.

Presentation layers should implement confirmation UX.

The application service should still make deletion explicit:

```go
DeleteEntry(ctx, id)
```

rather than exposing a generic destructive storage operation.

---

# 54. History Service

History behavior should be coordinated by `EntryService` or a dedicated `HistoryService`.

Presentation layers should not manually create history records.

---

# 55. Backup Package

`internal/backup` implements `.gvault` container handling.

Responsibilities:

```text
header encoding
header decoding
manifest serialization
payload serialization
backup creation
verification
restore preparation
format validation
```

---

# 56. Backup Package Boundary

The backup package may work with logical domain entries.

It must not depend on SQLite physical row layout.

---

# 57. Backup Service

`internal/app/backup_service.go` coordinates:

```text
storage
crypto
serialization
backup writer
filesystem
```

---

# 58. Backup Creation Flow

```text
CLI/TUI
  │
  ▼
BackupService
  │
  ├──► require unlocked session
  ├──► read entries
  ├──► decrypt entries
  ├──► read/decrypt history
  ├──► construct logical backup
  ├──► encrypt backup
  ├──► write temporary file
  ├──► verify
  └──► atomic rename
```

---

# 59. Restore Architecture

Restore should be implemented as a staged operation.

```text
.gvault
   │
   ▼
Backup Parser
   │
   ▼
Authentication
   │
   ▼
Logical Validation
   │
   ▼
Temporary Vault Builder
   │
   ▼
Deep Validation
   │
   ▼
Atomic Replacement
```

---

# 60. Restore Must Not Mutate Active Storage Early

The current vault remains untouched until the replacement vault has been completely built and validated.

---

# 61. Generator Package

`internal/generator` handles:

```text
random passwords
passphrases
entropy estimates
generator policies
embedded wordlists
```

It must use cryptographically secure randomness.

---

# 62. Password Generation

The generator should accept a policy.

Conceptually:

```go
type PasswordPolicy struct {
    Length         int
    Uppercase      bool
    Lowercase      bool
    Numbers        bool
    Symbols        bool
    AvoidAmbiguous bool
}
```

---

# 63. Generator Independence

Password generation should not depend on:

```text
storage
TUI
CLI
```

This makes it easy to test and reuse.

---

# 64. Passphrase Wordlist

The passphrase wordlist must ship with GoVault.

GoVault must not download dictionaries.

---

# 65. TOTP Package

`internal/totp` handles offline TOTP generation.

Responsibilities:

```text
secret validation
counter calculation
HMAC
digit truncation
period handling
supported algorithms
```

---

# 66. TOTP Has No Network Dependency

Current time comes from the local system clock.

No remote clock synchronization is performed by GoVault.

---

# 67. Audit Package

`internal/audit` analyzes decrypted logical vault data in memory.

Checks may include:

```text
weak passwords
reused passwords
old passwords
missing TOTP metadata
password uniqueness
```

---

# 68. Audit Privacy

Audit results remain local.

No password or fingerprint is transmitted.

---

# 69. Reuse Detection

Password reuse detection should compare transient in-memory derived values or direct values under tightly scoped processing.

Persistent password fingerprints are prohibited.

---

# 70. Clipboard Interface

Clipboard operations should be abstracted.

Conceptually:

```go
type Clipboard interface {
    Write(ctx context.Context, value []byte) error
    Clear(ctx context.Context) error
}
```

---

# 71. Clipboard Ownership

GoVault should track whether it believes it currently owns the clipboard content.

This helps avoid clearing unrelated user content.

---

# 72. Safe Clipboard Clearing

A stronger design is:

```text
write secret
remember expected clipboard value/fingerprint
wait timeout
check ownership where supported
clear only if still owned
```

Platform capabilities vary.

The implementation must document limitations.

---

# 73. Clipboard Timeout

Default example:

```text
30 seconds
```

Configurable.

---

# 74. Clipboard Cancellation

A new copy operation cancels the previous clear timer and establishes new clipboard ownership.

---

# 75. CLI Architecture

Cobra is the command routing layer.

CLI commands should remain thin.

Bad:

```text
cobra command
→ SQL
→ Argon2
→ encryption
```

Good:

```text
cobra command
→ application service
```

---

# 76. Root CLI

Conceptually:

```text
govault
```

without subcommands launches the TUI.

---

# 77. Core CLI Commands

Initial commands:

```text
govault
govault list
govault search
govault show
govault copy
govault add
govault edit
govault delete
govault generate
govault totp
govault audit
govault backup
govault restore
govault lock
govault status
govault doctor
```

---

# 78. Secret Output Policy

CLI commands must distinguish:

```text
display metadata
copy secret
explicitly reveal secret
machine-readable output
```

Passwords must never appear in normal output accidentally.

---

# 79. `show`

Example:

```text
$ govault show github

Name      GitHub Personal
Username  john@example.com
Password  ••••••••••••••••
TOTP      ••••••
Tags      personal, dev
```

---

# 80. Secret Reveal

If explicit reveal is supported:

```text
govault show github --reveal
```

must be an intentional action.

---

# 81. JSON Output

Machine-readable output should default to metadata-only.

Example:

```text
govault list --json
```

must not include password values.

---

# 82. Explicit Secret JSON

If GoVault ever supports secrets in JSON output, it should require a deliberately dangerous explicit flag.

This should not be part of the MVP unless clearly needed.

---

# 83. CLI Exit Codes

Define stable categories.

Conceptually:

```text
0 success
1 general failure
2 usage error
3 vault locked
4 entry not found
5 ambiguous selector
6 authentication failure
7 corrupted vault
```

Exact values should be documented before v1.0.

---

# 84. TUI Architecture

The Bubble Tea application is a state machine over application services.

Recommended conceptual hierarchy:

```text
Root Model
   │
   ├── Lock Screen
   ├── Dashboard
   ├── Search
   ├── Entry Detail
   ├── Entry Editor
   ├── Generator
   ├── TOTP
   ├── Audit
   ├── Backup
   ├── Restore
   ├── Doctor
   ├── Settings
   └── Help
```

---

# 85. TUI Screen Interface

A lightweight internal abstraction may be useful.

Conceptually:

```go
type Screen interface {
    Init() tea.Cmd
    Update(tea.Msg) (Screen, tea.Cmd)
    View() string
}
```

This should be adopted only if it simplifies Bubble Tea composition.

Avoid unnecessary framework-building.

---

# 86. Root TUI Model

The root model owns:

```text
current screen
session state
terminal dimensions
global key bindings
notifications
modal state
auto-lock state
```

---

# 87. TUI Global Shortcuts

Suggested:

```text
/       search
Ctrl+K  command palette
n       new entry
g       generator
a       audit
b       backup
?       help
Ctrl+L  lock
```

Context-specific screens may override non-security-sensitive shortcuts.

---

# 88. TUI Screen: First-Run Setup

When no vault exists:

```text
╭──────────────────────────────────────╮
│             ◆ GoVault                │
│                                      │
│       Create your first vault        │
│                                      │
│  Master Password                     │
│  [••••••••••••••••••••••••]         │
│                                      │
│  Confirm Password                    │
│  [••••••••••••••••••••••••]         │
│                                      │
│  Strength                            │
│  ████████████████░░░  Strong         │
│                                      │
│       [ Create Vault ]               │
╰──────────────────────────────────────╯
```

---

# 89. TUI Screen: Lock

```text
╭──────────────────────────────────────╮
│                                      │
│             ◆ GoVault                │
│                                      │
│            VAULT LOCKED              │
│                                      │
│  Master Password                     │
│  [••••••••••••••••••••••••]         │
│                                      │
│           [ Unlock ]                 │
│                                      │
│          100% offline                │
╰──────────────────────────────────────╯
```

---

# 90. TUI Screen: Dashboard

```text
◆ GoVault                         UNLOCKED

╭─ Vault ───────────╮ ╭─ Security ─────────╮
│ 196 entries       │ │ Score       92/100 │
│ 24 favorites      │ │ ███████████████░░  │
│ 38 TOTP           │ │                    │
╰───────────────────╯ ╰────────────────────╯

Recent
──────────────────────────────────────────
GitHub Personal          Login
Production Database      Database
AWS Personal             API Credential

/ Search   n New   g Generate   a Audit
b Backup   Ctrl+K Commands      ? Help
```

---

# 91. TUI Screen: Search

```text
◆ Search

╭──────────────────────────────────────╮
│ github_                              │
╰──────────────────────────────────────╯

GitHub Personal
john@example.com
login · personal · dev

GitHub Work
john@company.example
login · work

↑↓ Navigate   Enter Open   Esc Close
```

---

# 92. TUI Screen: Advanced Search

A future or early v0.2 screen can support filters:

```text
Type       [ Login ▼ ]
Tags       [ dev, work ]
Favorite   [ Any ▼ ]
TOTP       [ Any ▼ ]
Updated    [ Any time ▼ ]
```

Search filters remain entirely local.

---

# 93. TUI Screen: Entry Detail

```text
GoVault / Logins / GitHub Personal

╭──────────────────────────────────────╮
│ Username                             │
│ john@example.com              [copy] │
│                                      │
│ Password                             │
│ •••••••••••••••••••          [copy] │
│                                      │
│ TOTP                                 │
│ ••••••                        [copy] │
│                                      │
│ Website                              │
│ github.com                           │
│                                      │
│ Tags                                 │
│ personal  dev                        │
╰──────────────────────────────────────╯

p Password   y Username   t TOTP
r Reveal     e Edit       h History
Esc Back
```

---

# 94. TUI Screen: Secret Reveal

Secret reveal should be temporary.

Example:

```text
Password
correct-horse-battery-staple

Visible for 8s
██████████████░░░░
```

After timeout:

```text
••••••••••••••••••••••••••••
```

---

# 95. TUI Screen: New Entry Type

```text
◆ New Item

› Login
  Secure Note
  API Credential
  Database Credential
  Custom

↑↓ Select   Enter Continue   Esc Cancel
```

---

# 96. TUI Screen: Login Editor

```text
◆ New Login

Name
[ GitHub Personal                 ]

Username
[ john@example.com                ]

Password
[ •••••••••••••••••••••••        ]
Strength  ████████████████░ Strong

Website
[ github.com                      ]

Tags
[ personal, dev                   ]

[ Generate Password ]      [ Save ]

Tab Next   Ctrl+G Generate
Ctrl+S Save   Esc Cancel
```

---

# 97. TUI Screen: Secure Note Editor

```text
◆ Secure Note

Title
[ Recovery Codes                  ]

Content
╭──────────────────────────────────────╮
│ code-1                               │
│ code-2                               │
│ code-3                               │
│                                      │
╰──────────────────────────────────────╯

Tags
[ security, recovery ]

Ctrl+S Save   Esc Cancel
```

---

# 98. TUI Screen: Password Generator

```text
◆ Password Generator

  v9#zTq8!Lp2@Ax7$Kf4&

  Strength
  ████████████████████  Excellent

  Entropy
  ~126 bits

  Length          20
  Uppercase       ✓
  Lowercase       ✓
  Numbers         ✓
  Symbols         ✓
  Avoid ambiguous ✓

r Regenerate   c Copy   s Save
Tab Passphrase   Esc Back
```

---

# 99. TUI Screen: Passphrase Generator

```text
◆ Passphrase Generator

orbit-cactus-river-lantern-signal

Words           5
Separator       -
Capitalize      No
Add Number      No

Estimated entropy
█████████████████░  Strong

r Regenerate   c Copy   s Save
Tab Password   Esc Back
```

---

# 100. TUI Screen: TOTP

```text
◆ Authenticator

Search: _

GitHub Personal       284 921    ███████░  18s
AWS Personal          810 443    ████░░░░  11s
Company VPN           550 129    ██░░░░░░   6s

Enter Copy   / Search   Esc Back
```

---

# 101. TUI Screen: TOTP Setup

Because GoVault is offline, TOTP setup should allow manual entry.

```text
◆ Add TOTP

Issuer
[ GitHub                           ]

Account
[ john@example.com                 ]

Secret
[ ••••••••••••••••••••••••••••• ]

Algorithm
[ SHA-1 ▼ ]

Digits
[ 6 ]

Period
[ 30 ]

[ Validate ]              [ Save ]
```

QR decoding may be considered later if it can remain completely local.

---

# 102. TUI Screen: Security Audit

```text
◆ Security Audit

Vault Health
██████████████████░░  92/100

Strong & unique       173
Reused                  7
Weak                    5
Old                    11

Issues
──────────────────────────────────────
7 credentials reuse passwords
5 credentials have weak passwords
11 passwords are older than policy

r Reused   w Weak   o Old
Enter Inspect   Esc Back
```

---

# 103. TUI Screen: Reused Password Remediation

```text
◆ Reused Passwords

Password Group 1                     3 uses

› GitHub Personal
  Old Forum
  Test Server

[ Generate Replacement ]

n Next Group   Enter Open
Esc Back
```

The actual reused password should not need to be displayed.

---

# 104. TUI Screen: Backup

```text
◆ Create Encrypted Backup

Destination
[ ~/backups/govault-2026-09-11.gvault ]

Entries       196
History       824
Format        GoVault Backup v1
Encryption    Authenticated

            [ Create Backup ]

Esc Cancel
```

---

# 105. TUI Screen: Backup Success

```text
◆ Backup Complete

✓ Backup written
✓ Structure verified
✓ Cryptographic verification passed

Entries        196
History        824
Size           384 KiB

~/backups/govault-2026-09-11.gvault

GoVault does not upload or synchronize backups.

Enter Done
```

---

# 106. TUI Screen: Restore

```text
◆ Restore Backup

File
[ ~/backups/govault-2026-09-11.gvault ]

✓ Valid GoVault backup
✓ Payload authenticated

Created       Sep 11, 2026
Entries       196
History       824

WARNING

The current vault will be replaced.

A recovery snapshot will be created first.

        [ Restore Vault ]

Esc Cancel
```

---

# 107. TUI Screen: Doctor

```text
◆ GoVault Doctor

✓ Vault database healthy
✓ SQLite integrity
✓ Vault permissions
✓ Configuration permissions
✓ Crypto suite supported
✓ KDF parameters safe
✓ Random source available
✓ Backup path writable
✓ Clipboard integration
✓ Network dependencies: none

Everything looks good.

r Run Again   Esc Back
```

---

# 108. TUI Screen: History

```text
◆ GitHub Personal / History

Sep 10 2026  18:42
Password changed

Aug 03 2026  09:12
TOTP added

Jul 28 2026  15:05
Username changed

Jul 28 2026  14:58
Entry created

Enter Inspect   r Restore   Esc Back
```

---

# 109. TUI Screen: Tags

```text
◆ Tags

personal                 42
work                     38
development              27
finance                   12
security                   8

Enter Browse   / Search   Esc Back
```

Tags are obtained from the in-memory unlocked index.

---

# 110. TUI Screen: Favorites

```text
◆ Favorites

GitHub Personal
AWS Personal
Primary Email
Production Database
Recovery Codes

Enter Open   / Search   Esc Back
```

---

# 111. TUI Screen: Settings

```text
◆ Settings

Security
  Auto-lock                 5 minutes
  Clipboard clear          30 seconds
  Reveal timeout           10 seconds

Interface
  Startup screen           Search
  TOTP codes               Masked
  Theme                    Default

Vault
  History                  Enabled

Enter Change   Esc Back
```

---

# 112. TUI Screen: Change Master Password

```text
◆ Change Master Password

Current Password
[ ••••••••••••••••••••• ]

New Password
[ ••••••••••••••••••••• ]

Confirm
[ ••••••••••••••••••••• ]

Strength
██████████████████░  Strong

[ Change Master Password ]

Esc Cancel
```

The UI should explain that this operation rewraps the Vault Key rather than exporting the vault.

---

# 113. TUI Screen: Command Palette

```text
◆ Commands

> backup_

Create encrypted backup
Verify backup
Restore backup

────────────────────────────────────
↑↓ Navigate   Enter Run   Esc Close
```

---

# 114. TUI Screen: Help Overlay

```text
◆ Keyboard Shortcuts

Navigation
  ↑ ↓       Move
  Enter     Open
  Esc       Back

Global
  /         Search
  Ctrl+K    Commands
  Ctrl+L    Lock
  ?         Help

Entries
  n         New
  e         Edit
  p         Copy password
  y         Copy username
  r         Reveal secret
```

---

# 115. TUI Screen: Delete Confirmation

Deletion should require stronger confirmation.

```text
◆ Delete Entry

Delete:

GitHub Personal

This action cannot be undone from the active vault.

Type DELETE to confirm:

[                              ]

[ Cancel ]             [ Delete ]
```

If history/recovery behavior changes later, wording must remain accurate.

---

# 116. TUI Screen: Auto-Lock Warning

```text
╭──────────────────────────────────────╮
│                                      │
│        Vault locking soon            │
│                                      │
│  No activity detected.               │
│                                      │
│  GoVault will lock in 30 seconds.    │
│                                      │
│      [ Keep Vault Unlocked ]         │
│                                      │
╰──────────────────────────────────────╯
```

---

# 117. TUI Notifications

Short-lived notifications should be non-invasive.

Examples:

```text
✓ Password copied · clears in 30s

✓ Entry saved

✓ Vault locked

⚠ Clipboard could not be cleared
```

---

# 118. Secret Copy Notification

The notification must not include the secret.

Bad:

```text
Copied: my-super-secret-password
```

Good:

```text
Password copied · clears in 30s
```

---

# 119. TUI Error Screen

Fatal vault problems should have a dedicated screen.

Example:

```text
◆ GoVault

Unable to open vault.

The database appears to be corrupted.

Your vault was not modified.

Run:

govault doctor

or restore an encrypted backup.

q Quit
```

---

# 120. Migration Screen

If migration is required:

```text
◆ Vault Upgrade Required

Current format     v1
Target format      v2

A recovery backup will be created before migration.

[ Upgrade Vault ]

Esc Exit
```

---

# 121. Empty States

Every primary screen should define a useful empty state.

Example:

```text
◆ GoVault

Your vault is empty.

n Create your first login
g Generate a password
? Open help
```

---

# 122. Responsive TUI

The interface should support:

```text
standard terminals
small terminals
wide terminals
SSH sessions
```

without assuming a specific resolution.

---

# 123. Minimum Terminal Size

GoVault may define a minimum usable size.

Example:

```text
80 × 24 recommended
```

Below minimum:

```text
Terminal is too small.
Resize to at least 60 × 18.
```

Exact dimensions should be tested.

---

# 124. Unicode Fallback

Decorative symbols such as:

```text
◆
✓
⚠
```

should have ASCII fallbacks if terminal capabilities require them.

---

# 125. Color Is Not Security State

Critical states must not rely only on color.

For example:

```text
LOCKED
UNLOCKED
SECRET VISIBLE
```

must have textual indicators.

---

# 126. Configuration Package

`internal/config` owns non-secret configuration.

Possible settings:

```text
auto_lock_duration
clipboard_clear_duration
secret_reveal_duration
startup_screen
theme
totp_mask_codes
history_enabled
```

---

# 127. Configuration Format

A simple local format may be used.

Examples:

```text
TOML
JSON
```

The exact choice is not security-critical.

Configuration must never contain cryptographic keys or vault secrets.

---

# 128. Configuration Permissions

On Unix-like systems:

```text
0600
```

is recommended.

---

# 129. Configuration Corruption

Invalid configuration should not prevent vault recovery.

Prefer:

```text
warn
use safe defaults
```

for non-security-critical preferences.

---

# 130. Security Defaults

Security-related invalid values should fall back to safe defaults.

Example:

```text
clipboard_clear_duration = -1
```

must not silently disable clearing.

---

# 131. Default Configuration

Suggested initial defaults:

```text
auto-lock             5 minutes
clipboard clear       30 seconds
secret reveal         10 seconds
TOTP display          masked
history               enabled
startup screen        search/dashboard
```

Final defaults should be decided during UX testing.

---

# 132. Dependency Injection

GoVault should use explicit constructor-based dependency injection.

No dependency injection framework is necessary.

Example:

```go
app := app.New(
    store,
    cryptoService,
    clipboard,
    backupService,
    config,
)
```

---

# 133. Why Explicit Injection

Benefits:

```text
clear dependencies
simple tests
easy mocks/fakes
no global state
no runtime container magic
```

---

# 134. Avoid Global Vault State

Do not store:

```text
Vault Key
current entry
database connection
clipboard secret
```

in package-level global variables.

---

# 135. Application Composition Root

`cmd/govault/main.go` acts as the composition root.

Conceptually:

```text
load paths
load config
open storage
construct crypto services
construct clipboard adapter
construct application
construct CLI
execute
```

---

# 136. `main.go`

`main.go` should remain small.

It should not contain:

```text
SQL
cryptographic logic
Bubble Tea screen logic
backup parsing
```

---

# 137. Context Usage

Long-running or cancellable operations should accept:

```go
context.Context
```

Examples:

```text
backup
restore
doctor
audit
search-index construction
```

---

# 138. Cancellation

Cancellation must leave persistent state consistent.

A cancelled backup:

```text
must not replace existing backup
```

A cancelled restore:

```text
must not replace active vault
```

---

# 139. Logging

Logging must be conservative.

Never log:

```text
master password
password fields
TOTP secrets
API keys
private keys
decrypted secure notes
Vault Key
derived keys
ciphertext unless diagnostic need is clear
```

---

# 140. Default Logging

Normal users should receive concise errors rather than verbose internal logs.

A future debug mode must still redact secrets.

---

# 141. Panic Handling

Unexpected panics should not dump secret-bearing structs.

Recovery behavior should prioritize process termination and terminal cleanup.

---

# 142. Terminal Cleanup

On exit or lock, GoVault should attempt to:

```text
restore terminal mode
hide sensitive screen contents
clear temporary reveal state
```

Bubble Tea shutdown handling must be tested carefully.

---

# 143. Signal Handling

Relevant signals may trigger graceful cleanup.

Examples on Unix-like systems:

```text
SIGINT
SIGTERM
```

Cleanup should not delay termination excessively.

---

# 144. Clipboard Cleanup on Exit

If GoVault believes it owns the clipboard secret, it should attempt to clear it during graceful shutdown.

This is best-effort.

---

# 145. File Paths

Application paths should be centralized in `internal/config/paths.go`.

Conceptually:

```go
type Paths struct {
    DataDir    string
    ConfigDir  string
    VaultFile  string
}
```

---

# 146. Platform Paths

Use platform conventions rather than hardcoding:

```text
~/.govault
```

for every operating system.

---

# 147. Offline Enforcement

Offline operation should be verified at multiple levels.

First:

```text
no application network features
```

Second:

```text
avoid network-capable dependencies
```

Third:

```text
CI dependency inspection
```

Fourth:

```text
integration tests under blocked network
```

---

# 148. Forbidden Network Imports

Production code should not normally import:

```text
net/http
net/smtp
net/rpc
```

Direct `net` usage should also require explicit architectural review.

---

# 149. CI Network Check

A CI job may scan the dependency graph for unexpected networking packages.

Exceptions must be reviewed rather than silently ignored.

---

# 150. Runtime Offline Test

Integration tests should run GoVault in an environment where outbound network access is unavailable.

Core functionality must still work.

---

# 151. No Telemetry

The architecture contains no telemetry service.

There is no:

```text
analytics client
crash-report uploader
usage metrics endpoint
remote configuration
```

---

# 152. No Automatic Updates

GoVault does not contact servers to check for releases.

Package managers or users may manage upgrades externally.

---

# 153. Doctor Architecture

`DoctorService` coordinates diagnostics across components.

Checks may include:

```text
filesystem
configuration
SQLite
vault metadata
cryptographic configuration
randomness
clipboard
backup path
dependency/network policy
```

---

# 154. Doctor Check Model

Conceptually:

```go
type CheckResult struct {
    Name     string
    Status   CheckStatus
    Message  string
}
```

Statuses:

```text
Pass
Warning
Fail
```

---

# 155. Doctor Must Not Mutate by Default

Running:

```text
govault doctor
```

should be read-only unless a repair mode is explicitly requested.

---

# 156. Repair Architecture

A future:

```text
govault doctor --repair
```

should use explicit repair operations.

Never automatically rewrite corrupted encrypted records based on guesses.

---

# 157. Error Architecture

Errors should flow upward while preserving classification.

Example:

```text
SQLite error
   ↓
storage typed error
   ↓
application error
   ↓
CLI/TUI presentation
```

---

# 158. Secret-Safe Errors

Error strings must not contain serialized entry contents.

---

# 159. Error Wrapping

Go errors may be wrapped with context:

```go
fmt.Errorf("load vault metadata: %w", err)
```

provided nested errors do not contain secrets.

---

# 160. Testing Layers

GoVault should have:

```text
unit tests
integration tests
format compatibility tests
security invariant tests
fuzz tests
TUI model tests
CLI tests
```

---

# 161. Domain Tests

Test:

```text
entry validation
aliases
tags
field rules
ID behavior
```

---

# 162. Crypto Tests

Test:

```text
Argon2id parameters
key wrapping
key unwrapping
AAD
nonce lengths
cross-record substitution
cross-vault substitution
tampering
wrong passwords
```

---

# 163. Storage Tests

Test against temporary SQLite databases:

```text
schema creation
CRUD
transactions
history
migration
corruption handling
```

---

# 164. Backup Tests

Test:

```text
create
inspect
verify
restore
tampering
truncation
wrong password
old password backup
atomic replacement
```

---

# 165. Application Tests

Application services should be testable using in-memory fake implementations.

Example:

```text
FakeStore
FakeClipboard
DeterministicTestRandomSource
```

Cryptographic production randomness must never be replaced outside tests.

---

# 166. CLI Tests

CLI tests should verify:

```text
arguments
exit codes
stdout
stderr
secret redaction
JSON behavior
```

---

# 167. TUI Tests

Bubble Tea update logic should be testable without rendering a real terminal.

Test:

```text
screen transitions
key bindings
lock transitions
secret reveal timeout
confirmation flows
auto-lock messages
```

---

# 168. Snapshot Tests

TUI views may use snapshot/golden tests.

Snapshots must never contain real credentials.

---

# 169. Fuzzing Targets

Primary targets:

```text
vault metadata parser
entry serializer
entry deserializer
AAD encoder
backup header parser
backup payload parser
selectors
import parsers
```

---

# 170. Race Testing

CI should regularly run:

```text
go test -race ./...
```

especially because GoVault uses:

```text
clipboard timers
auto-lock timers
Bubble Tea commands
session state
```

---

# 171. Static Analysis

Recommended CI checks:

```text
go vet
staticcheck
gofmt
go test
go test -race
```

Additional security tooling may be introduced later.

---

# 172. Secret Scanner

Repository CI should use secret scanning.

This protects development infrastructure rather than vault contents.

---

# 173. Dependency Review

New dependencies should be justified.

Questions:

```text
Is this necessary?

Does the standard library already solve it?

Does it introduce networking?

Does it use CGO?

Does it significantly increase attack surface?

Is it maintained?

Can the functionality be isolated?
```

---

# 174. SQLite Driver Decision

The SQLite driver is an important architectural dependency.

Evaluation criteria:

```text
CGO requirements
cross-platform builds
maintenance
SQLite feature support
binary size
security updates
transaction support
WAL support
```

The final driver choice should be documented in an ADR.

---

# 175. Architecture Decision Records

Important decisions should be documented under:

```text
docs/adr/
```

Example:

```text
0001-sqlite-driver.md
0002-entry-serialization.md
0003-clipboard-strategy.md
0004-process-locking.md
```

---

# 176. Why ADRs

The specifications describe the system.

ADRs explain why a particular implementation choice was selected among alternatives.

---

# 177. Build Architecture

GoVault should build as a single executable where practical:

```text
govault
```

Benefits:

```text
simple installation
simple offline distribution
no daemon
no background service
```

---

# 178. No Daemon

GoVault v1 does not require a persistent background daemon.

Auto-lock and clipboard timers exist only while the process is running.

---

# 179. CLI Session Model

Non-interactive CLI commands generally:

```text
start
unlock if required
perform one operation
clean up
exit
```

---

# 180. TUI Session Model

TUI:

```text
start
unlock
remain active
auto-lock when needed
unlock again
exit
```

---

# 181. Master Password Prompt

Password input should use terminal-safe hidden input.

Do not accept the master password as a normal command-line flag such as:

```text
--password mysecret
```

because command-line arguments may be visible to other processes or shell history.

---

# 182. Environment Variables

GoVault should not encourage storing the master password in environment variables.

Automated unlock is outside MVP scope.

---

# 183. Stdin Password Input

If future automation requires password input through stdin or a file descriptor, it must be explicitly threat-modeled.

Not required for v0.1.

---

# 184. Clipboard vs Stdout

Default secret retrieval should favor:

```text
clipboard
```

over:

```text
stdout
```

for interactive usage.

Both have platform-specific risks, but accidental terminal history/screen exposure is reduced.

---

# 185. TUI Secret State

A screen showing a revealed secret should mark the state clearly:

```text
SECRET VISIBLE
```

This status should disappear when the value is masked again.

---

# 186. Screen Transition Secret Cleanup

Leaving a screen containing revealed data must immediately discard its reveal state.

---

# 187. Resize Handling

Terminal resize must not accidentally expose hidden content.

The render model should always derive visibility from explicit state rather than previous terminal contents.

---

# 188. Terminal Clear on Lock

Lock should request a full screen clear.

This reduces exposure from previous rendered content.

It cannot guarantee removal from terminal scrollback on every terminal emulator.

---

# 189. Scrollback Limitation

GoVault must not claim that terminal history can always be securely erased.

Terminal behavior is outside full application control.

---

# 190. First-Run Architecture

If no vault exists:

```text
CLI root
  ↓
detect no vault
  ↓
initialization workflow
  ↓
create master password
  ↓
generate Vault Key
  ↓
derive KEK
  ↓
wrap Vault Key
  ↓
create SQLite vault
  ↓
unlock session
```

---

# 191. Vault Initialization

Initialization should be atomic where practical.

A failed setup must not leave a vault that appears complete but cannot unlock.

---

# 192. Initialization Temporary File

A safe approach:

```text
create temporary vault
initialize schema
write metadata
validate
rename to final vault path
```

---

# 193. Change Master Password Architecture

```text
User
 │
 ▼
verify current password
 │
 ▼
recover Vault Key
 │
 ▼
validate new password
 │
 ▼
generate new salt
 │
 ▼
derive new KEK
 │
 ▼
generate new wrap nonce
 │
 ▼
wrap same Vault Key
 │
 ▼
atomic metadata update
```

Entries are not re-encrypted.

---

# 194. Audit Architecture

```text
Storage
  │
  ▼
Decrypt entries one at a time
  │
  ▼
Extract audit-relevant fields
  │
  ▼
Audit engine
  │
  ▼
Transient results
```

Avoid keeping every full decrypted entry alive simultaneously unless necessary.

---

# 195. Password Reuse Analysis

A possible processing strategy:

```text
entry
  ↓
password
  ↓
transient keyed digest
  ↓
reuse grouping
```

The digest key exists only for the audit run.

Results are discarded afterward.

---

# 196. TOTP Architecture

TOTP generation should receive the decrypted TOTP configuration only when required.

The long-term secret remains encrypted at rest.

---

# 197. TOTP UI Timer

Bubble Tea may issue periodic tick messages for countdown display.

These ticks must not reset auto-lock activity.

---

# 198. Import Architecture

Future imports should follow:

```text
external local file
  ↓
parser
  ↓
validation
  ↓
normalized domain entries
  ↓
encryption
  ↓
storage
```

---

# 199. Import Is Local Only

Import parsers operate only on user-selected local files.

They do not fetch remote exports.

---

# 200. Import Staging

Before commit, imported records should be previewable.

Example future screen:

```text
◆ Import Preview

196 records found

✓ 180 valid
⚠ 12 duplicates
⚠ 4 missing passwords

[ Review ]   [ Import ]
```

---

# 201. Bulk Operations

Bulk tagging or deletion should be application-level operations.

Destructive bulk operations require stronger confirmation.

---

# 202. No Plugin Execution in Vault Data

Entry data must never contain executable GoVault plugins, scripts, or hooks.

Vault content is data only.

---

# 203. No Shell Expansion

Fields such as:

```text
website
username
host
notes
```

must not be interpolated into shell commands automatically.

---

# 204. External Command Execution

Clipboard implementations may require platform utilities on some systems.

If external commands are used:

```text
arguments must be fixed
secrets should use stdin where appropriate
shell interpretation must be avoided
```

Never construct:

```text
sh -c "echo <secret> | ..."
```

---

# 205. Clipboard Adapter Examples

Platform implementations may differ:

```text
macOS
pbcopy / pbpaste

Linux
Wayland/X11-specific adapter

Windows
native/platform adapter
```

Exact implementation belongs in an ADR.

---

# 206. Clipboard Capability Detection

GoVault should detect unavailable clipboard support.

Fallback:

```text
Clipboard integration unavailable.
```

It should not automatically print the secret instead.

---

# 207. Feature Capability Model

Platform-dependent features can expose capabilities.

Conceptually:

```go
type Capabilities struct {
    Clipboard bool
}
```

The TUI can disable unavailable actions gracefully.

---

# 208. Application Status

The application may expose:

```go
type Status struct {
    VaultExists bool
    Locked      bool
    EntryCount  int
}
```

Sensitive state must not be included while locked.

---

# 209. `govault status`

Example locked output:

```text
Vault       found
Status      locked
Format      v1
```

Do not reveal entry names.

Whether entry count is shown while locked should follow metadata-leakage policy.

---

# 210. Command Palette Architecture

The command palette should invoke registered application actions rather than duplicate navigation logic.

Conceptually:

```go
type Action struct {
    ID       string
    Label    string
    Shortcut string
    Run      func() tea.Cmd
}
```

---

# 211. Action Availability

Actions may depend on state.

Example:

```text
Create Entry
requires unlocked vault

Unlock
requires locked vault

Restore
may be available while locked
```

---

# 212. Screen Navigation

Use explicit navigation state rather than arbitrary screen-to-screen imports.

Possible pattern:

```text
root model
→ receives navigation message
→ switches current screen
```

This keeps screen dependencies manageable.

---

# 213. TUI Messages

Application results can be represented as typed Bubble Tea messages.

Examples:

```text
VaultUnlockedMsg
VaultLockedMsg
EntryLoadedMsg
EntrySavedMsg
PasswordCopiedMsg
BackupCreatedMsg
AuditCompletedMsg
AppErrorMsg
```

---

# 214. Avoid Secret-Bearing Messages

Bubble Tea messages should not unnecessarily carry plaintext secrets through long-lived application state.

When unavoidable, scope them tightly.

---

# 215. TUI Commands

Slow operations should run through Bubble Tea commands.

Examples:

```text
unlock
backup
restore
audit
doctor
```

The UI remains responsive.

---

# 216. Loading States

Every long operation should have an explicit state.

Example:

```text
◆ Creating Backup

Encrypting vault...

████████████░░░░░░
```

Progress should only be shown if meaningful progress can actually be measured.

---

# 217. No Fake Progress

If progress is unknown, use a spinner rather than fabricated percentages.

---

# 218. Concurrency

Concurrency should be introduced only where useful.

Password-manager workloads do not require aggressive parallelism.

Security-sensitive code benefits from simple control flow.

---

# 219. Search Index Concurrency

Index construction may initially remain sequential.

Optimize only after profiling.

---

# 220. Database Connection Pool

SQLite does not need a large connection pool for GoVault.

A conservative connection strategy simplifies locking.

The exact pool configuration depends on the selected driver.

---

# 221. Write Serialization

Application writes may be serialized through storage/SQLite locking.

GoVault prioritizes consistency over high write throughput.

---

# 222. Process Locking

A process-level vault lock should be considered to prevent two interactive GoVault processes from modifying the same vault simultaneously.

This requires an ADR.

---

# 223. Read-Only Processes

Future architecture may allow concurrent read-only operations.

Not required for MVP.

---

# 224. Version Package

Build/version metadata may live in:

```text
internal/version
```

or be injected through build flags.

It should provide:

```text
application version
commit
build date
```

No remote version checking occurs.

---

# 225. `govault version`

Example:

```text
GoVault 0.1.0
Vault Format: 1
Backup Format: 1
Crypto Suite: 1
```

---

# 226. Build Reproducibility

The project should aim for reproducible builds where practical.

Avoid embedding machine-specific absolute paths or timestamps unless intentionally configured.

---

# 227. Release Artifacts

Expected release artifacts may include:

```text
govault-linux-amd64
govault-linux-arm64
govault-darwin-arm64
govault-darwin-amd64
govault-windows-amd64.exe
```

Actual supported platforms depend on SQLite and clipboard choices.

---

# 228. No Installer Requirement

The core application should remain usable as a standalone executable.

---

# 229. Documentation Architecture

Technical documentation should remain close to the codebase.

Recommended:

```text
docs/
├── prd.md
├── threat-model.md
├── cryptography.md
├── vault-format.md
├── backup-format.md
├── architecture.md
├── implementation-plan.md
└── adr/
```

---

# 230. Architecture Invariants

The following invariants are mandatory.

## ARCH-001

CLI and TUI use the same application services.

## ARCH-002

Presentation layers do not implement cryptography.

## ARCH-003

SQLite does not receive sensitive plaintext.

## ARCH-004

Storage persists encrypted envelopes.

## ARCH-005

Searchable plaintext metadata is memory-only.

## ARCH-006

Session keys are inaccessible to presentation code.

## ARCH-007

Backup format is independent of SQLite layout.

## ARCH-008

Network functionality is absent from the core architecture.

## ARCH-009

Master passwords are never accepted through ordinary CLI flags.

## ARCH-010

Secrets are not logged.

## ARCH-011

Destructive operations are explicit.

## ARCH-012

Restore validates before replacing the active vault.

## ARCH-013

Lock destroys unlocked application state where practical.

## ARCH-014

All cryptographic randomness uses `crypto/rand`.

## ARCH-015

Unknown format versions fail safely.

---

# 231. Architecture Review Checklist

Every significant feature should answer:

```text
Which layer owns this?

Does this expose plaintext to storage?

Does this expose keys to presentation?

Does this introduce a network dependency?

Does this duplicate CLI/TUI logic?

Does this require persistent decrypted metadata?

Does this introduce a new dependency?

Does this change the vault format?

Does this change the backup format?

Does this change the crypto suite?

Does this need an ADR?

Can it be tested without a real terminal?

Can it be tested without a real user vault?
```

---

# 232. MVP Package Set

For GoVault v0.1, the minimum practical package set is:

```text
internal/app
internal/domain
internal/crypto
internal/serialization
internal/storage/sqlite
internal/generator
internal/clipboard
internal/backup
internal/config
internal/cli
internal/tui
```

Potentially postpone:

```text
internal/audit
internal/totp
```

if MVP scope needs to be reduced.

---

# 233. MVP Screen Set

The minimum TUI should include:

```text
First Run
Lock
Search / Vault List
Entry Detail
New Login
Edit Login
Password Generator
Delete Confirmation
Backup
Restore
Help
```

---

# 234. Recommended v0.1 Features

```text
Vault initialization
Master password unlock
Argon2id
Vault Key hierarchy
XChaCha20-Poly1305
SQLite persistence
Login CRUD
Fuzzy search
Password generator
Clipboard copy
Clipboard timeout
Auto-lock
Encrypted backup
Restore
Cobra CLI
Bubble Tea TUI
```

---

# 235. Recommended v0.2 Features

```text
Tags
Favorites
Custom fields
Secure notes
History
Doctor
Advanced search
Settings TUI
```

---

# 236. Recommended v0.3 Features

```text
TOTP
Passphrase generator
Aliases
Security audit
API credentials
Database credentials
```

---

# 237. Recommended v0.4 Features

```text
SSH keys
Import
Migration UX
Bulk operations
Additional backup tooling
UI polish
```

---

# 238. Architecture Decision Gates

Before implementation begins, resolve:

```text
SQLite driver
canonical serialization format
clipboard strategy
process locking
configuration format
exact KDF defaults
exact storage size limits
```

These do not all need lengthy documents, but the decisions should be recorded.

---

# 239. Recommended Implementation Order

Implementation should proceed from the center outward:

```text
domain
   │
   ▼
serialization
   │
   ▼
crypto
   │
   ▼
storage
   │
   ▼
application services
   │
   ├────► CLI
   │
   └────► TUI
```

Backup can begin once crypto, serialization, and storage boundaries are stable.

---

# 240. Why TUI Should Not Be Built First

The TUI is important, but building it before stable application services tends to place business logic inside Bubble Tea models.

Instead:

```text
core use cases
→ CLI validation
→ TUI integration
```

produces a cleaner architecture.

---

# 241. CLI as Early Integration Harness

The CLI provides a simple way to validate the application core before the full TUI exists.

Example early milestones:

```text
govault init
govault add
govault list
govault show
govault copy
```

Once those use cases work, the TUI can invoke the same services.

---

# 242. Security-Critical Code Review Areas

The highest-priority review areas are:

```text
internal/crypto
internal/serialization
internal/backup
vault initialization
unlock
master password change
restore
clipboard
```

---

# 243. Smaller Trusted Core

GoVault should aim to keep its security-sensitive trusted code base relatively small.

UI rendering, styling, and command routing should remain outside that trusted core.

---

# 244. Final Architecture

The resulting system should resemble:

```text
                       ┌─────────────┐
                       │    User     │
                       └──────┬──────┘
                              │
                  ┌───────────┴───────────┐
                  ▼                       ▼
             Cobra CLI               Bubble Tea
                  │                       │
                  └───────────┬───────────┘
                              ▼
                     Application Layer
                              │
          ┌───────────────────┼───────────────────┐
          │                   │                   │
          ▼                   ▼                   ▼
       Domain              Backup             Generator
          │
          ▼
     Serialization
          │
          ▼
        Crypto
          │
          ▼
Encrypted Envelopes
          │
          ▼
        Storage
          │
          ▼
        SQLite
```

Supporting adapters:

```text
Application
    │
    ├── Clipboard
    ├── Configuration
    ├── Filesystem
    └── Local Clock
```

There is intentionally no:

```text
HTTP Client
Cloud Client
Telemetry Client
Analytics Service
Update Service
Remote API
```

---

# 245. Architecture Summary

GoVault is a local, terminal-native application with a shared core beneath two interfaces:

```text
Cobra CLI
Bubble Tea TUI
```

Its most important boundary is:

```text
plaintext domain data
        │
        ▼
serialization
        │
        ▼
cryptography
        │
        ▼
encrypted envelopes
        │
        ▼
SQLite
```

The inverse boundary is:

```text
SQLite
   │
   ▼
encrypted envelope
   │
   ▼
authentication/decryption
   │
   ▼
strict deserialization
   │
   ▼
domain object
```

This boundary must remain intact as the project grows.

---

# 246. Next Document

The next document should be:

```text
docs/implementation-plan.md
```

Unlike the previous specifications, the implementation plan should be execution-oriented.

It should define:

```text
development phases
milestones
tasks
dependencies between tasks
acceptance criteria
test requirements
security gates
recommended commit sequence
MVP definition
v0.1 exit criteria
```

The document should turn:

```text
architecture
```

into:

```text
an ordered engineering plan
```

that can be implemented incrementally without designing critical security behavior during coding.

---

# 247. Document Sequence

The project specification sequence is now:

```text
PRD
 │
 ▼
threat-model.md
 │
 ▼
cryptography.md
 │
 ▼
vault-format.md
 │
 ▼
backup-format.md
 │
 ▼
architecture.md
 │
 ▼
implementation-plan.md
 │
 ▼
ADR decisions
 │
 ▼
implementation
 │
 ▼
security review
 │
 ▼
GoVault v0.1
```

---

# 248. Architecture Completion Criteria

This architecture should be considered ready for implementation when:

```text
✓ SQLite driver has been selected

✓ Canonical serialization has been selected

✓ KDF defaults have been finalized

✓ Clipboard strategy has been selected

✓ Process-locking decision has been made

✓ Core interfaces have been reviewed

✓ Package dependency graph has been validated

✓ Vault and backup formats are sufficiently stable

✓ MVP scope has been frozen

✓ Security invariants have corresponding tests planned
```

Once these decisions are complete, implementation can begin without requiring major architectural decisions inside feature code.

