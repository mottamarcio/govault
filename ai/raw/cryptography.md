# GoVault Cryptographic Design

**Document:** `docs/cryptography.md`
**Project:** GoVault
**Status:** Draft / Pre-v1 Cryptographic Specification
**Applies to:** GoVault v0.1+
**Last Updated:** September 2026

---

# 1. Purpose

This document specifies the cryptographic architecture of GoVault.

It defines:

* Cryptographic primitives.
* Key hierarchy.
* Master password derivation.
* Vault key generation.
* Key wrapping.
* Domain separation.
* Record encryption.
* Metadata authentication.
* Nonce generation.
* Backup encryption.
* Master password rotation.
* Cryptographic versioning.
* Migration requirements.
* Validation rules.
* Test requirements.

This document is security-sensitive.

Changes to the cryptographic format must receive explicit review and must not be introduced as incidental implementation changes.

The threat model is defined separately in:

```text
docs/threat-model.md
```

This specification assumes the guarantees and limitations defined there.

---

# 2. Design Goals

The GoVault cryptographic design must provide:

1. Confidentiality of vault secrets.
2. Integrity of encrypted records.
3. Authentication of cryptographic context.
4. Strong resistance to offline password guessing.
5. Isolation between different vaults.
6. Isolation between different records.
7. Isolation between cryptographic purposes.
8. Safe master password rotation.
9. Cryptographically protected backups.
10. Future algorithm and format migration.
11. Safe failure when cryptographic validation fails.

---

# 3. Non-Goals

The cryptographic layer does not attempt to provide:

* Protection against a fully compromised host.
* Protection against memory inspection while unlocked.
* Perfect memory zeroization.
* Protection against keyloggers.
* Protection against malicious terminal emulators.
* Hardware-backed key protection.
* Cloud key management.
* Multi-user key sharing.
* Threshold cryptography.
* Password recovery.
* Plausible deniability.
* Full rollback protection.

---

# 4. Cryptographic Suite v1

GoVault Crypto Suite v1 uses:

```text
Password KDF
    Argon2id

Key derivation / domain separation
    HKDF-SHA-256

Authenticated encryption
    XChaCha20-Poly1305

Cryptographically secure randomness
    crypto/rand

Identifiers
    128-bit random identifiers

Root vault key
    256-bit random key
```

GoVault must not implement custom cryptographic primitives.

---

# 5. Why XChaCha20-Poly1305

GoVault v1 uses XChaCha20-Poly1305 as the primary authenticated encryption algorithm.

The Go implementation should use:

```text
golang.org/x/crypto/chacha20poly1305
```

Specifically:

```go
chacha20poly1305.NewX(...)
```

XChaCha20-Poly1305 provides:

* Authenticated encryption.
* A 256-bit key.
* A 192-bit nonce.
* Strong software performance.
* No dependency on AES hardware acceleration.
* A large nonce space suitable for random nonce generation.

The large nonce makes random nonce generation operationally simpler than schemes with smaller nonce spaces.

---

# 6. Why Not AES-GCM for v1

AES-GCM is a valid modern AEAD construction.

However, GoVault v1 prefers XChaCha20-Poly1305 because:

* Random nonce generation is easier to manage safely.
* The 192-bit nonce provides an extremely large collision space.
* Performance is strong across hardware.
* It avoids depending on AES acceleration characteristics.
* It simplifies the record encryption design.

This does not imply that AES-GCM is insecure.

Future crypto suites may support alternative AEAD constructions if required.

---

# 7. Why Argon2id

The master password is vulnerable to offline guessing if an attacker obtains the vault.

GoVault therefore uses:

```text
Argon2id
```

Argon2id provides memory-hard password derivation intended to increase the cost of password guessing.

Implementation:

```text
golang.org/x/crypto/argon2
```

GoVault must use:

```go
argon2.IDKey(...)
```

and must not substitute:

```text
SHA-256(password)
PBKDF without explicit migration
custom iterative hashing
```

for master password derivation.

---

# 8. Why HKDF

GoVault uses HKDF-SHA-256 to derive independent cryptographic subkeys from the root Vault Key.

This allows one randomly generated root key to produce purpose-specific keys.

Conceptually:

```text
Vault Key
   │
   ├── Entry Encryption Key
   ├── History Encryption Key
   ├── Backup Encryption Key
   └── Future Domain Keys
```

Compartmentalizing cryptographic purposes reduces accidental key reuse.

---

# 9. High-Level Key Hierarchy

The core hierarchy is:

```text
Master Password
       │
       │ Argon2id
       ▼
Key Encryption Key (KEK)
       │
       │ XChaCha20-Poly1305
       ▼
Encrypted Vault Key
       │
       ▼
Vault Key (VK)
       │
       │ HKDF-SHA-256
       │
       ├───────────────┐
       ▼               ▼
Entry Key         History Key
       │
       ├───────────────┐
       │               │
       ▼               ▼
Backup Key        Future Keys
```

The master password does not directly encrypt vault records.

---

# 10. Master Password

The master password is user-provided secret input.

It must:

* Never be persisted.
* Never be logged.
* Never be accepted through ordinary command-line arguments.
* Never be stored in configuration.
* Never be placed in environment variables by GoVault.
* Be discarded as soon as practical after key derivation.

GoVault should prefer mutable byte buffers over immutable strings for password handling where practical.

---

# 11. Master Password Encoding

Before Argon2id processing, the master password is encoded as UTF-8.

GoVault v1 does not automatically perform Unicode normalization.

The exact bytes entered by the user are cryptographically significant.

This avoids silently changing password semantics between versions.

The UI should therefore preserve user input exactly.

---

# 12. KDF Salt

Every vault receives a cryptographically random salt.

Required size:

```text
32 bytes / 256 bits
```

Generation:

```go
crypto/rand.Read(...)
```

The salt is not secret.

It is stored in the vault cryptographic header.

---

# 13. Argon2id Parameters

GoVault v1 defines an initial recommended profile.

```text
Algorithm      Argon2id
Memory         64 MiB
Iterations     3
Parallelism    4
Output         32 bytes
Salt           32 bytes
```

Conceptually:

```go
argon2.IDKey(
    password,
    salt,
    3,
    64*1024,
    4,
    32,
)
```

These values are the initial interoperability profile, not an eternal security constant.

---

# 14. KDF Calibration

Future versions may provide local calibration.

For v0.x, deterministic parameters are preferred because they simplify:

* Testing.
* Compatibility.
* Documentation.
* Reproducibility.

Future versions may target an approximate unlock duration while enforcing minimum security parameters.

Any calibration feature must remain offline.

---

# 15. KDF Parameter Validation

KDF parameters are attacker-controlled when an attacker can modify the vault header.

GoVault must validate them before allocating memory.

Example acceptable implementation bounds:

```text
Memory

minimum     32 MiB
default     64 MiB
maximum     1 GiB

Iterations

minimum     1
default     3
maximum     10

Parallelism

minimum     1
default     4
maximum     16

Output

exactly     32 bytes for Crypto Suite v1
```

These bounds primarily protect the application from denial-of-service through malicious vault files.

The accepted minimum for existing legacy vaults may eventually differ from the minimum used when creating new vaults.

---

# 16. Key Encryption Key

Argon2id produces a 256-bit Key Encryption Key:

```text
KEK = Argon2id(
    master_password,
    kdf_salt,
    parameters
)
```

The KEK exists only temporarily.

Its purpose is to decrypt or encrypt the Vault Key.

The KEK must not encrypt normal vault records.

---

# 17. Vault Key

Each vault receives a random 256-bit root key:

```text
Vault Key (VK)
Size: 32 bytes
```

Generation:

```go
crypto/rand
```

The Vault Key is independent of the master password.

This distinction is fundamental to GoVault's architecture.

---

# 18. Why the Vault Key Is Random

The Vault Key is generated randomly instead of being derived directly from the master password.

This allows:

* Master password changes without re-encrypting every record.
* KDF upgrades without re-encrypting every record.
* Cleaner separation between authentication and storage encryption.
* Future key hierarchy extensions.

---

# 19. Vault Key Persistence

The Vault Key must never be persisted in plaintext.

Instead:

```text
Vault Key
   │
   ▼
XChaCha20-Poly1305
using KEK
   │
   ▼
Wrapped Vault Key
```

Only the wrapped Vault Key is stored.

---

# 20. Vault Identifier

Each vault receives a random 128-bit identifier:

```text
vault_id = random(16 bytes)
```

The vault identifier is not secret.

It is used for:

* Domain separation.
* Associated data.
* Cross-vault substitution protection.

The identifier must never be reused intentionally when creating a new independent vault.

---

# 21. Key Wrapping

The Vault Key is wrapped using:

```text
XChaCha20-Poly1305
```

with the KEK.

Conceptually:

```text
wrapped_vk =
    AEAD_Encrypt(
        key        = KEK,
        nonce      = random 24-byte nonce,
        plaintext  = VaultKey,
        aad        = VaultKeyAAD
    )
```

---

# 22. Vault Key Associated Data

The wrapped Vault Key should authenticate contextual information.

Conceptually:

```text
VaultKeyAAD =
    protocol_identifier
    || crypto_version
    || vault_id
    || kdf_algorithm
    || kdf_parameters
```

Serialization must be canonical.

Do not construct cryptographic associated data using ambiguous string concatenation.

---

# 23. Canonical Encoding

Cryptographic structures must have deterministic encoding.

GoVault must not rely on:

* Go map iteration order.
* Locale-specific formatting.
* Human-readable string concatenation.
* Whitespace-sensitive ad hoc encoding.

Potential formats include:

* Explicit binary structures.
* Deterministic CBOR.
* Carefully specified length-prefixed binary encoding.

The final format will be defined in:

```text
docs/vault-format.md
```

---

# 24. Unlock Flow

Unlock conceptually performs:

```text
read vault header
        │
        ▼
validate format
        │
        ▼
validate KDF parameters
        │
        ▼
read master password
        │
        ▼
Argon2id
        │
        ▼
KEK
        │
        ▼
authenticate + decrypt wrapped Vault Key
        │
        ├── failure → unlock rejected
        │
        ▼
Vault Key
        │
        ▼
derive domain keys
        │
        ▼
UNLOCKED
```

---

# 25. Incorrect Master Password

An incorrect master password produces an incorrect KEK.

The wrapped Vault Key will fail AEAD authentication.

The application should return a generic unlock failure.

Example:

```text
Unable to unlock vault.
Check the master password and vault integrity.
```

GoVault should not expose low-level authentication errors in normal UI.

---

# 26. Domain Separation

The root Vault Key should not directly encrypt every data category.

Subkeys are derived using HKDF-SHA-256.

Conceptually:

```text
EntryKey =
    HKDF(
        VaultKey,
        info = "govault:v1:entry"
    )

HistoryKey =
    HKDF(
        VaultKey,
        info = "govault:v1:history"
    )

BackupKey =
    HKDF(
        VaultKey,
        info = "govault:v1:backup"
    )
```

The exact labels are protocol constants.

Changing a label changes the resulting key.

---

# 27. HKDF Inputs

For Crypto Suite v1:

```text
Hash:

SHA-256

Input Key Material:

Vault Key

Salt:

vault_id

Info:

explicit ASCII protocol label
```

Conceptually:

```text
HKDF-SHA256(
    ikm  = VaultKey,
    salt = vault_id,
    info = "govault:v1:entry"
)
```

Output:

```text
32 bytes
```

---

# 28. Domain Labels

Initial labels:

```text
govault:v1:entry
govault:v1:history
govault:v1:backup
```

Future cryptographic domains must receive distinct labels.

Labels must never be repurposed.

---

# 29. Entry Encryption

Each vault entry is independently encrypted.

Conceptually:

```text
Entry Plaintext
      │
      ▼
Canonical Serialization
      │
      ▼
XChaCha20-Poly1305
      │
      ▼
Encrypted Entry
```

Each encryption uses a fresh random nonce.

---

# 30. Entry Nonce

XChaCha20-Poly1305 uses a 24-byte nonce.

For every encryption operation:

```text
nonce = random(24 bytes)
```

using:

```go
crypto/rand
```

GoVault does not use counters for entry nonces in Crypto Suite v1.

---

# 31. Nonce Reuse

Nonce reuse with the same key is prohibited.

The 192-bit random nonce space makes accidental collision extraordinarily unlikely when using a secure random source.

GoVault must:

* Generate a new nonce for every encryption.
* Never derive nonces from timestamps.
* Never use record IDs directly as nonces.
* Never use counters stored in unreliable mutable state.
* Never reuse a nonce when updating an entry.

---

# 32. Random Failure

If `crypto/rand` fails:

```text
the operation fails
```

GoVault must never fall back to:

```text
math/rand
timestamps
process IDs
pseudo-random fallback
```

---

# 33. Record Identifier

Each entry receives a random 128-bit identifier:

```text
entry_id = random(16 bytes)
```

The identifier is immutable for the lifetime of the entry.

It may be represented to users as an encoded UUID-like value, but cryptographic processing should use canonical raw bytes.

---

# 34. Entry Associated Data

Entry ciphertext must be bound to its context.

Conceptually:

```text
EntryAAD =
    protocol_identifier
    || crypto_version
    || vault_id
    || entry_id
    || record_type
    || serialization_version
```

This prevents a valid ciphertext from being freely substituted into another cryptographic context.

---

# 35. Cross-Vault Substitution

Because `vault_id` is authenticated in the record context, ciphertext from Vault A must not authenticate as a valid record in Vault B.

---

# 36. Cross-Record Substitution

Because `entry_id` is authenticated, ciphertext from Entry A must not authenticate when substituted for Entry B.

---

# 37. Record Type Authentication

The entry type should be included in authenticated context.

Example:

```text
login
secure_note
api_credential
database_credential
custom
```

An attacker must not be able to change an encrypted record from one semantic type to another without authentication failure.

---

# 38. Entry Payload

The encrypted payload may conceptually contain:

```text
name
username
password
website
tags
notes
TOTP secret
custom fields
favorite
internal metadata
```

The exact payload depends on entry type.

Human-readable identifying metadata should preferably remain inside the encrypted payload.

---

# 39. Plaintext Database Metadata

The database should contain only metadata required for storage mechanics or pre-decryption routing.

Potential plaintext fields:

```text
entry_id
record_type identifier where unavoidable
crypto version
nonce
ciphertext
created storage timestamp where accepted
updated storage timestamp where accepted
```

Even record type may be encrypted if the storage architecture permits it.

The final policy belongs in:

```text
docs/vault-format.md
```

---

# 40. Sensitive Metadata

The following should be encrypted by default:

```text
entry name
username
website
tags
aliases
notes
favorite label if considered sensitive
TOTP issuer/account
custom field names when revealing
```

This reduces useful information available from a stolen database.

---

# 41. Search Architecture Implication

Because identifying metadata is encrypted, GoVault cannot rely on SQLite plaintext full-text search.

After unlock, GoVault may:

1. Decrypt searchable metadata.
2. Build an in-memory search index.
3. Keep the index only while unlocked.
4. Destroy the index on lock.

Passwords and other high-value secret fields do not need to be part of the search index.

---

# 42. Search Cache

An in-memory search record might contain:

```text
entry_id
name
username
tags
type
alias
```

It must never be persisted as plaintext.

The search cache is destroyed on lock.

---

# 43. Entry Update

Updating an entry requires:

```text
load/decrypt current entry
        │
        ▼
create history version if enabled
        │
        ▼
modify plaintext
        │
        ▼
serialize
        │
        ▼
generate NEW nonce
        │
        ▼
encrypt
        │
        ▼
atomic database transaction
```

The old nonce must not be reused.

---

# 44. History Encryption

Entry history uses a separate domain key:

```text
HistoryKey
```

derived from the Vault Key.

Each history version receives:

```text
history_id
entry_id
nonce
encrypted_payload
```

---

# 45. History Associated Data

Conceptually:

```text
HistoryAAD =
    protocol_identifier
    || crypto_version
    || vault_id
    || entry_id
    || history_id
    || serialization_version
```

This binds historical ciphertext to both the vault and parent entry.

---

# 46. Deleted Entries

Deletion removes the active database record.

Because stored records contain ciphertext, residual SQLite pages should not contain plaintext secret content.

Secure physical erasure is not guaranteed.

Encrypted backups or historical versions may continue to contain the deleted secret.

The UI should make history/backup implications clear where appropriate.

---

# 47. Aliases

Aliases may reveal sensitive information.

Preferred v1 design:

```text
aliases are stored inside encrypted entry metadata
```

or within another encrypted metadata structure.

The application builds alias resolution structures in memory after unlock.

---

# 48. TOTP Secrets

TOTP seeds are part of the encrypted entry payload.

The seed must never receive weaker encryption than passwords.

Generated TOTP codes are ephemeral and must not be persisted.

---

# 49. Password Generator

Password generation uses:

```go
crypto/rand
```

Selection algorithms must avoid modulo bias.

For a character set of length `N`, implementation should use unbiased random selection rather than simply:

```text
random_byte % N
```

unless rejection sampling or another unbiased construction makes it safe.

---

# 50. Passphrase Generator

Passphrase word selection must use cryptographically secure unbiased random selection.

The wordlist itself is public and may be embedded in the binary.

Entropy estimates must be based on the actual selection process.

---

# 51. Key Lifetime

The Vault Key and derived domain keys exist only while the vault is unlocked.

On lock, GoVault should:

* Remove references.
* Clear mutable buffers where practical.
* Clear cached decrypted metadata.
* Clear decrypted entries.
* Clear domain keys.
* Trigger garbage collection only if justified by measured behavior, not as a claimed security guarantee.

GoVault cannot guarantee perfect erasure from Go runtime memory.

---

# 52. Secret Types in Go

Security-sensitive code should avoid storing long-lived secrets as immutable Go strings where practical.

Preferred conceptual type:

```go
type Secret []byte
```

Security-sensitive APIs should minimize conversions between:

```text
[]byte
string
[]rune
```

because each conversion may create additional copies.

---

# 53. Secret String Formatting

Secret-bearing types must not implement `String()` methods that expose plaintext.

Debug output must not expose secrets.

A safe formatter may return:

```text
[REDACTED]
```

---

# 54. Master Password Change

Changing the master password does not re-encrypt vault records.

Flow:

```text
Current Master Password
        │
        ▼
Old KEK
        │
        ▼
Decrypt Vault Key
        │
        ▼
Generate new KDF salt
        │
        ▼
New Master Password
        │
        ▼
Argon2id
        │
        ▼
New KEK
        │
        ▼
Generate new wrapping nonce
        │
        ▼
Encrypt same Vault Key
        │
        ▼
Atomically update vault header
```

---

# 55. New Salt on Password Change

Every master password change generates a new random KDF salt.

The old salt must not be reused.

---

# 56. New Wrapping Nonce

Every Vault Key rewrap operation generates a fresh 24-byte XChaCha20-Poly1305 nonce.

---

# 57. KDF Upgrade

The master password change mechanism also enables KDF upgrades.

For example:

```text
old:

64 MiB
3 iterations

new:

128 MiB
3 iterations
```

The Vault Key remains unchanged.

Only the wrapping layer changes.

---

# 58. Opportunistic KDF Upgrade

A future version may detect that a vault uses outdated KDF parameters.

After successful unlock, it may offer:

```text
Your vault uses older key-derivation settings.

Upgrade protection?
```

The application must never silently weaken KDF parameters.

---

# 59. Backup Cryptographic Model

GoVault backups must remain encrypted independently of their storage location.

Crypto Suite v1 uses a backup key derived from the Vault Key.

```text
BackupKey =
    HKDF-SHA256(
        VaultKey,
        salt = vault_id,
        info = "govault:v1:backup"
    )
```

---

# 60. Backup Implication

A backup created using a Vault Key-derived Backup Key is cryptographically tied to that vault.

This provides a simple model for v0.x.

However, it creates an important recovery consideration:

> Restoring the backup requires access to cryptographic material capable of deriving the Backup Key.

Therefore the backup container must include enough protected key metadata to permit restoration using the user's master password.

The backup must not assume that the original `vault.db` still exists.

---

# 61. Self-Contained Backup Requirement

A `.gvault` backup must be independently restorable using:

```text
backup file
+
master password valid for that backup
```

No external vault database should be required.

Therefore the backup must include:

```text
vault_id
crypto version
KDF algorithm
KDF parameters
KDF salt
wrapped Vault Key
Vault Key wrapping nonce
encrypted backup payload
backup nonce
authenticated format metadata
```

---

# 62. Backup Master Password Semantics

For v1, a backup is protected using the master-password configuration that existed when the backup was created.

Example:

```text
September 1
Master password = A

September 2
Backup created

September 10
Master password changed to B
```

The September 2 backup remains unlockable using:

```text
A
```

because it contains the historical wrapped Vault Key configuration.

This behavior must be clearly documented.

---

# 63. Why Old Backups Keep Old Passwords

Changing the current vault master password cannot retroactively modify backup files stored elsewhere.

Therefore:

> Changing the master password does not invalidate existing backups.

Users who believe an old master password was compromised must create new backups and securely dispose of old backups under their control.

---

# 64. Backup Encryption

Conceptually:

```text
Backup Payload
      │
      ▼
Canonical Serialization
      │
      ▼
XChaCha20-Poly1305
using BackupKey
      │
      ▼
Encrypted Backup Payload
```

Each backup receives a new random 24-byte nonce.

---

# 65. Backup Associated Data

Conceptually:

```text
BackupAAD =
    protocol_identifier
    || backup_format_version
    || crypto_version
    || vault_id
    || backup_id
```

The backup identifier is random.

---

# 66. Backup Identifier

Each backup receives:

```text
backup_id = random(16 bytes)
```

This ID is authenticated and may be displayed during backup inspection.

---

# 67. Backup Manifest

Sensitive manifest information should reside inside the encrypted payload.

For example:

```text
entry counts
entry names
tags
history information
vault statistics
```

The unencrypted header should contain only information necessary to:

* Parse the file.
* Validate supported versions.
* Derive the KEK.
* Unwrap the Vault Key.
* Authenticate/decrypt the backup.

---

# 68. Backup Verification

There are two useful levels of verification.

## Structural Verification

Can be performed without the master password.

Checks:

```text
magic bytes
supported version
length bounds
header structure
KDF bounds
ciphertext presence
```

This does not prove cryptographic authenticity.

## Cryptographic Verification

Requires the master password.

Checks:

```text
unwrap Vault Key
derive Backup Key
authenticate backup payload
parse decrypted structure
validate internal consistency
```

The CLI should distinguish these modes clearly.

---

# 69. Backup Restore

Restore flow:

```text
Read backup
    │
    ▼
Validate header
    │
    ▼
Validate KDF bounds
    │
    ▼
Request backup master password
    │
    ▼
Derive KEK
    │
    ▼
Authenticate/decrypt wrapped Vault Key
    │
    ▼
Derive Backup Key
    │
    ▼
Authenticate/decrypt backup payload
    │
    ▼
Validate payload
    │
    ▼
Create recovery snapshot
    │
    ▼
Restore to temporary database
    │
    ▼
Verify
    │
    ▼
Atomic replacement
```

---

# 70. Recovery Snapshot

A recovery snapshot should use the same protected storage principles as an encrypted backup.

No plaintext temporary copy of the vault is permitted.

---

# 71. Backup Password Independence — Future

A future format may allow:

```text
govault backup create --separate-password
```

This would use a separately derived backup KEK.

Such functionality is not part of Crypto Suite v1 and requires threat-model review.

---

# 72. Cryptographic Versioning

GoVault must distinguish:

```text
application version
database schema version
vault format version
crypto suite version
backup format version
entry serialization version
```

These are not interchangeable.

Example:

```text
Application             0.4.2
Database Schema         3
Vault Format            1
Crypto Suite            1
Backup Format           1
Entry Serialization     2
```

---

# 73. Crypto Suite Identifier

Crypto Suite v1 should have a stable machine-readable identifier.

Conceptually:

```text
1
```

or:

```text
GOVAULT-CRYPTO-1
```

The exact binary representation belongs in `vault-format.md`.

---

# 74. No Silent Algorithm Negotiation

GoVault does not negotiate cryptographic algorithms with external systems.

A vault explicitly declares its crypto suite.

Unsupported suites result in:

```text
Unsupported GoVault cryptographic format.
```

GoVault must not guess.

---

# 75. No Silent Downgrade

If a vault requires Crypto Suite v2 and the application only supports v1, GoVault must fail.

It must not reinterpret the vault using weaker settings.

---

# 76. Future Crypto Migration

Future versions may introduce:

```text
Crypto Suite v2
```

Migration should follow:

```text
unlock old format
        │
        ▼
authenticate all relevant data
        │
        ▼
create recovery snapshot
        │
        ▼
generate/derive required new keys
        │
        ▼
re-encrypt into new format
        │
        ▼
verify
        │
        ▼
atomic commit
```

The original vault should remain recoverable until migration succeeds.

---

# 77. Cryptographic Agility

Cryptographic agility should be explicit and versioned rather than overly generic.

Avoid designs such as arbitrary user-selectable:

```text
cipher = ...
hash = ...
kdf = ...
```

This increases testing complexity and creates insecure combinations.

Instead:

```text
Crypto Suite 1
=
Argon2id
+ HKDF-SHA-256
+ XChaCha20-Poly1305
```

Future changes create a new reviewed suite.

---

# 78. Authentication Failure

All AEAD authentication failures must fail closed.

The application must never:

* Return partial plaintext.
* Ignore authentication errors.
* Attempt "best effort" recovery.
* Continue parsing unauthenticated secret data.

---

# 79. Corruption Recovery

Cryptographic authentication failure indicates:

```text
wrong key
corruption
tampering
unsupported interpretation
```

Normal user-facing errors should not claim which one occurred unless independently known.

---

# 80. Cryptographic Error Logging

Low-level errors may be useful for debugging, but logs must not contain:

```text
keys
passwords
plaintext
non-redacted payloads
```

Authentication errors may include:

```text
entry ID
operation type
crypto version
```

only if these values are classified as safe metadata.

---

# 81. Constant-Time Operations

GoVault should rely on established cryptographic libraries for authentication and comparison behavior.

Custom MAC comparison logic is prohibited.

Where secret comparisons are necessary outside established primitives, use constant-time operations where appropriate.

---

# 82. Password Comparison During Audit

Reuse detection requires determining whether passwords are equal.

The simplest safe v1 approach is:

1. Decrypt required password values while unlocked.
2. Compare them during the audit.
3. Keep comparison state only in memory.
4. Discard it after completion.

Persistent password fingerprints should not be introduced merely for performance.

---

# 83. Password Strength Analysis

Password strength analysis occurs locally on plaintext available while the vault is unlocked.

Analysis results should not require persisting password-derived fingerprints.

---

# 84. Cryptographic Material Types

Implementation should create distinct Go types where useful.

Conceptually:

```go
type VaultKey [32]byte
type EntryKey [32]byte
type HistoryKey [32]byte
type BackupKey [32]byte
type KeyEncryptionKey [32]byte
type Nonce [24]byte
type VaultID [16]byte
type EntryID [16]byte
```

Distinct types reduce accidental key interchange.

---

# 85. Do Not Use Generic `[]byte` Everywhere

Security-critical APIs should make misuse difficult.

Bad conceptual API:

```go
Encrypt(key []byte, data []byte)
```

Better:

```go
EncryptEntry(
    key EntryKey,
    id EntryID,
    plaintext []byte,
) (...)
```

This creates stronger boundaries between cryptographic domains.

---

# 86. Proposed Crypto Package Structure

```text
internal/crypto/
├── suite.go
├── argon2.go
├── hkdf.go
├── random.go
├── wrap.go
├── entry.go
├── history.go
├── backup.go
├── aad.go
├── types.go
└── errors.go
```

Serialization should preferably remain outside low-level primitive wrappers where possible.

---

# 87. Primitive Wrapper Boundary

Application code should not directly call:

```text
argon2.IDKey
chacha20poly1305.NewX
hkdf
```

throughout the project.

Instead, cryptographic operations should be centralized.

Example:

```text
crypto.DeriveKEK(...)
crypto.WrapVaultKey(...)
crypto.UnwrapVaultKey(...)
crypto.DeriveEntryKey(...)
crypto.EncryptEntry(...)
crypto.DecryptEntry(...)
```

This reduces inconsistent cryptographic usage.

---

# 88. Random API

GoVault should centralize secure randomness.

Conceptually:

```go
func RandomBytes(dst []byte) error
func NewVaultID() (VaultID, error)
func NewEntryID() (EntryID, error)
func NewNonce() (Nonce, error)
func NewVaultKey() (VaultKey, error)
```

All production implementations use:

```go
crypto/rand
```

---

# 89. Testing Randomness

Tests may inject deterministic randomness for reproducible test vectors.

Production code must never use deterministic test RNGs.

Dependency injection boundaries must make this distinction obvious.

---

# 90. Cryptographic Test Vectors

Before stabilizing v1, GoVault must publish test vectors.

Each vector should include:

```text
master password
KDF salt
KDF parameters
expected KEK
Vault Key
wrapping nonce
wrapped Vault Key
vault_id
entry_id
entry plaintext
entry nonce
entry AAD
entry ciphertext
```

This makes the format independently testable.

---

# 91. Backup Test Vectors

Backup vectors should include:

```text
backup_id
vault_id
Vault Key
Backup Key
backup nonce
backup AAD
plaintext payload
encrypted payload
```

---

# 92. Test Vector Warning

Test vectors necessarily contain known keys and passwords.

They must be clearly marked:

```text
TEST DATA ONLY — NEVER USE THESE KEYS
```

They must never be used by production vault generation.

---

# 93. Tampering Tests

Tests must verify authentication failure after changing one bit in:

```text
ciphertext
authentication tag
nonce
AAD
vault ID
entry ID
record type
crypto version
backup ID
```

where the modified field is cryptographically bound.

---

# 94. Cross-Record Tests

Test:

```text
Encrypt Entry A
Copy ciphertext to Entry B
Attempt decrypt using Entry B AAD
```

Expected:

```text
authentication failure
```

---

# 95. Cross-Vault Tests

Test:

```text
Encrypt entry under Vault A
Copy record to Vault B
Attempt decrypt
```

Expected:

```text
authentication failure
```

---

# 96. Wrong Password Tests

Test:

```text
correct master password
incorrect master password
similar master password
empty master password
Unicode master password
very long master password
```

The implementation must behave deterministically and safely.

---

# 97. KDF DoS Tests

Vault headers containing extreme values such as:

```text
memory = 4 TiB
iterations = 4 billion
parallelism = 65535
```

must be rejected before expensive allocation or computation.

---

# 98. Backup Corruption Tests

Test:

```text
truncated header
truncated ciphertext
modified KDF salt
modified wrapped key
modified backup nonce
modified backup ciphertext
unsupported crypto version
unsupported backup version
oversized lengths
```

All must fail safely.

---

# 99. Known Plaintext Persistence Tests

Integration tests should store recognizable secret markers.

Example:

```text
GOVAULT_SECRET_TEST_8F72A
```

Then inspect:

```text
vault.db
vault.db-wal
vault.db-journal
backup.gvault
recovery snapshots
```

The marker must not appear in plaintext.

---

# 100. Logging Tests

Tests should trigger failures while using recognizable secrets.

Then scan logs for those markers.

No secret marker may appear.

---

# 101. Fuzzing

Recommended fuzz targets:

```text
vault header parser
wrapped key parser
entry envelope parser
backup header parser
backup payload parser
AAD serializer
KDF parameter parser
```

Cryptographic primitives themselves should not be reimplemented or fuzzed as custom algorithms.

---

# 102. Memory Considerations

Go does not provide guaranteed secret-memory erasure.

GoVault should still reduce unnecessary exposure.

Guidelines:

* Avoid long-lived plaintext strings.
* Keep decrypted entry lifetime short.
* Avoid caching passwords.
* Clear mutable buffers where practical.
* Avoid duplicating secrets for formatting.
* Never serialize plaintext secrets for debug output.

Documentation must not claim perfect zeroization.

---

# 103. Clipboard Encryption

Clipboard contents cannot remain encrypted while being useful to another application.

Once a password is copied, it temporarily leaves GoVault's cryptographic protection boundary.

GoVault should:

* Copy only on explicit request.
* Clear after timeout.
* Avoid logging.
* Avoid displaying unnecessarily.

---

# 104. Terminal Reveal

Similarly, revealing a password places plaintext into terminal rendering infrastructure.

Cryptography cannot protect it after that point.

The UI must therefore treat reveal as a deliberate temporary operation.

---

# 105. File Permissions

Cryptography remains the primary confidentiality boundary.

Filesystem permissions provide defense in depth.

Expected Unix permissions:

```text
vault database     0600
config             0600
backup default     0600
recovery snapshot  0600
```

GoVault should validate permissions through `doctor`.

---

# 106. Full-Disk Encryption

GoVault does not require full-disk encryption.

However, full-disk encryption is complementary protection against:

* Swap leakage.
* Temporary system files.
* Metadata leakage.
* Other local application data.

Documentation may recommend it without treating it as a prerequisite.

---

# 107. Root Compromise

If an attacker can inspect GoVault memory while unlocked, cryptographic keys may be recoverable.

No at-rest cryptographic design can fully prevent this.

This limitation must remain explicit.

---

# 108. No Recovery Backdoor

There is no developer key.

There is no universal recovery password.

There is no remote recovery service.

There is no escrow key.

If the user loses:

```text
master password
+
all usable backups/recovery material
```

the encrypted vault may be permanently unrecoverable.

---

# 109. Crypto Configuration

Users should not manually select low-level cryptographic primitives.

Do not expose settings such as:

```text
cipher = aes
hash = sha1
argon_iterations = 1
```

Normal users select GoVault format versions implicitly through the application.

Advanced KDF tuning, if introduced, must enforce safe limits.

---

# 110. Crypto Suite v1 Summary

```text
┌──────────────────────────────────────────────┐
│              GoVault Crypto v1              │
├──────────────────────────────────────────────┤
│ Password KDF                                │
│ Argon2id                                    │
│                                              │
│ Default memory       64 MiB                 │
│ Default iterations   3                      │
│ Default parallelism  4                      │
│ KDF salt             32 bytes               │
│ KEK                  32 bytes               │
│                                              │
│ Root Vault Key                              │
│ Random               32 bytes               │
│                                              │
│ Key Derivation                              │
│ HKDF-SHA-256                                │
│                                              │
│ Authenticated Encryption                    │
│ XChaCha20-Poly1305                          │
│                                              │
│ AEAD key             32 bytes               │
│ AEAD nonce           24 bytes random        │
│                                              │
│ Vault ID             16 bytes random        │
│ Entry ID             16 bytes random        │
│ Backup ID            16 bytes random        │
│                                              │
│ Randomness                                  │
│ crypto/rand                                 │
└──────────────────────────────────────────────┘
```

---

# 111. Complete Vault Creation Flow

```text
User chooses master password
           │
           ▼
Generate Vault ID
random 16 bytes
           │
           ▼
Generate Vault Key
random 32 bytes
           │
           ▼
Generate KDF Salt
random 32 bytes
           │
           ▼
Argon2id(master password, salt, params)
           │
           ▼
KEK
           │
           ▼
Generate wrapping nonce
random 24 bytes
           │
           ▼
Create VaultKeyAAD
           │
           ▼
XChaCha20-Poly1305(
    KEK,
    nonce,
    VaultKey,
    VaultKeyAAD
)
           │
           ▼
Wrapped Vault Key
           │
           ▼
Persist:

vault ID
crypto version
KDF algorithm
KDF parameters
KDF salt
wrapping nonce
wrapped Vault Key
```

The master password, KEK, and plaintext Vault Key must not be written to disk.

---

# 112. Complete Entry Encryption Flow

```text
Unlocked Vault Key
       │
       ▼
HKDF-SHA-256
"govault:v1:entry"
       │
       ▼
Entry Key
       │
       │
       ├───────────────┐
       │               │
       ▼               ▼
Generate Entry ID   Serialize Entry
       │               │
       └──────┬────────┘
              ▼
        Generate nonce
        random 24 bytes
              │
              ▼
         Build EntryAAD
              │
              ▼
       XChaCha20-Poly1305
              │
              ▼
       encrypted payload
```

Persist:

```text
entry_id
nonce
ciphertext
required routing/version metadata
```

---

# 113. Complete Entry Decryption Flow

```text
Load record
    │
    ▼
Validate structure
    │
    ▼
Validate crypto version
    │
    ▼
Build canonical EntryAAD
    │
    ▼
Derive Entry Key
    │
    ▼
XChaCha20-Poly1305 Open
    │
    ├── authentication failure → reject
    │
    ▼
Parse plaintext
    │
    ▼
Validate semantic structure
    │
    ▼
Return entry
```

---

# 114. Complete Master Password Change Flow

```text
Request current password
       │
       ▼
derive old KEK
       │
       ▼
unwrap Vault Key
       │
       ▼
request new password twice
       │
       ▼
generate new KDF salt
       │
       ▼
derive new KEK
       │
       ▼
generate new wrapping nonce
       │
       ▼
wrap SAME Vault Key
       │
       ▼
write new header state
to temporary/transactional storage
       │
       ▼
verify new header
       │
       ▼
atomic commit
```

Entry ciphertext does not change.

---

# 115. Complete Backup Flow

```text
Unlocked Vault
     │
     ▼
Generate Backup ID
     │
     ▼
Derive Backup Key
from Vault Key
     │
     ▼
Serialize backup payload
     │
     ▼
Generate backup nonce
     │
     ▼
Build BackupAAD
     │
     ▼
Encrypt backup payload
     │
     ▼
Create self-contained header containing:

vault ID
backup ID
crypto version
backup format version
KDF parameters
KDF salt
wrapped Vault Key
wrapping nonce

     │
     ▼
Write temporary .gvault
     │
     ▼
Verify
     │
     ▼
Atomic rename
```

---

# 116. Crypto Code Review Rules

Any change touching:

```text
internal/crypto/
vault header parsing
backup header parsing
AAD construction
key derivation
nonce generation
serialization used by crypto
```

must receive security-focused review.

Reviewers should specifically check:

```text
nonce reuse
key reuse
domain separation
AAD consistency
error handling
secret logging
integer bounds
format compatibility
randomness source
```

---

# 117. Dependency Rules

The cryptographic core should depend on as little as possible.

Expected dependencies:

```text
Go standard library
golang.org/x/crypto
```

Additional dependencies require justification.

A large serialization framework should not be introduced into the cryptographic core merely for convenience.

---

# 118. Cryptographic Compatibility Contract

Once GoVault reaches v1.0:

> A vault created with Crypto Suite v1 must remain readable by future versions unless the format is formally retired with an explicit migration path.

Therefore serialized cryptographic structures must not depend directly on unstable Go struct layouts.

---

# 119. Open Questions for Vault Format Design

The cryptographic primitives are now selected, but the next document must define exact byte-level representation.

`docs/vault-format.md` must answer:

1. What are the database tables?
2. What is the exact vault header structure?
3. How are binary values represented in SQLite?
4. How is canonical AAD serialized?
5. How is an encrypted entry envelope represented?
6. Which metadata remains plaintext?
7. How are timestamps represented?
8. How are entry types encoded?
9. How are aliases represented?
10. How are history records linked?
11. How are migrations represented?
12. How are format versions stored?
13. How are unknown fields handled?
14. How are payload sizes bounded?
15. How is corruption distinguished from unsupported format?
16. How is atomic master-password rewrap persisted?
17. How is an in-memory search index reconstructed?

---

# 120. Security Decisions Established by This Document

The following decisions are now part of the proposed GoVault design:

```text
✓ Argon2id for master-password derivation

✓ 32-byte random KDF salt

✓ 32-byte random Vault Key

✓ Vault Key separated from master password

✓ XChaCha20-Poly1305 authenticated encryption

✓ 24-byte random nonces

✓ HKDF-SHA-256 domain separation

✓ Dedicated Entry Key

✓ Dedicated History Key

✓ Dedicated Backup Key

✓ Random 128-bit Vault IDs

✓ Random 128-bit Entry IDs

✓ Random 128-bit Backup IDs

✓ Context-bound AEAD Associated Data

✓ Cross-record substitution protection

✓ Cross-vault substitution protection

✓ Encrypted human-readable metadata by default

✓ In-memory search index after unlock

✓ Self-contained encrypted backups

✓ Backups remain tied to the master-password
  configuration present when created

✓ Master password changes rewrap the Vault Key

✓ No custom cryptographic primitives

✓ No user-selectable cipher combinations

✓ Explicit crypto suite versioning
```

---

# 121. Required External Review Before v1

The GoVault project must treat this document as an engineering design, not as proof that the implementation is secure.

Before declaring the storage format stable for v1.0, the project should seek independent review of:

```text
threat model
key hierarchy
Argon2id parameters
AEAD usage
AAD design
backup format
vault format
serialization
nonce handling
migration strategy
implementation
```

Security-sensitive cryptographic formats become expensive to change once real users depend on them.

---

# 122. Next Specification

The next design document should be:

```text
docs/vault-format.md
```

It will translate this cryptographic architecture into the concrete persistent representation used by SQLite.

After that:

```text
docs/backup-format.md
```

should specify the `.gvault` container independently from the SQLite database.

The intended design sequence is therefore:

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

