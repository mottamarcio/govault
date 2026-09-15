# GoVault Vault Format Specification

**Document:** `docs/vault-format.md`
**Project:** GoVault
**Status:** Draft / Pre-v1 Storage Specification
**Applies to:** GoVault v0.1+
**Last Updated:** September 2026

---

# 1. Purpose

This document defines the persistent vault format used by GoVault.

It translates the cryptographic architecture defined in:

```text
docs/cryptography.md
```

into a concrete SQLite representation.

This specification defines:

* SQLite schema.
* Vault metadata layout.
* Encrypted record envelopes.
* Plaintext versus encrypted fields.
* Record identifiers.
* Cryptographic version storage.
* Canonical Associated Data inputs.
* Entry serialization boundaries.
* History storage.
* Alias storage.
* Migration strategy.
* Crash-safe updates.
* Integrity expectations.
* Search-index reconstruction.
* Storage limits.
* Corruption behavior.

This document does not define the `.gvault` backup file format.

That belongs in:

```text
docs/backup-format.md
```

---

# 2. Design Goals

The GoVault vault format must provide:

1. Encrypted persistence of sensitive data.
2. Independent encryption of records.
3. Minimal useful metadata leakage.
4. Safe schema migration.
5. Safe cryptographic migration.
6. Atomic updates.
7. Crash resilience.
8. Efficient local search after unlock.
9. Clear compatibility boundaries.
10. Stable representation independent of Go struct layout.
11. Safe handling of malformed or attacker-controlled databases.

---

# 3. Storage Engine

GoVault uses SQLite as the local persistence engine.

SQLite is responsible for:

* Durable storage.
* Transactions.
* Indexes.
* Schema constraints.
* Crash recovery.
* Local querying of non-secret routing data.

SQLite is not considered an encryption boundary.

Sensitive data must already be encrypted before reaching SQLite.

---

# 4. Database File

The primary vault database is conceptually:

```text
vault.db
```

Platform-specific storage locations must follow operating-system conventions.

Example on Linux:

```text
~/.local/share/govault/vault.db
```

---

# 5. Database Journal Mode

GoVault should use SQLite WAL mode unless testing or platform constraints justify another choice.

Conceptually:

```sql
PRAGMA journal_mode = WAL;
```

WAL provides good transactional behavior and performance.

Because all sensitive records are encrypted before persistence, WAL files must contain ciphertext rather than plaintext secrets.

---

# 6. SQLite Security Requirements

Sensitive values must never be passed to SQLite in plaintext.

This includes:

```text
passwords
TOTP seeds
secure note bodies
API secrets
SSH private keys
entry names
usernames
URLs
tags
aliases
custom field values
entry history payloads
```

unless explicitly reclassified in a future format version.

---

# 7. Database-Level Versioning

The database maintains distinct version numbers.

At minimum:

```text
schema_version
vault_format_version
crypto_suite_version
entry_serialization_version
```

These version numbers are not interchangeable.

---

# 8. Initial Versions

For the initial format:

```text
Schema Version               1
Vault Format Version         1
Crypto Suite Version         1
Entry Serialization Version  1
```

---

# 9. Top-Level Tables

Proposed initial schema:

```text
vault_meta
entries
entry_history
schema_migrations
```

Optional future tables:

```text
audit_cache
settings
tombstones
migration_state
```

should not be added unless justified.

---

# 10. `vault_meta`

`vault_meta` contains the information required to identify and unlock the vault.

Conceptual schema:

```sql
CREATE TABLE vault_meta (
    id                      INTEGER PRIMARY KEY CHECK (id = 1),

    vault_format_version    INTEGER NOT NULL,
    schema_version          INTEGER NOT NULL,
    crypto_suite_version    INTEGER NOT NULL,

    vault_id                BLOB NOT NULL,

    kdf_algorithm           INTEGER NOT NULL,
    kdf_salt                BLOB NOT NULL,
    kdf_memory_kib          INTEGER NOT NULL,
    kdf_iterations          INTEGER NOT NULL,
    kdf_parallelism         INTEGER NOT NULL,

    key_wrap_nonce          BLOB NOT NULL,
    wrapped_vault_key       BLOB NOT NULL,

    created_at_unix         INTEGER NOT NULL,
    updated_at_unix         INTEGER NOT NULL
);
```

Exactly one row must exist.

---

# 11. Why `vault_meta` Is Plaintext

The fields in `vault_meta` are required before the Vault Key can be recovered.

Therefore they cannot themselves depend on the Vault Key for decryption.

They are not intended to contain user-facing secret metadata.

---

# 12. `vault_id`

`vault_id` is a random 16-byte identifier.

Constraints:

```text
length = 16 bytes
immutable
generated once
```

It is used in cryptographic domain separation and Associated Data.

---

# 13. KDF Fields

For Crypto Suite v1:

```text
kdf_algorithm = Argon2id
kdf_salt = 32 random bytes
kdf_memory_kib = configurable within allowed bounds
kdf_iterations = configurable within allowed bounds
kdf_parallelism = configurable within allowed bounds
```

KDF fields must be validated before use.

---

# 14. KDF Algorithm Encoding

Do not store the KDF algorithm as free-form text.

Use a stable numeric identifier.

Example:

```text
1 = Argon2id
```

Unknown identifiers must cause safe failure.

---

# 15. Crypto Suite Encoding

Use stable numeric identifiers.

Example:

```text
1 = GoVault Crypto Suite v1
```

Unknown suites must not be interpreted heuristically.

---

# 16. Wrapped Vault Key

`wrapped_vault_key` stores:

```text
AEAD ciphertext + authentication tag
```

for the encrypted Vault Key.

The nonce is stored separately in:

```text
key_wrap_nonce
```

---

# 17. Wrapped Vault Key Size

For XChaCha20-Poly1305:

```text
Vault Key plaintext       32 bytes
Authentication tag        16 bytes
Expected ciphertext       48 bytes
```

The implementation should validate exact length for Crypto Suite v1.

---

# 18. `key_wrap_nonce`

For Crypto Suite v1:

```text
length = 24 bytes
```

Any other length must be rejected.

---

# 19. Vault Metadata AAD

The wrapped Vault Key is authenticated with canonical metadata.

Conceptually:

```text
protocol_id
vault_format_version
crypto_suite_version
vault_id
kdf_algorithm
kdf_salt
kdf_memory_kib
kdf_iterations
kdf_parallelism
```

The exact canonical binary encoding must not depend on SQL row order or textual formatting.

---

# 20. Entry Table

Proposed schema:

```sql
CREATE TABLE entries (
    entry_id                BLOB PRIMARY KEY,

    record_kind             INTEGER NOT NULL,
    crypto_suite_version    INTEGER NOT NULL,
    serialization_version   INTEGER NOT NULL,

    nonce                   BLOB NOT NULL,
    ciphertext              BLOB NOT NULL,

    created_at_unix         INTEGER NOT NULL,
    updated_at_unix         INTEGER NOT NULL
);
```

---

# 21. `entry_id`

`entry_id` is:

```text
16 random bytes
```

It is immutable.

It serves as:

* Primary identifier.
* AAD input.
* Record reference.
* History parent key.

---

# 22. Why `entry_id` Is Plaintext

The database needs a stable row identity.

The identifier itself is random and contains no human-readable information.

This leakage is accepted.

---

# 23. `record_kind`

`record_kind` indicates how the decrypted payload should be interpreted.

Initial values:

```text
1 = Login
2 = Secure Note
3 = API Credential
4 = Database Credential
5 = Custom
```

Future values may include:

```text
6 = SSH Key
```

Unknown record kinds must not be interpreted as another known type.

---

# 24. Record Kind Leakage

Keeping `record_kind` plaintext leaks the rough category of an entry.

Example:

```text
Login
Secure Note
```

This is considered acceptable for Vault Format v1 because it simplifies routing and validation.

If stronger metadata confidentiality is desired later, a new vault format may move this information inside the encrypted payload.

---

# 25. Timestamps

`created_at_unix` and `updated_at_unix` are stored as Unix timestamps.

Recommended representation:

```text
signed 64-bit integer
UTC seconds
```

Millisecond precision is not required initially.

---

# 26. Timestamp Leakage

Plaintext timestamps reveal activity patterns.

This is an accepted tradeoff for Vault Format v1.

Human-readable labels remain encrypted.

---

# 27. Entry Encryption Envelope

Each row contains:

```text
entry_id
record_kind
crypto_suite_version
serialization_version
nonce
ciphertext
created_at_unix
updated_at_unix
```

Only:

```text
ciphertext
```

contains user-facing content.

---

# 28. Entry AAD

Entry Associated Data is derived from stable routing metadata.

Conceptually:

```text
protocol_id
vault_id
entry_id
record_kind
crypto_suite_version
serialization_version
```

Timestamps are not required in AAD for v1.

---

# 29. Why Timestamps Are Excluded From AAD

Including mutable timestamps in AAD would require careful synchronization every time a record changes.

Since updates already produce a new ciphertext, timestamp tampering is considered lower risk than secret substitution.

This can be revisited in a future format.

---

# 30. Ciphertext Storage

SQLite stores the AEAD output as raw binary:

```text
BLOB
```

Do not base64-encode encrypted data inside SQLite unless required by an external interoperability constraint.

---

# 31. Nonce Storage

Nonces are stored as raw binary:

```text
BLOB
```

Expected length:

```text
24 bytes
```

for Crypto Suite v1.

---

# 32. Entry Plaintext Serialization

The encrypted entry payload must use a deterministic, documented serialization format.

The preferred direction for Vault Format v1 is a compact explicit binary format.

Alternative acceptable choice:

```text
deterministic CBOR
```

if the implementation can guarantee stable canonical encoding.

Plain JSON is not preferred for the encrypted canonical format because of ambiguity around field ordering, number encoding, and unnecessary overhead.

---

# 33. Go Struct Layout Is Not a Format

The vault format must never depend on:

```text
unsafe memory layout
gob
raw Go struct serialization
reflection-dependent field order
```

A future Go compiler or architecture must not change persistent data interpretation.

---

# 34. Entry Payload Versioning

Every encrypted entry declares:

```text
serialization_version
```

This allows new payload fields without requiring a new crypto suite.

---

# 35. Login Payload

Conceptually:

```text
LoginEntryV1 {
    name
    username
    password
    website
    notes
    tags[]
    totp?
    favorite
    aliases[]
    custom_fields[]
}
```

All values above are encrypted.

---

# 36. Secure Note Payload

Conceptually:

```text
SecureNoteV1 {
    title
    content
    tags[]
    favorite
    aliases[]
    custom_fields[]
}
```

---

# 37. API Credential Payload

Conceptually:

```text
APICredentialV1 {
    name
    key
    secret
    environment
    notes
    tags[]
    aliases[]
    custom_fields[]
}
```

---

# 38. Database Credential Payload

Conceptually:

```text
DatabaseCredentialV1 {
    name
    host
    port
    database
    username
    password
    environment
    notes
    tags[]
    aliases[]
    custom_fields[]
}
```

---

# 39. Custom Fields

Custom fields should conceptually support:

```text
name
value
sensitive
```

Example:

```text
{
    name: "Account Number",
    value: "...",
    sensitive: true
}
```

Both field name and value remain inside the encrypted payload.

---

# 40. TOTP Payload

TOTP data should conceptually contain:

```text
secret
issuer
account_name
algorithm
digits
period
```

All fields are encrypted.

---

# 41. TOTP Algorithm Identifiers

Use stable numeric identifiers, not arbitrary strings.

For example:

```text
1 = SHA1
2 = SHA256
3 = SHA512
```

Unsupported algorithms must fail safely.

---

# 42. Tags

Tags are stored inside the encrypted entry payload.

This means SQLite cannot query tags directly while locked.

This is intentional.

---

# 43. Aliases

Aliases are also stored inside encrypted payloads in Vault Format v1.

Advantages:

* No plaintext alias leakage.
* No extra alias table.
* Simpler confidentiality model.

Tradeoff:

* Alias resolution requires unlock.

This is acceptable because the vault cannot expose secrets while locked anyway.

---

# 44. Favorite State

Favorite status is encrypted.

This prevents leakage about which entries are important to the user.

---

# 45. Search Index

GoVault builds a transient in-memory search index after successful unlock.

The index may contain:

```text
entry_id
record_kind
name
username
website
tags
aliases
favorite
```

It must not include:

```text
password
TOTP secret
API secret
private key
secure note body
```

unless a future search feature explicitly requires it.

---

# 46. Search Index Lifecycle

The search index exists only while the vault is unlocked.

On lock:

```text
clear index
drop references
clear decrypted metadata where practical
```

The index is never persisted to disk.

---

# 47. Search Index Reconstruction

After unlock:

```text
read entries
→ decrypt each entry
→ extract searchable metadata
→ build in-memory index
```

For typical vault sizes this should remain fast.

---

# 48. Large Vault Optimization

If vaults with many tens of thousands of entries become common, future formats may introduce encrypted searchable indexes.

This is explicitly out of scope for Vault Format v1.

---

# 49. Entry History Table

Proposed schema:

```sql
CREATE TABLE entry_history (
    history_id              BLOB PRIMARY KEY,
    entry_id                BLOB NOT NULL,

    crypto_suite_version    INTEGER NOT NULL,
    serialization_version   INTEGER NOT NULL,

    nonce                   BLOB NOT NULL,
    ciphertext              BLOB NOT NULL,

    created_at_unix         INTEGER NOT NULL,

    FOREIGN KEY(entry_id)
        REFERENCES entries(entry_id)
        ON DELETE CASCADE
);
```

---

# 50. `history_id`

`history_id` is:

```text
16 random bytes
```

It is used for:

* Unique version identification.
* AAD binding.
* History inspection.
* Restoration.

---

# 51. History Payload

A history payload stores the previous logical entry state.

Conceptually:

```text
HistoryRecordV1 {
    parent_entry_id
    previous_entry_payload
    change_kind
}
```

---

# 52. History Encryption Key

History uses the dedicated History Key defined in `cryptography.md`.

It must not use the Entry Key directly.

---

# 53. History AAD

Conceptually:

```text
protocol_id
vault_id
entry_id
history_id
crypto_suite_version
serialization_version
```

---

# 54. History Creation

Before updating an existing entry:

```text
decrypt current record
→ encrypt old state as history
→ write history + new entry
→ commit transaction
```

Both operations should occur in the same SQLite transaction.

---

# 55. History Restore

Restoring history does not mutate historical ciphertext directly.

Instead:

```text
decrypt historical version
→ optionally store current version as new history
→ encrypt restored state as current entry
→ commit
```

---

# 56. History Retention

Initial behavior may retain all history.

Future configuration may support:

```text
last N versions
maximum age
disabled
```

Retention policy changes do not require changing the cryptographic format.

---

# 57. Entry Deletion

Deleting an entry should remove:

```text
active entry
associated history
```

depending on configured history policy.

For MVP:

```text
ON DELETE CASCADE
```

is acceptable.

---

# 58. Soft Delete

Vault Format v1 does not require soft deletion.

A recycle bin feature may be introduced later using encrypted tombstone records.

---

# 59. Schema Migration Table

Proposed schema:

```sql
CREATE TABLE schema_migrations (
    version         INTEGER PRIMARY KEY,
    applied_at_unix INTEGER NOT NULL
);
```

This tracks completed schema migrations.

---

# 60. Schema Migration Rules

Migrations must be:

```text
ordered
idempotence-aware
transactional where possible
tested
version-controlled
```

---

# 61. Migration Sequence

Conceptually:

```text
read schema version
→ validate supported range
→ create recovery snapshot if required
→ begin transaction
→ apply migration
→ validate
→ update version
→ commit
```

---

# 62. Migration Failure

On migration failure:

```text
rollback transaction
preserve original vault
report actionable error
```

GoVault must not partially advance schema version state.

---

# 63. Cryptographic Migration

Schema migrations and cryptographic migrations are separate concepts.

A schema migration might alter tables without re-encrypting records.

A cryptographic migration may require:

```text
decrypt
re-serialize
re-encrypt
```

These must be explicitly distinguished in code.

---

# 64. Migration State

Long-running crypto migrations may require resumable state in future versions.

Vault Format v1 does not require resumable crypto migration.

Before v1.0, migration requirements should be reevaluated.

---

# 65. Schema Constraints

SQLite should enforce reasonable constraints.

Examples:

```sql
CHECK(length(entry_id) = 16)
CHECK(length(nonce) = 24)
CHECK(length(vault_id) = 16)
CHECK(length(kdf_salt) = 32)
```

Application validation remains mandatory even when database constraints exist.

---

# 66. Foreign Keys

GoVault should enable:

```sql
PRAGMA foreign_keys = ON;
```

and rely on explicit foreign-key behavior.

---

# 67. Record Size Limits

Attacker-controlled vault files may contain extremely large BLOB lengths.

GoVault must enforce application-level size limits before allocating or decrypting.

Suggested initial limits:

```text
Max encrypted entry           16 MiB
Max secure note               8 MiB
Max custom field value        1 MiB
Max tag length                256 bytes
Max alias length              256 bytes
Max entry name                4 KiB
Max history record            16 MiB
```

These are implementation safety bounds and may be refined.

---

# 68. Entry Count Limits

An attacker may provide a database containing millions of rows.

GoVault should avoid blindly loading unlimited metadata into memory.

Suggested defensive limit for early versions:

```text
maximum supported active entries:
1,000,000
```

The application may reject larger vaults with a clear error.

---

# 69. Tag Count Limit

Suggested:

```text
maximum tags per entry: 256
```

This prevents pathological serialized payloads.

---

# 70. Alias Count Limit

Suggested:

```text
maximum aliases per entry: 128
```

---

# 71. Custom Field Count Limit

Suggested:

```text
maximum custom fields per entry: 256
```

---

# 72. UTF-8 Validation

Human-readable strings should be valid UTF-8.

Malformed UTF-8 must be rejected at deserialization boundaries unless a field is explicitly defined as arbitrary bytes.

---

# 73. Binary Secret Fields

Some future fields may contain binary data.

Example:

```text
SSH private key bytes
binary token
certificate
```

Such fields must be typed explicitly rather than overloaded into text fields.

---

# 74. No Filesystem Paths Inside Entry Payload Semantics

Entry fields may store informational paths, but encrypted entry parsing must never cause automatic filesystem access.

Data must remain inert.

---

# 75. Unknown Fields

Entry serialization should support forward-compatible evolution where possible.

Preferred behavior:

```text
unknown optional field
→ ignore safely
```

but:

```text
unknown required structure
→ reject
```

The exact mechanism depends on chosen serialization format.

---

# 76. Required Fields

For example, a Login entry may require:

```text
name
```

while:

```text
username
password
website
notes
```

may be optional depending on product rules.

Serialization validation should distinguish:

```text
syntactic validity
semantic validity
```

---

# 77. Empty Passwords

The storage format may represent an empty password.

Product-level validation may warn or reject based on entry type.

Storage should not silently transform empty values.

---

# 78. Canonical AAD Encoding

Associated Data must use a deterministic binary encoding.

Recommended conceptual pattern:

```text
field_tag
field_length
field_value
```

or fixed-width encodings where possible.

Example conceptual entry AAD:

```text
"GOVAULT"
u16(vault_format_version)
u16(crypto_suite_version)
16-byte vault_id
16-byte entry_id
u16(record_kind)
u16(serialization_version)
```

Exact endianness and lengths must be specified.

---

# 79. Endianness

Vault Format v1 should use:

```text
big-endian
```

for multi-byte integers in cryptographic canonical encodings.

SQLite integer storage remains SQLite-native.

This distinction must be clear.

---

# 80. Protocol Identifier

Canonical cryptographic data should begin with an explicit protocol marker.

Conceptually:

```text
GOVAULT
```

or a compact binary constant.

This reduces accidental cross-protocol reuse.

---

# 81. No Ambiguous Concatenation

Never build AAD like:

```text
"1" + "23"
```

where ambiguity exists.

Use fixed-width or length-prefixed encoding.

---

# 82. Entry Serialization vs AAD

The encrypted payload and AAD serve different purposes.

AAD contains:

```text
routing/context
```

Encrypted payload contains:

```text
user data
```

Do not duplicate large user-visible fields into AAD.

---

# 83. Created and Updated Times Inside Payload

If stronger metadata integrity is desired, logical timestamps may also exist inside the encrypted payload.

Vault Format v1 may treat SQLite timestamps as operational metadata only.

The product must not rely on them for security decisions.

---

# 84. SQLite `user_version`

GoVault may also set:

```sql
PRAGMA user_version;
```

for diagnostic convenience.

However, the canonical schema version remains stored in `vault_meta` and/or migration records.

Do not rely exclusively on `PRAGMA user_version`.

---

# 85. Database Integrity Check

`govault doctor` may execute:

```sql
PRAGMA integrity_check;
```

or appropriate equivalent.

Passing SQLite structural integrity does not imply cryptographic integrity.

Both are separate checks.

---

# 86. Cryptographic Integrity Check

A full cryptographic integrity check requires:

```text
vault unlocked
→ decrypt/authenticate all records
→ decrypt/authenticate history
```

This may be exposed through:

```bash
govault doctor --deep
```

in a future release.

---

# 87. Locked Doctor Checks

While locked, `doctor` can verify:

```text
database readable
SQLite integrity
schema version
file permissions
vault header lengths
supported crypto suite
KDF bounds
```

It cannot prove encrypted record authenticity without the Vault Key.

---

# 88. Atomic Entry Update

Entry update should use one SQLite transaction.

Conceptually:

```sql
BEGIN;

INSERT INTO entry_history (...);

UPDATE entries
SET nonce = ?,
    ciphertext = ?,
    updated_at_unix = ?
WHERE entry_id = ?;

COMMIT;
```

---

# 89. Atomic Entry Creation

Entry creation:

```sql
BEGIN;

INSERT INTO entries (...);

COMMIT;
```

No partially initialized plaintext row should exist.

---

# 90. Atomic Entry Delete

Deletion should occur within a transaction.

If history is cascaded, SQLite foreign-key handling should be tested.

---

# 91. Master Password Rewrap Storage

Master password changes modify only relevant fields in `vault_meta`.

Conceptually:

```text
kdf_salt
kdf_memory_kib
kdf_iterations
kdf_parallelism
key_wrap_nonce
wrapped_vault_key
updated_at_unix
```

---

# 92. Atomic Master Password Change

Recommended flow:

```text
derive old KEK
unwrap Vault Key
derive new KEK
prepare new wrapped Vault Key
verify in memory
BEGIN IMMEDIATE
update vault_meta
COMMIT
```

If commit fails, the old database state remains valid.

---

# 93. No Intermediate Header State

Do not update:

```text
kdf_salt
```

in one transaction and:

```text
wrapped_vault_key
```

later.

All cryptographically coupled fields must change atomically.

---

# 94. WAL and Master Password Changes

Tests must verify that interrupted password changes do not leave a database that references mismatched:

```text
KDF parameters
salt
wrapped Vault Key
nonce
```

---

# 95. Database Open Validation

Before attempting unlock, GoVault should validate:

```text
SQLite file opens
required tables exist
exactly one vault_meta row exists
supported schema version
supported vault format
supported crypto suite
vault ID length
KDF bounds
salt length
key-wrap nonce length
wrapped key length
```

---

# 96. Unlock Does Not Trust Database Content

Even after Vault Key recovery, every encrypted record remains untrusted until authenticated.

One valid Vault Key does not imply every row is valid.

---

# 97. Entry Load Validation

For every row:

```text
validate entry_id length
validate record kind
validate crypto suite
validate serialization version
validate nonce length
validate ciphertext size
construct AAD
authenticate/decrypt
validate plaintext structure
```

---

# 98. Partial Corruption

A vault may contain:

```text
99 valid records
1 corrupted record
```

GoVault should distinguish:

```text
vault cannot unlock
```

from:

```text
vault unlocked but record corruption detected
```

---

# 99. Corrupted Record UX

Suggested behavior:

```text
Vault unlocked with errors.

1 record could not be authenticated.

Run:
govault doctor --deep
```

GoVault must not display corrupted plaintext.

---

# 100. Whether to Continue After Record Corruption

For read-only browsing, GoVault may allow access to valid records while clearly reporting corrupted ones.

Write operations should be more conservative.

The exact product policy should be documented.

---

# 101. Duplicate Entry IDs

Because `entry_id` is the primary key, SQLite prevents duplicates in normal operation.

A malformed database violating schema assumptions must be rejected.

---

# 102. Duplicate Aliases

Aliases are decrypted in memory.

GoVault must detect ambiguous aliases during index construction.

Example:

```text
gh → Entry A
gh → Entry B
```

This must be rejected or surfaced as a conflict.

Aliases must never resolve nondeterministically.

---

# 103. Duplicate Names

Entry names do not need to be unique.

Search and CLI selectors may require explicit disambiguation.

Example:

```text
GitHub
GitHub
```

is valid storage.

---

# 104. CLI Lookup IDs

The CLI may permit lookup by:

```text
exact alias
exact ID prefix
fuzzy name
```

but ambiguity handling belongs to application logic, not storage.

---

# 105. ID Encoding for Display

Internally:

```text
16-byte binary ID
```

For display:

```text
lowercase hex
```

or canonical UUID formatting may be used.

The chosen display encoding does not alter persistent binary storage.

---

# 106. ID Prefixes

CLI commands may accept unique prefixes.

Example:

```bash
govault show a31f
```

The application must reject ambiguous prefixes.

---

# 107. File Permissions

On Unix-like systems:

```text
vault.db       0600
vault.db-wal   inherited/restricted by directory
vault.db-shm   inherited/restricted by directory
```

Parent directories should also be private where practical.

---

# 108. Database File Creation

The vault database should be created with restrictive permissions from the start.

Avoid:

```text
create broadly readable file
→ chmod later
```

where the platform allows secure creation directly.

---

# 109. Symlink Handling

Vault initialization should avoid writing through unexpected symlinks when practical.

The canonical application data directory should be controlled by the user.

---

# 110. SQLite Extensions

GoVault must not load arbitrary SQLite extensions.

Dynamic extension loading should remain disabled.

---

# 111. SQL Injection

All dynamic data must use parameterized SQL.

Never interpolate:

```text
entry IDs
timestamps
ciphertext
record types
```

into SQL strings.

---

# 112. No User SQL

GoVault does not expose raw SQL execution against the vault.

---

# 113. No Plaintext FTS

SQLite FTS indexes must not contain decrypted entry metadata in Vault Format v1.

Fuzzy search remains in memory.

---

# 114. Vacuum

SQLite maintenance operations such as:

```sql
VACUUM;
```

may rewrite ciphertext pages.

This does not decrypt content.

`VACUUM` must not be presented as secure deletion.

---

# 115. Secure Delete Claims

GoVault must not claim that deleting an entry securely erases all previous encrypted bytes from physical storage.

Backups, SSD wear leveling, filesystem snapshots, and SQLite page reuse make such guarantees difficult.

Because content is encrypted at rest, residual bytes should remain ciphertext.

---

# 116. Database Copying

Copying `vault.db` while GoVault is running may produce inconsistent results depending on SQLite state.

Users should use:

```text
GoVault encrypted backup
```

rather than manually copying database files.

---

# 117. Manual Database Copy Disclaimer

A manual copy of:

```text
vault.db
```

is not considered an official backup mechanism.

Especially when WAL mode is active, the WAL may contain required committed state.

---

# 118. Official Backup Boundary

`docs/backup-format.md` will define the supported portable backup format.

The application should encourage users to use:

```bash
govault backup create ...
```

---

# 119. Vault Export

Plaintext export is not part of Vault Format v1.

If implemented later, it is an application-layer feature and not part of persistent vault storage.

---

# 120. Import

Imported data must be transformed into normal encrypted GoVault entries before persistence.

Imported plaintext must not remain in staging tables.

---

# 121. Temporary Import State

Import flows should remain memory-based where practical.

If temporary storage becomes necessary in the future, it must be separately threat-modeled.

---

# 122. Corruption Classification

GoVault should distinguish:

```text
SQLite structural corruption
unsupported schema
unsupported crypto suite
invalid vault header
record authentication failure
entry serialization failure
semantic validation failure
```

User messages may simplify these categories, but internal diagnostics should preserve them.

---

# 123. Error Codes

Internal typed errors are preferred.

Conceptually:

```text
ErrUnsupportedSchema
ErrUnsupportedVaultFormat
ErrUnsupportedCryptoSuite
ErrInvalidVaultHeader
ErrAuthenticationFailed
ErrCorruptRecord
ErrInvalidSerialization
ErrKDFBounds
```

Do not rely on string matching for security-sensitive error handling.

---

# 124. Schema Compatibility Policy

Before v1.0:

```text
breaking format changes are allowed
```

provided migration tooling is updated.

After v1.0:

```text
format compatibility becomes a stable contract
```

and breaking changes require explicit migrations.

---

# 125. Unknown Schema Version

If the database schema version is newer than the application supports:

```text
refuse to open for writing
```

Recommended user message:

```text
This vault was created by a newer version of GoVault.
Upgrade GoVault before opening it.
```

---

# 126. Older Schema Version

If migration is supported:

```text
offer/perform migration
```

with recovery safeguards.

---

# 127. Read-Only Future Compatibility

Future versions may support limited read-only access to some older formats.

Vault Format v1 does not require this behavior.

---

# 128. Entry Serialization Migration

If only payload serialization changes:

```text
serialization v1 → v2
```

the application may migrate entries individually.

This does not necessarily require changing:

```text
crypto suite
vault format
schema version
```

---

# 129. Cryptographic Format Migration

If AEAD or key hierarchy changes, the crypto suite version must change.

This is more significant than ordinary payload migration.

---

# 130. Vault Format Migration

Vault format version changes when the persistent vault contract changes materially.

Examples:

```text
new header architecture
encrypted routing metadata
new record envelope format
```

---

# 131. Backup Format Is Independent

The backup format has its own version.

Example:

```text
Vault Format   1
Backup Format  2
```

is valid.

Do not couple these unnecessarily.

---

# 132. Recommended SQL Schema v1

A consolidated conceptual schema:

```sql
CREATE TABLE vault_meta (
    id                      INTEGER PRIMARY KEY CHECK (id = 1),

    vault_format_version    INTEGER NOT NULL,
    schema_version          INTEGER NOT NULL,
    crypto_suite_version    INTEGER NOT NULL,

    vault_id                BLOB NOT NULL CHECK(length(vault_id) = 16),

    kdf_algorithm           INTEGER NOT NULL,
    kdf_salt                BLOB NOT NULL CHECK(length(kdf_salt) = 32),
    kdf_memory_kib          INTEGER NOT NULL,
    kdf_iterations          INTEGER NOT NULL,
    kdf_parallelism         INTEGER NOT NULL,

    key_wrap_nonce          BLOB NOT NULL CHECK(length(key_wrap_nonce) = 24),
    wrapped_vault_key       BLOB NOT NULL,

    created_at_unix         INTEGER NOT NULL,
    updated_at_unix         INTEGER NOT NULL
);

CREATE TABLE entries (
    entry_id                BLOB PRIMARY KEY
                            CHECK(length(entry_id) = 16),

    record_kind             INTEGER NOT NULL,
    crypto_suite_version    INTEGER NOT NULL,
    serialization_version   INTEGER NOT NULL,

    nonce                   BLOB NOT NULL
                            CHECK(length(nonce) = 24),

    ciphertext              BLOB NOT NULL,

    created_at_unix         INTEGER NOT NULL,
    updated_at_unix         INTEGER NOT NULL
);

CREATE TABLE entry_history (
    history_id              BLOB PRIMARY KEY
                            CHECK(length(history_id) = 16),

    entry_id                BLOB NOT NULL,

    crypto_suite_version    INTEGER NOT NULL,
    serialization_version   INTEGER NOT NULL,

    nonce                   BLOB NOT NULL
                            CHECK(length(nonce) = 24),

    ciphertext              BLOB NOT NULL,

    created_at_unix         INTEGER NOT NULL,

    FOREIGN KEY(entry_id)
        REFERENCES entries(entry_id)
        ON DELETE CASCADE
);

CREATE INDEX idx_entry_history_entry
ON entry_history(entry_id, created_at_unix DESC);

CREATE TABLE schema_migrations (
    version                 INTEGER PRIMARY KEY,
    applied_at_unix         INTEGER NOT NULL
);
```

This is still a design specification, not final production SQL.

---

# 133. Recommended SQLite Pragmas

Initial direction:

```sql
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA synchronous = FULL;
```

Performance changes must be measured against durability requirements.

Avoid weakening durability casually.

---

# 134. `synchronous`

For a password manager, durability is more important than maximum write throughput.

The initial preference is:

```text
FULL
```

unless platform testing shows a reason to change it.

---

# 135. Busy Timeout

GoVault should configure a reasonable SQLite busy timeout.

This avoids immediate failures if another local operation temporarily holds the database.

---

# 136. Multiple GoVault Processes

Concurrent writers should be rare.

GoVault should handle them safely rather than assume single-process access.

Potential behavior:

```text
SQLite locking + application-level vault lock
```

A later architecture document should define process coordination.

---

# 137. Application Lock File

A local lock file may be introduced to avoid multiple interactive GoVault instances editing the same vault.

This is an application coordination mechanism, not part of cryptographic storage.

---

# 138. Database Integrity on Startup

Basic startup should not decrypt every record.

Recommended sequence:

```text
open SQLite
validate schema
validate vault_meta
unlock
build search index
surface any record errors
```

---

# 139. Lazy vs Eager Decryption

Two approaches exist.

## Eager Metadata Decryption

After unlock, decrypt every entry once to build search index.

Recommended for v1.

## Lazy Record Decryption

Decrypt entries only when opened.

This reduces immediate work but complicates search.

GoVault v1 prefers eager decryption of searchable metadata and lazy retention of full secrets.

---

# 140. Secret Retention

After search-index construction, GoVault should not retain full decrypted entries unnecessarily.

Recommended pattern:

```text
decrypt entry
→ extract searchable metadata
→ discard full plaintext
```

When an entry is opened:

```text
decrypt again
```

This trades some CPU for reduced secret lifetime.

---

# 141. Vault Unlock Performance

For typical vaults:

```text
100–10,000 entries
```

index reconstruction should remain practical.

Performance should be measured before introducing plaintext indexing.

---

# 142. Encryption Overhead

Each encrypted record adds:

```text
24 bytes nonce
16 bytes authentication tag
serialization overhead
SQLite row overhead
```

This is acceptable for password-manager workloads.

---

# 143. Attachments

Binary attachments are not part of Vault Format v1.

Adding attachments would raise:

```text
record size
memory use
backup size
streaming encryption
resource exhaustion
```

concerns and should receive a separate design.

---

# 144. SSH Keys

SSH key storage can initially reuse normal encrypted record payloads.

Large key sizes still fit comfortably within entry limits.

---

# 145. Audit Cache

Security audit results should not be stored in plaintext.

Vault Format v1 does not persist audit results.

They are computed on demand.

---

# 146. Password Reuse Fingerprints

Vault Format v1 must not store persistent password fingerprints.

Reuse detection happens in memory.

---

# 147. Security Score

Security score is recomputed and not persisted as trusted vault state.

---

# 148. Settings

Non-secret application settings belong outside `vault.db` unless there is a clear reason otherwise.

Examples:

```text
clipboard timeout
UI theme
startup screen
auto-lock duration
```

belong in local configuration.

---

# 149. Secret Settings

If a future setting contains sensitive data, it must not be placed in plaintext config.

Such a value likely belongs inside the encrypted vault.

---

# 150. Vault Name

If GoVault later supports named multiple vaults, the vault name may be sensitive.

The default direction should be:

```text
store user-facing vault name encrypted
```

while using a non-secret filesystem identifier externally.

---

# 151. Multiple Vaults

Vault Format v1 can support multiple independent vault files because every vault has its own:

```text
vault_id
Vault Key
KDF salt
wrapped key
```

No global shared secret is required.

---

# 152. Vault Cloning

A raw clone of a vault database preserves the same `vault_id`.

This means both copies are cryptographically the same vault lineage.

If the user wants a new independent vault, GoVault should provide an explicit clone/export workflow that generates a new:

```text
vault_id
Vault Key
nonces
```

and re-encrypts records.

---

# 153. Why Raw Copies Are Not New Vaults

Copying:

```text
vault.db → another-vault.db
```

creates another copy of the same cryptographic vault.

This matters for domain separation and backup lineage.

---

# 154. Future Vault Fork Command

A future command may provide:

```bash
govault fork
```

to create an independent cryptographic vault from current contents.

Not required for MVP.

---

# 155. Restore Behavior

Restoring an official backup may intentionally restore the original:

```text
vault_id
Vault Key
```

because it restores that vault's identity.

---

# 156. Import Behavior

Importing into an existing vault uses the destination vault's:

```text
Vault Key
vault_id
```

Imported records receive new Entry IDs.

---

# 157. Entry Copy Within Vault

Duplicating an entry must generate:

```text
new entry_id
new nonce
new ciphertext
```

It must not duplicate the encrypted envelope byte-for-byte.

---

# 158. Entry Move Between Vaults

Moving an entry between vaults requires:

```text
decrypt under source vault
→ generate destination entry ID
→ encrypt under destination vault
```

Copying ciphertext directly must fail due to vault-bound AAD.

---

# 159. History Across Vaults

History records cannot be copied as raw ciphertext between vaults.

They must be re-encrypted.

---

# 160. Backup Relation

Official backup format will serialize logical vault data rather than depending on raw SQLite page layout.

This gives GoVault freedom to change SQLite schema independently.

---

# 161. Why Backups Should Not Be SQLite Copies

Portable backup should not depend on:

```text
journal mode
WAL state
SQLite implementation details
database page size
```

A logical encrypted container is more stable.

---

# 162. File Magic

SQLite already has its own file header.

GoVault does not need to modify SQLite file magic.

GoVault identifies a valid vault by inspecting expected schema and `vault_meta`.

---

# 163. Database Recognition

A SQLite database is a GoVault vault only if:

```text
required tables exist
vault_meta exists
versions are supported
metadata validates
```

Do not treat any SQLite file as a GoVault vault merely because it opens.

---

# 164. Tampered Database Header

If the SQLite file itself cannot be opened:

```text
report database corruption
```

No cryptographic recovery should be attempted against arbitrary bytes.

---

# 165. Tampered Vault Metadata

If the SQLite structure is valid but `vault_meta` is malformed:

```text
fail before password derivation where possible
```

especially for invalid KDF bounds.

---

# 166. Wrong Master Password vs Tampering

AEAD failure during Vault Key unwrap may mean:

```text
wrong master password
tampered header
corrupted wrapped key
```

Normal UI should not claim certainty.

---

# 167. Deep Doctor

A future deep diagnostic flow may provide more detailed classification after the user confirms.

It still must avoid leaking keys or plaintext.

---

# 168. Storage API Boundary

Application code should not manually build SQL queries throughout the project.

Storage should be behind an interface.

Conceptually:

```go
type VaultStore interface {
    LoadMetadata(ctx context.Context) (...)
    ListEntryEnvelopes(ctx context.Context) (...)
    GetEntryEnvelope(ctx context.Context, id EntryID) (...)
    InsertEntry(...)
    UpdateEntry(...)
    DeleteEntry(...)
    InsertHistory(...)
}
```

---

# 169. Encrypted Storage Boundary

The SQLite storage package should generally receive encrypted envelopes rather than domain plaintext.

Ideal dependency direction:

```text
domain plaintext
    ↓
crypto layer
    ↓
encrypted envelope
    ↓
storage layer
```

---

# 170. Storage Must Not Know Secrets

`internal/storage/sqlite` should not need to understand:

```text
passwords
TOTP secrets
secure note contents
```

It persists opaque encrypted payloads.

---

# 171. Deserialization Boundary

The reverse path:

```text
storage
→ encrypted envelope
→ crypto authentication/decryption
→ serialization parser
→ domain object
```

This ordering matters.

---

# 172. Never Parse Ciphertext as Plaintext

Serialization parsing happens only after successful AEAD authentication.

---

# 173. Memory Allocation Safety

Before decrypting ciphertext, validate:

```text
ciphertext length <= allowed maximum
nonce length correct
record IDs correct
version supported
```

This limits resource-exhaustion attacks.

---

# 174. Post-Decryption Length Validation

Decrypted payloads must also respect expected size bounds.

Authenticated data can still be malformed if produced by a buggy older GoVault version.

---

# 175. Semantic Validation

Example Login validation may ensure:

```text
tag count within limit
alias count within limit
custom field count within limit
valid UTF-8
supported TOTP parameters
```

---

# 176. No Automatic URL Access

A decrypted `website` field remains plain text only.

GoVault must not automatically:

```text
resolve DNS
fetch favicon
open remote metadata
```

Storage content must not trigger network behavior.

---

# 177. No Automatic File Access

Similarly, fields containing paths must not cause automatic file reads.

---

# 178. Testing Requirements

Vault Format tests should include:

```text
create vault
reopen vault
unlock vault
create every entry type
update entries
history creation
history restore
delete entries
migration
master password rewrap
corrupted metadata
corrupted entry ciphertext
corrupted history
unsupported versions
oversized fields
duplicate aliases
invalid UTF-8
```

---

# 179. Known Plaintext Test

Tests must store recognizable secret values.

Then scan:

```text
vault.db
vault.db-wal
vault.db-shm
vault.db-journal
```

for those byte sequences.

Expected:

```text
not found
```

---

# 180. Metadata Leakage Test

Tests should confirm that human-readable metadata such as:

```text
GitHub Personal
john@example.com
production
```

does not appear in raw database bytes when encryption policy says it must remain encrypted.

---

# 181. Cross-Record Substitution Test

Swap:

```text
ciphertext + nonce
```

between two records.

Authentication must fail.

---

# 182. Cross-Vault Substitution Test

Insert a valid encrypted record from Vault A into Vault B.

Authentication must fail.

---

# 183. Record Kind Tampering Test

Change:

```text
record_kind
```

without re-encrypting.

Authentication must fail because record kind is part of AAD.

---

# 184. Serialization Version Tampering Test

Change:

```text
serialization_version
```

without re-encrypting.

Authentication must fail.

---

# 185. Entry ID Tampering Test

Change:

```text
entry_id
```

and preserve ciphertext.

Authentication must fail.

---

# 186. KDF Bounds Test

Set:

```text
kdf_memory_kib = extremely large value
```

The application must reject before allocating memory.

---

# 187. Migration Tests

Every migration should be tested against:

```text
valid old database
partially corrupted database
interrupted migration
already migrated database
unsupported future database
```

---

# 188. Power-Loss Simulation

Where practical, integration tests should simulate interruption during:

```text
entry update
history write
master password change
schema migration
```

and verify a recoverable consistent state.

---

# 189. Fuzz Targets

Strong fuzz targets:

```text
entry serializer/deserializer
AAD encoder
vault metadata validator
record envelope parser
history parser
migration input validator
```

---

# 190. Performance Tests

Measure:

```text
unlock 100 entries
unlock 1,000 entries
unlock 10,000 entries
search-index build
single-entry decrypt
entry update
history write
```

Do not weaken confidentiality merely to optimize before profiling.

---

# 191. Storage Compatibility Test Vectors

Before v1.0, the repository should contain fixtures for:

```text
valid Vault Format v1
valid entries of each type
valid history
wrong master password
corrupted entry
future schema version
future crypto suite
```

---

# 192. Golden Vault Fixtures

Test vaults may contain fixed known cryptographic values for compatibility testing.

They must be clearly marked as non-production fixtures.

---

# 193. Security Review Checklist

Any change to the vault format should answer:

```text
Does this add plaintext metadata?

Does this change AAD?

Does this change key usage?

Does this change nonce behavior?

Does this require migration?

Can this create partial states?

Can this expose secrets in WAL?

Can this increase allocation risk?

Can an attacker modify it pre-unlock?

Is the field validated before use?
```

---

# 194. Decisions Established by Vault Format v1

The proposed v1 design establishes:

```text
✓ SQLite local persistence

✓ WAL mode preferred

✓ Sensitive data encrypted before SQLite

✓ One vault_meta record

✓ 16-byte random vault IDs

✓ 16-byte random entry IDs

✓ 16-byte random history IDs

✓ 24-byte XChaCha20-Poly1305 nonces

✓ Opaque ciphertext BLOBs

✓ Plaintext routing metadata kept minimal

✓ Human-readable metadata encrypted

✓ Tags encrypted

✓ Aliases encrypted

✓ Favorites encrypted

✓ Search index rebuilt in memory after unlock

✓ History stored as independently encrypted records

✓ No persistent password fingerprints

✓ No plaintext SQLite FTS index

✓ Transactional updates

✓ Explicit schema versioning

✓ Explicit crypto-suite versioning

✓ Explicit serialization versioning

✓ Defensive size limits

✓ SQLite storage isolated from domain plaintext
```

---

# 195. Open Questions Before Implementation

The following still require concrete decisions:

1. Which canonical entry serialization format will be used?
2. What exact binary encoding will AAD use?
3. Should timestamps remain plaintext in final v1?
4. Should `record_kind` remain plaintext?
5. What exact size limits should be adopted?
6. Should history be enabled by default in v0.1?
7. Should SQLite WAL be mandatory or configurable?
8. Should GoVault use an application-level process lock?
9. How should corrupted individual records affect write access?
10. Should the vault format support multiple vault metadata rows in the future?

These should be resolved before declaring Vault Format v1 stable.

---

# 196. Next Specification

The next document should be:

```text
docs/backup-format.md
```

It should define the portable encrypted `.gvault` container.

It must specify:

```text
file magic
header
versioning
KDF metadata
wrapped Vault Key
backup ID
encrypted payload
manifest
canonical AAD
length encoding
size limits
integrity verification
restore semantics
atomic backup creation
```

The design sequence is now:

```text
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
implementation
```

