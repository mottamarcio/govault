# GoVault Backup Format Specification

**Document:** `docs/backup-format.md`
**Project:** GoVault
**Status:** Draft / Pre-v1 Portable Backup Specification
**Applies to:** GoVault v0.1+
**Last Updated:** September 2026

---

# 1. Purpose

This document defines the portable encrypted backup format used by GoVault.

The backup format is intentionally independent from SQLite.

A GoVault backup must represent the logical vault state rather than a raw copy of:

```text
vault.db
vault.db-wal
vault.db-shm
```

The primary backup file extension is:

```text
.gvault
```

This specification defines:

* File structure.
* File magic.
* Header layout.
* Backup identifiers.
* KDF metadata.
* Wrapped Vault Key storage.
* Backup payload encryption.
* Associated Data.
* Manifest structure.
* Entry serialization.
* History serialization.
* Integrity verification.
* Backup inspection.
* Restore behavior.
* Size limits.
* Corruption handling.
* Versioning.
* Compatibility.
* Atomic creation.
* Security expectations.

This document depends on:

```text
docs/threat-model.md
docs/cryptography.md
docs/vault-format.md
```

---

# 2. Design Goals

The GoVault backup format must provide:

1. Confidentiality.
2. Integrity.
3. Authentication.
4. Portability.
5. Independence from SQLite internals.
6. Self-contained restoration.
7. Format versioning.
8. Safe parsing.
9. Corruption detection.
10. Future migration support.
11. Human-inspectable non-secret metadata where useful.
12. No plaintext vault secrets.

---

# 3. Non-Goals

The backup format does not attempt to provide:

* Cloud synchronization.
* Incremental network backup.
* Deduplication across backups.
* Streaming synchronization.
* Team sharing.
* Multi-recipient encryption.
* Password recovery.
* Hidden volumes.
* Rollback prevention.
* Remote verification.
* Automatic upload.

---

# 4. Backup Philosophy

A `.gvault` file should be usable as:

```text
backup file
+
master password valid when that backup was created
```

without requiring the original SQLite database.

This is a hard requirement.

---

# 5. File Extension

The standard extension is:

```text
.gvault
```

Example:

```text
govault-2026-09-11.gvault
```

The extension is only a convention.

File identity must be determined from the file header.

---

# 6. File Magic

Every backup begins with fixed magic bytes.

Proposed ASCII representation:

```text
GOVAULTB
```

Exact byte sequence:

```text
47 4F 56 41 55 4C 54 42
```

This corresponds to:

```text
G O V A U L T B
```

---

# 7. Why File Magic Exists

The magic value allows GoVault to quickly distinguish:

```text
valid-looking GoVault backup
```

from:

```text
SQLite database
ZIP file
random data
other application files
```

It does not prove authenticity.

---

# 8. High-Level File Structure

Conceptually:

```text
+------------------------------+
| Fixed Header                 |
+------------------------------+
| KDF / Key-Wrap Metadata      |
+------------------------------+
| Encrypted Backup Payload     |
+------------------------------+
```

The encrypted payload contains:

```text
manifest
entries
history
logical vault metadata
```

---

# 9. Self-Contained Requirement

The unencrypted header must include enough information to:

1. Identify the format.
2. Validate supported versions.
3. Validate safe KDF bounds.
4. Derive the Key Encryption Key.
5. Authenticate and unwrap the Vault Key.
6. Derive the Backup Key.
7. Authenticate and decrypt the backup payload.

---

# 10. Backup Header Fields

Conceptually:

```text
magic
backup_format_version
crypto_suite_version
header_length
backup_id
vault_id
created_at
kdf_algorithm
kdf_salt
kdf_memory_kib
kdf_iterations
kdf_parallelism
key_wrap_nonce
wrapped_vault_key
backup_nonce
ciphertext_length
```

The exact encoding must be canonical.

---

# 11. Backup Format Version

Initial value:

```text
1
```

This version describes the `.gvault` container structure.

It is independent from:

```text
vault format version
schema version
entry serialization version
crypto suite version
```

---

# 12. Crypto Suite Version

For the initial format:

```text
1
```

meaning:

```text
Argon2id
HKDF-SHA-256
XChaCha20-Poly1305
```

Unsupported crypto suites must fail safely.

---

# 13. Backup Identifier

Each backup receives a random 16-byte identifier:

```text
backup_id = random(16 bytes)
```

Generated using:

```text
crypto/rand
```

The ID is not secret.

It is authenticated through Backup AAD.

---

# 14. Vault Identifier

The backup stores the original:

```text
vault_id
```

This preserves vault lineage.

The `vault_id` must be exactly:

```text
16 bytes
```

---

# 15. Created Timestamp

The backup header may contain:

```text
created_at_unix
```

as a signed 64-bit UTC Unix timestamp.

This metadata is not secret.

It allows basic structural inspection before decryption.

---

# 16. Timestamp Leakage

A stolen backup may reveal when it was created.

This is accepted in Backup Format v1.

If stronger metadata confidentiality becomes necessary, future formats may move the timestamp into the encrypted manifest.

---

# 17. KDF Metadata

The backup stores the KDF configuration that was active when the backup was created.

For Crypto Suite v1:

```text
kdf_algorithm
kdf_salt
kdf_memory_kib
kdf_iterations
kdf_parallelism
```

---

# 18. Why Backup KDF Metadata Is Preserved

Suppose:

```text
September 1
Master Password = A

September 2
Backup created

September 10
Master Password changed to B
```

The September 2 backup remains protected by:

```text
A
```

because it contains the wrapped Vault Key state from September 2.

This is required for independent restoration.

---

# 19. Wrapped Vault Key

The backup stores the wrapped Vault Key exactly as a protected recovery object.

Required fields:

```text
key_wrap_nonce
wrapped_vault_key
```

The plaintext Vault Key is never stored.

---

# 20. Backup Key Derivation

After the Vault Key is recovered:

```text
BackupKey =
    HKDF-SHA256(
        VaultKey,
        salt = vault_id,
        info = "govault:v1:backup"
    )
```

---

# 21. Backup Nonce

Each backup receives a fresh 24-byte random nonce:

```text
backup_nonce = random(24 bytes)
```

used with XChaCha20-Poly1305.

Nonce reuse is prohibited.

---

# 22. Backup Payload

The encrypted backup payload contains the logical state of the vault.

Conceptually:

```text
BackupPayloadV1 {
    manifest
    entries[]
    history[]
}
```

---

# 23. Backup Manifest

The encrypted manifest should contain:

```text
logical vault format version
source schema version
entry count
history count
entry serialization versions
created_by application version
logical creation timestamp
feature flags
```

Potential future fields:

```text
attachments count
custom object types
migration metadata
```

---

# 24. Manifest Confidentiality

The manifest is encrypted because it may reveal:

```text
number of credentials
history usage
record types
vault feature usage
```

Only the minimum pre-decryption header metadata remains plaintext.

---

# 25. Entries in Backup

Backup entries should represent logical records rather than SQLite rows.

Conceptually:

```text
BackupEntry {
    entry_id
    record_kind
    serialization_version
    payload
    created_at
    updated_at
}
```

The `payload` is the logical serialized entry content.

---

# 26. Double Encryption Decision

There are two possible backup models.

## Model A

Store already-encrypted SQLite record ciphertext inside the backup.

## Model B

Decrypt logical entries and re-encrypt the complete backup container.

Backup Format v1 chooses:

```text
Model B
```

---

# 27. Why Logical Backup Is Preferred

Logical backup avoids coupling backup compatibility to:

```text
SQLite schema
row layout
record-level nonce format
database migrations
```

It allows restoration into a newer database schema.

---

# 28. Security of Logical Backup Creation

During backup creation, GoVault may need to decrypt logical records in memory.

This is acceptable because the vault is already unlocked.

Plaintext must never be written to temporary files.

The serialized logical backup payload must be encrypted before it reaches persistent storage.

---

# 29. History in Backup

History is serialized logically.

Conceptually:

```text
BackupHistoryRecord {
    history_id
    entry_id
    serialization_version
    payload
    created_at
}
```

---

# 30. Backup Payload Encryption

The complete canonical payload is encrypted with:

```text
XChaCha20-Poly1305
```

using:

```text
BackupKey
backup_nonce
BackupAAD
```

---

# 31. Backup AAD

Conceptually:

```text
BackupAAD =
    protocol_id
    backup_format_version
    crypto_suite_version
    backup_id
    vault_id
    created_at_unix
```

The exact representation must be canonical.

---

# 32. Why KDF Metadata Is Not All in Backup AAD

KDF metadata is authenticated indirectly through Vault Key wrapping.

The wrapped Vault Key AAD must continue to bind:

```text
KDF algorithm
KDF parameters
KDF salt
vault ID
crypto suite
```

as defined in `cryptography.md`.

---

# 33. Header Canonical Encoding

Backup Format v1 should use an explicit binary representation.

Recommended properties:

```text
fixed-width integers where practical
big-endian integers
explicit lengths
no textual number encoding
no Go struct memory dumps
no gob
no map-order dependence
```

---

# 34. Proposed Header Layout

Conceptually:

```text
Offset   Size    Field

0        8       magic
8        2       backup_format_version
10       2       crypto_suite_version
12       4       header_length

16       16      backup_id
32       16      vault_id

48       8       created_at_unix

56       2       kdf_algorithm
58       2       reserved

60       4       kdf_memory_kib
64       4       kdf_iterations
68       4       kdf_parallelism

72       32      kdf_salt

104      24      key_wrap_nonce

128      2       wrapped_vault_key_length
130      N       wrapped_vault_key

...      24      backup_nonce
...      8       ciphertext_length
...      N       encrypted payload
```

This layout is conceptual and must be finalized before implementation.

---

# 35. Reserved Fields

Reserved bytes should be initialized to zero.

On read:

```text
unknown non-zero reserved bits
```

should either:

```text
cause rejection
```

or:

```text
be ignored only if explicitly declared forward-compatible
```

The policy must be deterministic.

---

# 36. Integer Endianness

All multi-byte integers in the backup binary format use:

```text
big-endian
```

---

# 37. Length Validation

All length fields must be validated before allocation.

Examples:

```text
header_length
wrapped_vault_key_length
ciphertext_length
entry_count
history_count
individual payload lengths
```

---

# 38. Maximum Backup Size

GoVault should enforce an implementation limit.

Suggested initial maximum:

```text
8 GiB
```

This is intentionally far above expected personal vault sizes.

The limit protects against pathological or malicious files.

---

# 39. Maximum Header Size

Suggested:

```text
64 KiB
```

Backup Format v1 headers should normally be much smaller.

---

# 40. Maximum Wrapped Vault Key Size

For Crypto Suite v1:

```text
48 bytes expected
```

Any larger value should be rejected for suite v1.

---

# 41. Maximum Entry Count

Suggested initial defensive maximum:

```text
1,000,000 entries
```

---

# 42. Maximum History Count

Suggested defensive maximum:

```text
10,000,000 records
```

This may be refined before v1.

---

# 43. Payload Serialization

The backup payload should use the same canonical serialization family selected for logical entry storage.

It must not use:

```text
gob
raw Go memory
unstable reflection-based encoding
```

---

# 44. Backup Payload Root

Conceptually:

```text
BackupPayloadV1 {
    payload_version
    manifest
    entries
    history
}
```

---

# 45. Payload Version

The encrypted payload has its own version.

Initial:

```text
1
```

This allows payload evolution without necessarily changing the outer container.

---

# 46. Manifest Structure

Conceptually:

```text
BackupManifestV1 {
    source_vault_format_version
    source_schema_version
    application_version
    entry_count
    history_count
    created_at_unix
}
```

---

# 47. Entry Ordering

Entries should be serialized in deterministic order.

Recommended:

```text
sort by raw entry_id bytes
```

This improves:

```text
test reproducibility
deterministic fixtures
format inspection
```

It does not make encryption deterministic because every backup uses a fresh random nonce.

---

# 48. History Ordering

Recommended:

```text
sort by entry_id
then created_at
then history_id
```

---

# 49. Deterministic Plaintext Payload

Given the same logical vault state and serialization version, the plaintext serialized backup payload should ideally be deterministic.

This simplifies test vectors.

The final ciphertext remains non-deterministic due to random nonce generation.

---

# 50. Compression

Backup Format v1 should initially avoid compression.

Reasons:

```text
simpler implementation
smaller attack surface
no decompression bombs
clear size limits
simpler fuzzing
```

Vault data is usually small enough that compression is unnecessary.

---

# 51. Future Compression

A future format may support compression inside the encrypted payload.

If introduced:

```text
compress plaintext
→ encrypt compressed bytes
```

Compression must never occur after encryption.

Strict decompression size limits would be required.

---

# 52. Backup Creation Flow

Conceptually:

```text
unlock vault
    │
    ▼
read logical entries
    │
    ▼
read logical history
    │
    ▼
construct manifest
    │
    ▼
canonical serialize payload
    │
    ▼
derive Backup Key
    │
    ▼
generate backup_id
    │
    ▼
generate backup_nonce
    │
    ▼
construct BackupAAD
    │
    ▼
encrypt payload
    │
    ▼
construct header
    │
    ▼
write temporary file
    │
    ▼
fsync
    │
    ▼
reopen + structurally verify
    │
    ▼
optionally cryptographically verify
    │
    ▼
atomic rename
```

---

# 53. Backup Destination Safety

GoVault should not overwrite an existing backup silently.

Possible behavior:

```text
destination already exists
→ require explicit overwrite
```

---

# 54. Temporary Backup File

Use a temporary file in the same destination directory where practical.

Example:

```text
.govault.tmp-<random>
```

The temporary filename must not contain secrets.

---

# 55. Temporary File Permissions

On Unix-like systems:

```text
0600
```

from creation time.

---

# 56. No Plaintext Temporary Backup

At no point should GoVault write:

```text
manifest plaintext
entry plaintext
history plaintext
```

to temporary disk storage.

---

# 57. Atomic Rename

After successful creation:

```text
temp file
→ atomic rename
→ final .gvault
```

where supported.

This avoids partially written files appearing as completed backups.

---

# 58. Directory Sync

Where practical and supported, GoVault should consider syncing the parent directory after rename for stronger durability.

This is an implementation detail that should be evaluated per platform.

---

# 59. Backup Verification Modes

GoVault should support two conceptual verification modes.

## Structural Verification

No password required.

## Cryptographic Verification

Password required.

---

# 60. Structural Verification

Command concept:

```bash
govault backup verify backup.gvault
```

may first perform:

```text
magic check
version check
header length validation
KDF bounds validation
wrapped-key length validation
ciphertext length validation
file truncation detection
trailing data policy
```

---

# 61. Structural Verification Is Not Authentication

A structurally valid backup may still:

```text
contain tampered ciphertext
contain incorrect wrapped key
require wrong password
```

Therefore CLI output must not say:

```text
Backup is authentic
```

after structural verification alone.

---

# 62. Cryptographic Verification

After requesting the master password:

```text
derive KEK
authenticate wrapped Vault Key
derive Backup Key
authenticate/decrypt backup payload
parse payload
validate manifest
validate record counts
validate logical entries
```

Only then can GoVault claim cryptographic validity.

---

# 63. Verification Output

Example:

```text
✓ GoVault backup header
✓ Supported format
✓ KDF parameters valid
✓ Vault Key authenticated
✓ Backup payload authenticated
✓ Manifest valid
✓ 196 entries validated
✓ 824 history records validated

Backup is valid.
```

---

# 64. Wrong Password

Wrong password behavior:

```text
Unable to unlock backup.
Check the password and backup integrity.
```

Avoid exposing a low-level Poly1305 error.

---

# 65. Backup Inspection

A structural inspection may safely display:

```text
backup format version
crypto suite
backup ID
vault ID prefix
created timestamp
KDF profile
file size
```

It must not claim entry counts unless the encrypted payload has been authenticated and decrypted.

---

# 66. Authenticated Inspection

After password authentication:

```text
Created       Sep 11, 2026 10:42 UTC
Entries       196
History       824
Vault Format  1
Backup Format 1
Crypto Suite  1
```

---

# 67. Restore Preconditions

Before restore:

1. Open backup.
2. Validate structure.
3. Validate versions.
4. Validate KDF bounds.
5. Request password.
6. Authenticate wrapped Vault Key.
7. Decrypt and authenticate payload.
8. Parse payload.
9. Validate logical contents.
10. Check destination compatibility.

Only then may restore begin.

---

# 68. Restore Must Not Stream Unauthenticated Records Into Vault

GoVault must not:

```text
decrypt one chunk
write it immediately into live vault
continue before full validation
```

for Backup Format v1.

Preferred behavior:

```text
authenticate complete payload
validate
then restore
```

---

# 69. Why Full Authentication Comes First

This prevents partially applying attacker-controlled or corrupted backup data before authenticity is established.

---

# 70. Recovery Snapshot

Before replacing an existing vault, GoVault creates a protected recovery snapshot.

The snapshot may use the same `.gvault` backup mechanism.

Example:

```text
govault-recovery-<timestamp>.gvault
```

---

# 71. Restore Destination

Restore should target a temporary new SQLite vault.

Conceptually:

```text
backup payload
→ temporary vault database
→ validate
→ atomic replacement
```

---

# 72. Restore Into Temporary Database

Restore creates:

```text
new vault_meta
new encrypted entry envelopes
new encrypted history envelopes
```

using the restored Vault Key and current supported storage representation.

---

# 73. Important Restore Design Choice

The backup stores logical plaintext-equivalent data inside an encrypted container.

On restore, entries are re-encrypted into the local vault database.

This means restored SQLite ciphertext:

```text
does not need to match original SQLite ciphertext
```

because new record nonces may be generated.

---

# 74. Vault ID on Restore

A normal restore preserves:

```text
vault_id
```

because it restores the same logical vault identity.

---

# 75. Vault Key on Restore

A normal restore preserves:

```text
Vault Key
```

from the backup.

This preserves cryptographic lineage.

---

# 76. Restore and New Record Nonces

Entries restored into SQLite should receive fresh record-level nonces.

Do not reuse record encryption nonces from an old database.

---

# 77. Restore and Entry IDs

Entry IDs should be preserved.

This keeps history and references coherent.

---

# 78. Restore and History IDs

History IDs should also be preserved.

---

# 79. Restore and Schema Version

The backup may originate from an older schema.

Because it contains logical data, restore should target:

```text
current supported SQLite schema
```

rather than recreate the original physical schema.

---

# 80. Restore and Serialization Versions

If old logical payload versions are supported:

```text
parse old payload
→ migrate domain object
→ serialize current record format
```

---

# 81. Unsupported Old Backup

If the application cannot safely interpret a backup:

```text
Unsupported GoVault backup format.
```

It must not guess.

---

# 82. Newer Backup Version

If a backup comes from a newer version:

```text
This backup was created by a newer GoVault format.
Upgrade GoVault before restoring it.
```

---

# 83. Restore Failure

If any restore step fails:

```text
current vault remains untouched
temporary restore is removed where practical
recovery snapshot remains available
```

---

# 84. Restore Verification

Before final replacement, GoVault should:

```text
open temporary SQLite vault
validate schema
validate vault metadata
unlock using restored key material internally
authenticate restored entries
authenticate restored history
```

---

# 85. Atomic Replacement

After successful validation:

```text
existing vault → protected old/recovery state
temporary vault → active vault
```

Exact platform-safe mechanics belong in architecture implementation documentation.

---

# 86. Backup Data Model

Conceptually:

```text
BackupPayloadV1
├── Manifest
├── Entries[]
└── History[]
```

---

# 87. Entry Representation

Conceptually:

```text
BackupEntryV1 {
    entry_id: [16]byte
    record_kind: uint16
    serialization_version: uint16
    created_at_unix: int64
    updated_at_unix: int64
    payload_length: uint32/uint64
    payload: bytes
}
```

---

# 88. History Representation

Conceptually:

```text
BackupHistoryV1 {
    history_id: [16]byte
    entry_id: [16]byte
    serialization_version: uint16
    created_at_unix: int64
    payload_length
    payload
}
```

---

# 89. Manifest Entry Count

The manifest should include declared counts.

Parser behavior:

```text
declared entry count
must equal parsed entry count
```

Same for history.

Mismatch means corruption or malformed input.

---

# 90. Duplicate IDs

Backups containing duplicate:

```text
entry_id
history_id
```

must be rejected.

---

# 91. Broken History References

A history record referencing a missing entry should be rejected unless future tombstone semantics explicitly permit it.

Backup Format v1 does not permit orphaned history.

---

# 92. Duplicate Aliases

Because aliases are inside entry payloads, restore validation should rebuild the alias index and detect ambiguity before replacing the active vault.

---

# 93. Logical Validation

Before restore, validate:

```text
record kinds
UTF-8
tag limits
custom field limits
alias limits
TOTP parameters
payload sizes
entry counts
history references
```

---

# 94. Payload Size Limits

Use the same logical limits defined in `vault-format.md`.

A backup must not be able to bypass vault-level safety limits.

---

# 95. Ciphertext Length

The outer encrypted payload length must be bounded before allocation.

For the initial implementation, prefer streaming file reads into bounded buffers or files rather than trusting a single attacker-supplied allocation size.

---

# 96. Large Backup Handling

Even though typical backups are small, GoVault should avoid requiring:

```text
backup_size × multiple copies
```

in memory.

Implementation may parse decrypted content through bounded readers after successful AEAD handling strategy is established.

---

# 97. AEAD and Large Payloads

XChaCha20-Poly1305's standard AEAD API operates on complete messages.

For v1, this is acceptable because the expected vault size is relatively small.

If attachments or very large vaults are introduced, GoVault may need a chunked authenticated container in Backup Format v2.

---

# 98. v1 Payload Size Recommendation

To keep the initial implementation simple, Backup Format v1 should impose a practical payload ceiling.

Suggested:

```text
256 MiB
```

for the decrypted payload.

This is ample for password-manager data without attachments.

---

# 99. Why 256 MiB

Without attachments, a personal vault with hundreds of thousands of credentials should still remain below this bound.

This significantly simplifies:

```text
AEAD handling
memory bounds
fuzzing
DoS protection
```

---

# 100. Attachments and Backup v2

If GoVault later supports attachments, a chunked backup container should likely be introduced.

Do not silently expand v1 into an unbounded multi-gigabyte AEAD blob.

---

# 101. Chunking Is Not Part of v1

Backup Format v1 contains one authenticated encrypted payload.

This simplifies correctness.

---

# 102. Trailing Bytes

Backup Format v1 should reject unexpected trailing bytes after the declared ciphertext.

This helps detect:

```text
concatenated files
format confusion
corruption
```

---

# 103. Empty Vault

A backup containing:

```text
0 entries
0 history records
```

is valid.

---

# 104. Empty History

A backup with entries and no history is valid.

---

# 105. Password Change After Backup

A backup remains tied to the password/KDF configuration captured when created.

Changing the active vault password does not alter existing backup files.

---

# 106. Compromised Old Password

If an old master password is believed compromised:

1. Change the current vault password.
2. Create new backups.
3. Remove old backups under the user's control.

GoVault cannot revoke copies already obtained by an attacker.

---

# 107. Optional Backup Password

Separate backup passwords are not part of Backup Format v1.

Future support would require a new key-wrapping mode and threat-model update.

---

# 108. Export Is Not Backup

A future plaintext JSON or CSV export is not a GoVault backup.

Only authenticated encrypted `.gvault` files are considered native backups.

---

# 109. File Permission Defaults

On Unix-like systems:

```text
0600
```

for newly created `.gvault` files.

---

# 110. User Copying the File

Once created, the user may:

```text
copy it
move it
store it on USB
place it in cloud storage
send it elsewhere
```

GoVault does not manage those destinations.

The backup's confidentiality must not rely on destination trust.

---

# 111. Cloud Storage Neutrality

GoVault does not integrate with cloud providers.

A user may independently place `.gvault` files into a synced folder.

That behavior occurs outside GoVault.

---

# 112. Filename Confidentiality

The filename itself is outside encrypted backup content.

Users should avoid sensitive filenames such as:

```text
company-production-admin-passwords.gvault
```

if filename metadata matters to them.

---

# 113. Backup Metadata Leakage Summary

An attacker obtaining a `.gvault` file may learn:

```text
file size
backup format version
crypto suite
creation timestamp
KDF profile
vault identifier
backup identifier
```

They should not learn:

```text
entry names
usernames
passwords
URLs
tags
TOTP secrets
notes
history contents
entry counts
```

before successful decryption.

---

# 114. KDF DoS Protection

Before Argon2id execution, validate:

```text
memory
iterations
parallelism
salt length
output assumptions
```

against the bounds defined in `cryptography.md`.

---

# 115. Header Authentication Caveat

Not all header fields can be protected by the backup payload because some are required to derive keys.

Fields affecting key derivation must be protected through the wrapped Vault Key authentication design.

---

# 116. Backup ID Tampering

Because `backup_id` is part of Backup AAD:

```text
modify backup_id
→ payload authentication failure
```

---

# 117. Vault ID Tampering

`vault_id` participates in:

```text
Vault Key wrapping context
HKDF backup key derivation
Backup AAD
```

Tampering should cause authentication failure.

---

# 118. Timestamp Tampering

If `created_at_unix` is part of Backup AAD:

```text
modify timestamp
→ payload authentication failure
```

This is preferred.

---

# 119. Crypto Version Tampering

Changing:

```text
crypto_suite_version
```

must not cause fallback to another suite.

Unsupported or inconsistent values are rejected.

---

# 120. Backup Format Tampering

Changing:

```text
backup_format_version
```

must not cause heuristic parsing.

---

# 121. Magic Tampering

Invalid magic:

```text
reject immediately
```

---

# 122. Truncated Header

Reject before any KDF execution if mandatory header bytes are missing.

---

# 123. Truncated Ciphertext

Reject before AEAD open if declared size is inconsistent with actual file size.

---

# 124. Modified Ciphertext

AEAD authentication must fail.

No partial plaintext is returned.

---

# 125. Modified Authentication Tag

Authentication fails.

---

# 126. Modified Wrapped Vault Key

Vault Key unwrap fails.

---

# 127. Modified Key-Wrap Nonce

Vault Key unwrap fails.

---

# 128. Modified KDF Salt

Derived KEK changes.

Vault Key unwrap fails.

---

# 129. Modified KDF Parameters

Either:

```text
parameters rejected by safety bounds
```

or:

```text
different KEK produced
→ unwrap fails
```

---

# 130. Recovery From Corrupted Backup

GoVault must never attempt unauthenticated "best effort" recovery of secret content from a corrupted backup.

Recovery tooling may inspect structural portions, but plaintext extraction requires successful authentication.

---

# 131. Backup Listing

A future local backup directory browser may list files by structural metadata.

Example:

```text
Sep 11  384 KiB  Format 1
Sep 04  372 KiB  Format 1
Aug 28  365 KiB  Format 1
```

Entry counts require authentication.

---

# 132. Backup Verification Without Restore

Verification must never modify the active vault.

---

# 133. Verify Command

Recommended CLI:

```bash
govault backup verify ~/backups/personal.gvault
```

Potential interactive flow:

```text
Backup structure valid.

Verify cryptographically?
Master password: ********
```

---

# 134. Backup Inspect Command

Recommended:

```bash
govault backup inspect backup.gvault
```

Structural fields may be displayed before password entry.

Detailed manifest requires authentication.

---

# 135. Backup Create Command

Recommended:

```bash
govault backup create ~/backups/personal.gvault
```

---

# 136. Restore Command

Recommended:

```bash
govault restore ~/backups/personal.gvault
```

---

# 137. TUI Backup Integration

The TUI backup screen should display:

```text
Destination
Format
Encryption status
Entry count
History count
Estimated file size
```

Entry count is available because the current vault is already unlocked.

---

# 138. TUI Verification

After creation:

```text
✓ Backup written
✓ Structure verified
✓ Cryptographic verification passed
```

GoVault should preferably verify its own newly written backup before reporting success.

---

# 139. Self-Verification

Recommended creation flow:

```text
write
→ reopen
→ parse
→ authenticate
→ compare expected manifest
→ success
```

This catches write/path bugs.

---

# 140. Backup Success Semantics

GoVault should only print:

```text
Backup created successfully.
```

after the final file exists and passes configured verification.

---

# 141. Failure Cleanup

If creation fails:

```text
temporary file should be removed where practical
existing destination must remain untouched
active vault must remain untouched
```

---

# 142. Backup Overwrite

If overwrite is explicitly requested:

```text
new temp backup
→ verify
→ replace destination atomically
```

Do not truncate the existing valid backup first.

---

# 143. Backup Rotation

Automatic rotation is not required for v1.

Future local-only retention could support:

```text
keep last N
keep daily/weekly/monthly
```

but this is application policy, not format behavior.

---

# 144. Backup Filename Generation

Suggested default:

```text
govault-YYYY-MM-DD-HHMMSS.gvault
```

Filename generation should not include vault entry data.

---

# 145. Backup Comparison

Two backups of the same logical vault should normally have different ciphertext because of:

```text
different backup IDs
different backup nonces
possibly different timestamps
```

This is expected.

---

# 146. Determinism Is Not Required for Ciphertext

Do not attempt deterministic encryption for backup deduplication.

---

# 147. Logical Equality

A future diagnostic tool may compare authenticated logical manifests after decrypting backups.

Not required for v1.

---

# 148. Backup Cloning

Copying a `.gvault` file byte-for-byte is safe as a backup duplication operation.

Both copies represent the same backup identity.

---

# 149. Backup Re-Encryption

A future command might support:

```bash
govault backup rekey old.gvault new.gvault
```

to protect the same logical backup with new wrapping credentials.

Not required for MVP.

---

# 150. Separate Backup Passphrase — Future

If introduced later:

```text
Backup KEK
=
Argon2id(separate backup password)
```

may wrap the same Vault Key or a backup-specific key.

This requires a new format or explicit wrapping mode.

---

# 151. Import From Backup

Native `.gvault` restore is not the same as generic import.

Restore preserves:

```text
entry IDs
history IDs
vault identity
```

Generic import into another vault should instead:

```text
generate new IDs
re-encrypt under destination vault
```

---

# 152. Merge Restore

Backup Format v1 defines full restore.

Merge restore is not required.

Future command:

```bash
govault restore --merge backup.gvault
```

would need conflict semantics for:

```text
entry IDs
aliases
history
duplicates
```

---

# 153. Conflict Handling

Full restore has no logical merge conflicts because it replaces the current vault.

---

# 154. Backup Manifest Feature Flags

The manifest may include feature flags.

Example:

```text
HAS_HISTORY
HAS_TOTP
HAS_CUSTOM_FIELDS
```

Flags must be informational and authenticated.

They must not allow the parser to skip required validation.

---

# 155. Unknown Feature Flags

Unknown required features:

```text
reject
```

Unknown optional features:

```text
may ignore
```

if the format explicitly distinguishes required vs optional flags.

---

# 156. Application Version

The manifest may store:

```text
created_by = "govault 0.4.0"
```

This is diagnostic metadata only.

Restore behavior must depend on format versions, not application version strings.

---

# 157. No Semantic Dependence on Version Strings

Never parse:

```text
"GoVault 1.2.3"
```

to decide cryptographic behavior.

Use explicit numeric protocol fields.

---

# 158. Backup Parser Architecture

Recommended package:

```text
internal/backup/
├── format.go
├── header.go
├── payload.go
├── reader.go
├── writer.go
├── verify.go
├── restore.go
├── limits.go
└── errors.go
```

---

# 159. Crypto Boundary

Backup code should not invoke low-level primitives arbitrarily.

Preferred flow:

```text
backup package
→ internal/crypto backup APIs
```

---

# 160. Proposed Crypto APIs

Conceptually:

```go
DeriveBackupKey(vaultKey, vaultID)

EncryptBackupPayload(
    key BackupKey,
    nonce BackupNonce,
    aad BackupAAD,
    plaintext []byte,
)

DecryptBackupPayload(...)
```

---

# 161. Header Parser Boundary

Header parsing occurs before password derivation.

It must therefore be designed as hostile-input parsing code.

---

# 162. Parser Rules

The parser must:

```text
never trust lengths
never trust counts
never trust versions
never trust reserved bits
never allocate unbounded memory
never panic on malformed input
```

---

# 163. Error Types

Recommended typed errors:

```text
ErrInvalidBackupMagic
ErrUnsupportedBackupVersion
ErrUnsupportedCryptoSuite
ErrInvalidHeader
ErrInvalidLength
ErrKDFBounds
ErrAuthenticationFailed
ErrCorruptPayload
ErrDuplicateEntry
ErrOrphanHistory
ErrSemanticValidation
```

---

# 164. User-Facing Error Separation

Low-level parser errors should be translated into clear messages.

Example:

```text
The backup file is truncated or corrupted.
```

rather than:

```text
unexpected EOF at offset 130
```

unless verbose diagnostics are explicitly requested.

---

# 165. Fuzzing

Backup parsing is a primary fuzzing target.

Targets:

```text
header parser
length parser
payload parser
manifest parser
entry parser
history parser
```

---

# 166. Fuzz Invariants

Fuzzing should verify:

```text
no panic
no unbounded allocation
no hangs
no out-of-range indexing
no unauthenticated plaintext exposure
```

---

# 167. Test Vectors

Backup Format v1 should ship deterministic plaintext fixtures and cryptographic vectors.

Test vectors should include:

```text
master password
KDF salt
KDF parameters
Vault Key
wrapped Vault Key
backup ID
vault ID
backup nonce
AAD
plaintext payload
ciphertext
```

---

# 168. Golden Files

Repository fixtures may include:

```text
valid-v1.gvault
wrong-tag-v1.gvault
truncated-v1.gvault
future-version.gvault
malformed-header.gvault
```

All must use non-production test keys.

---

# 169. Cross-Version Restore Tests

Once Backup Format v2 exists, tests should verify:

```text
new GoVault restores v1
new GoVault restores v2
old GoVault rejects v2 safely
```

---

# 170. Tampering Tests

Flip one bit in:

```text
backup ID
vault ID
timestamp
KDF salt
KDF parameters
key-wrap nonce
wrapped key
backup nonce
ciphertext
authentication tag
```

Expected behavior must be documented.

---

# 171. Known Plaintext Test

Backup tests should contain:

```text
GOVAULT_BACKUP_SECRET_TEST_XYZ
```

Then scan raw `.gvault` bytes.

Expected:

```text
not found
```

---

# 172. Metadata Leakage Test

Human-readable strings such as:

```text
GitHub Personal
john@example.com
production
```

must not appear in raw backup bytes.

---

# 173. Backup Password Test Cases

Test:

```text
correct password
wrong password
empty password
Unicode password
very long password
password from before current vault password rotation
```

---

# 174. Rotation Compatibility Test

Scenario:

```text
create vault using Password A
create Backup 1
change master password to B
create Backup 2
```

Expected:

```text
Backup 1 unlocks with A
Backup 1 does not unlock with B

Backup 2 unlocks with B
Backup 2 does not unlock with A
```

unless passwords happen to be equal.

---

# 175. Restore Test

Test full cycle:

```text
create vault
add entries
add history
backup
delete original vault
restore backup
compare logical contents
```

---

# 176. Restore Does Not Require Original DB

This must be an explicit automated test.

---

# 177. Restore Into Newer Schema

When schema v2 exists:

```text
Backup Format 1
created from Schema 1
→ restore into Schema 2
```

should work if logical payload compatibility exists.

---

# 178. Corruption Tests

Test:

```text
truncated file
extra trailing bytes
wrong magic
unsupported version
huge header length
huge ciphertext length
invalid wrapped key length
duplicate entry IDs
orphan history
malformed UTF-8
oversized fields
```

---

# 179. Power Failure Testing

Simulate failure:

```text
during temporary write
during fsync
before rename
during overwrite replacement
```

The previous valid backup should remain intact.

---

# 180. Restore Power Failure Testing

Simulate failure:

```text
while constructing temporary vault
while verifying temporary vault
before active vault replacement
```

The original active vault must remain usable.

---

# 181. Performance Targets

For typical personal vaults:

```text
backup creation should feel immediate after KDF/unlock costs
verification should complete quickly
restore should scale linearly with record count
```

Correctness and safety take priority over maximum throughput.

---

# 182. Streaming Future

If GoVault later stores:

```text
attachments
large binary secrets
large secure documents
```

Backup Format v2 should consider:

```text
chunked AEAD
per-chunk authentication
authenticated index
bounded streaming
```

Do not retrofit unsafe streaming into v1.

---

# 183. No Random Access Requirement

Backup Format v1 does not need random-access entry retrieval.

The primary operations are:

```text
create
inspect
verify
restore
```

---

# 184. No In-Place Mutation

A `.gvault` backup is immutable.

GoVault must never edit a backup file in place.

Changes require writing a new file.

---

# 185. Why Immutable Backups

Immutability simplifies:

```text
integrity guarantees
atomic writes
testing
format reasoning
recovery
```

---

# 186. Backup Identity

Two files with the same `backup_id` should represent copies of the same logical backup file.

A newly created backup must receive a new ID even if the logical vault state is identical.

---

# 187. Backup Lineage

Because the backup includes:

```text
vault_id
```

GoVault can determine whether a backup belongs to the same logical vault lineage as the current vault.

---

# 188. Restore Warning for Different Vault

If restoring over an existing vault with a different `vault_id`, the UI should clearly warn:

```text
This backup belongs to a different GoVault vault.
Restoring will replace the current vault.
```

---

# 189. Same-Vault Restore

If `vault_id` matches, the UI may display:

```text
Backup belongs to the current vault.
```

This is informational only.

---

# 190. Backup Age Warning

If a backup is significantly older than current state, GoVault may warn:

```text
This backup is older than the active vault.
```

This is not cryptographic rollback protection.

---

# 191. Rollback Limitation

A valid old backup is indistinguishable cryptographically from a newer valid backup based solely on authenticity.

GoVault cannot fully prevent rollback without trusted external state.

---

# 192. Backup Format Security Invariants

The following are mandatory:

## BAK-001

No plaintext vault secret appears in a `.gvault` file.

## BAK-002

The Vault Key is never stored plaintext.

## BAK-003

Backup payloads use authenticated encryption.

## BAK-004

Every backup uses a fresh random nonce.

## BAK-005

Every backup uses a fresh random backup ID.

## BAK-006

KDF parameters are validated before use.

## BAK-007

Backup payload authentication occurs before restore.

## BAK-008

A failed restore does not destroy the active vault.

## BAK-009

Backups are self-contained.

## BAK-010

Backups are independent of SQLite physical layout.

## BAK-011

The backup format is explicitly versioned.

## BAK-012

GoVault never silently interprets unknown backup versions.

---

# 193. Security Review Checklist

Changes to backup handling should answer:

```text
Does this expose new plaintext metadata?

Does this alter Backup AAD?

Does this alter key derivation?

Does this alter Vault Key wrapping?

Does this introduce nonce reuse risk?

Does this add unbounded allocation?

Does this require a new backup version?

Does this preserve old backup compatibility?

Can malformed input trigger a panic?

Can restore partially modify the active vault?

Can temporary plaintext reach disk?
```

---

# 194. Decisions Established by Backup Format v1

Backup Format v1 establishes:

```text
✓ `.gvault` native backup extension

✓ Fixed file magic

✓ Explicit backup format version

✓ Explicit crypto suite version

✓ Random 128-bit backup ID

✓ Original vault ID preserved

✓ KDF metadata embedded

✓ Wrapped Vault Key embedded

✓ Self-contained restoration

✓ Backup Key derived from Vault Key

✓ XChaCha20-Poly1305 payload encryption

✓ Fresh random 24-byte backup nonce

✓ Authenticated outer metadata

✓ Logical backup rather than raw SQLite copy

✓ Encrypted manifest

✓ Encrypted entries

✓ Encrypted history

✓ No compression in v1

✓ One authenticated payload in v1

✓ Practical payload size ceiling

✓ Atomic temporary-file creation

✓ Verification before success

✓ Restore into temporary database

✓ Active vault preserved on restore failure

✓ Existing backups retain old master-password semantics
```

---

# 195. Open Questions Before Implementation

The following should be finalized before declaring Backup Format v1 stable:

1. Exact binary header layout.
2. Exact magic bytes.
3. Exact canonical payload serialization format.
4. Exact integer sizes for entry counts and payload lengths.
5. Exact maximum backup payload size.
6. Whether `created_at_unix` should remain plaintext.
7. Whether application version belongs inside the encrypted manifest.
8. Whether self-verification is mandatory or configurable.
9. Whether recovery snapshots should always use `.gvault`.
10. Whether same-vault versus different-vault restore warnings belong in CLI and TUI.
11. Whether large-vault support requires chunking before v1.0.
12. Whether separate backup passwords should be postponed to Backup Format v2.

---

# 196. Recommended Next Document

The next document should be:

```text
docs/architecture.md
```

It should turn the specifications into concrete Go architecture.

It should define:

```text
packages
dependencies
interfaces
domain models
application services
storage abstractions
crypto boundaries
backup boundaries
TUI architecture
CLI architecture
error model
configuration
process lifecycle
unlock lifecycle
dependency injection
testing structure
```

---

# 197. Specification Sequence

The project now has the following design path:

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
code
```

---

# 198. Final Backup Model

The complete backup model can be summarized as:

```text
Master Password
      │
      ▼
   Argon2id
      │
      ▼
      KEK
      │
      ▼
Unwrap Vault Key
      │
      ▼
   Vault Key
      │
      ▼
HKDF-SHA-256
"govault:v1:backup"
      │
      ▼
  Backup Key
      │
      ▼
XChaCha20-Poly1305
      │
      ▼
Encrypted logical vault payload
      │
      ▼
.gvault
```

The core recovery guarantee is:

> A valid GoVault backup contains everything required to reconstruct the logical vault, except the user's master password.

