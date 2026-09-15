# GoVault Threat Model

**Document:** `docs/threat-model.md`
**Project:** GoVault
**Status:** Draft
**Applies to:** GoVault v0.1+
**Last Updated:** September 2026

---

# 1. Purpose

This document defines the security threat model for GoVault.

GoVault is an offline, terminal-native password and secrets manager designed to protect sensitive information stored on a user's local machine.

The purpose of this document is to establish:

* What GoVault is designed to protect.
* Which attackers are considered.
* Which attacks are in scope.
* Which attacks are outside GoVault's security boundary.
* Which security guarantees GoVault intends to provide.
* Which guarantees GoVault explicitly cannot provide.
* Which design decisions follow from these assumptions.

This document should be treated as a security specification rather than only as descriptive documentation.

Future architectural, cryptographic, storage, backup, CLI, and TUI decisions should be evaluated against this threat model.

---

# 2. Product Security Goal

GoVault's primary security goal is:

> An attacker who obtains the GoVault database, encrypted backups, or configuration files should not be able to recover protected secrets without the user's master password or equivalent decrypted key material.

GoVault should additionally minimize accidental disclosure of secrets during normal terminal usage.

---

# 3. Security Philosophy

GoVault follows several security principles.

## 3.1 Offline by Design

GoVault does not depend on network connectivity.

The application must not intentionally:

* Contact remote APIs.
* Upload secrets.
* Synchronize vaults.
* Send telemetry.
* Send analytics.
* Check for remote updates.
* Retrieve favicons.
* Perform remote breach checks.
* Resolve metadata from websites.

The absence of network functionality reduces the attack surface and limits opportunities for accidental secret transmission.

---

## 3.2 Assume Local Files Can Be Stolen

GoVault assumes that an attacker may obtain copies of:

```text
vault.db
config.toml
encrypted backups
application logs
temporary files
```

The vault must remain protected under this scenario.

---

## 3.3 Do Not Trust Storage Encryption Alone

SQLite is a persistence mechanism, not a security boundary.

Sensitive data must be encrypted by GoVault before being stored.

Filesystem encryption such as:

```text
LUKS
FileVault
BitLocker
ZFS encryption
```

may provide additional protection but must not be required for GoVault's confidentiality guarantees.

---

## 3.4 Minimize Secret Exposure

GoVault should minimize the amount of time secrets exist in:

* Terminal output.
* Clipboard contents.
* Application memory.
* Temporary buffers.
* Logs.
* Error messages.

---

## 3.5 Fail Closed

If cryptographic verification fails, GoVault must not attempt to recover or partially interpret protected content.

Examples:

```text
authentication failure
corrupted ciphertext
unsupported crypto version
invalid backup authentication
invalid vault key
```

should result in safe failure.

---

# 4. Assets

The following assets require protection.

## 4.1 Master Password

The user's master password provides access to the vault key.

It is one of the highest-value secrets in the system.

GoVault must never persist the master password.

---

# 4.2 Vault Key

The Vault Key is the cryptographic key used to protect vault contents.

Compromise of this key may compromise all encrypted vault entries.

It must therefore receive the same protection level as the master password.

---

# 4.3 Passwords

Passwords stored in GoVault are confidential data.

They must never be persisted in plaintext.

---

# 4.4 TOTP Secrets

TOTP seeds allow generation of authentication codes.

They must be treated as credentials, not merely metadata.

TOTP secrets must always be encrypted.

---

# 4.5 API Keys and Tokens

API credentials may provide direct access to infrastructure and services.

They must be treated as high-sensitivity secrets.

---

# 4.6 SSH Private Keys

If SSH key storage is implemented, private keys must be encrypted as sensitive vault content.

---

# 4.7 Secure Notes

Secure notes may contain:

* Recovery codes.
* PINs.
* Account information.
* Infrastructure details.
* Private personal information.

Secure notes must be treated as encrypted secrets.

---

# 4.8 Custom Fields

Custom fields marked as sensitive must receive the same protection as passwords.

---

# 4.9 Entry History

Historical entry versions may contain old passwords or secrets.

History must therefore be encrypted with the same confidentiality guarantees as current entries.

---

# 4.10 Backups

GoVault backups contain copies of vault data.

A backup must provide security equivalent to or stronger than the primary vault.

Backups must never contain plaintext secrets.

---

# 5. Security Boundary

The GoVault security boundary includes:

```text
GoVault process
encrypted vault format
crypto layer
backup format
local database
local configuration
TUI
CLI
clipboard integration
```

The following components are outside GoVault's direct control:

```text
operating system kernel
terminal emulator
shell
clipboard manager
window manager
desktop environment
filesystem implementation
hardware
firmware
malware running with sufficient privileges
physical machine security
```

GoVault may mitigate risks involving these components but cannot provide absolute guarantees against compromise of them.

---

# 6. Threat Actors

GoVault considers several attacker models.

---

# 6.1 Offline File Attacker

The attacker obtains copies of:

```text
vault.db
config.toml
backup.gvault
```

but does not control the running machine.

Examples:

* Stolen laptop drive.
* Exposed backup disk.
* Accidentally uploaded backup.
* Compromised filesystem snapshot.
* Leaked home directory.
* Stolen USB drive.

This attacker is fully in scope.

---

# 6.2 Local Unprivileged Attacker

The attacker has access to the same operating system but does not have administrator/root privileges.

Potential capabilities:

* Read world-readable files.
* Observe some process metadata.
* Attempt to inspect temporary files.
* Observe command-line arguments.
* Monitor insecure file permissions.

This attacker is partially in scope.

GoVault should use restrictive permissions and avoid exposing secrets through command arguments or files.

---

# 6.3 Shoulder-Surfing Attacker

The attacker can visually observe the terminal.

This attacker may:

* Observe revealed passwords.
* Observe TOTP codes.
* Observe usernames.
* Observe search results.
* Observe entry names.

This attacker is partially in scope.

GoVault should mask secrets by default.

---

# 6.4 Clipboard Observer

The attacker or another application can read clipboard contents.

This risk is partially in scope.

GoVault should minimize exposure by:

* Copying only when explicitly requested.
* Clearing copied secrets after a timeout where possible.
* Avoiding clipboard use by default.

GoVault cannot guarantee that clipboard history managers or malicious applications did not capture the secret before it was cleared.

---

# 6.5 Terminal Logging Attacker

The user's terminal may be:

* Recorded.
* Logged.
* Captured by a multiplexer.
* Shared.
* Stored in scrollback.
* Included in support logs.

This risk is partially in scope.

GoVault should avoid printing secrets by default.

---

# 6.6 Malicious Backup Recipient

An attacker obtains an encrypted GoVault backup.

This is fully in scope.

The backup must not reveal secrets without the required credentials.

---

# 6.7 Database Tampering Attacker

An attacker modifies:

```text
vault.db
backup files
encrypted records
metadata
```

and causes GoVault to open the modified data.

This is fully in scope.

GoVault must detect unauthorized modification of authenticated encrypted content.

---

# 6.8 Compromised Host Attacker

The attacker controls the machine while the vault is unlocked.

Potential capabilities:

* Keylogging.
* Reading process memory.
* Reading clipboard contents.
* Injecting code.
* Replacing the GoVault binary.
* Modifying the operating system.
* Capturing terminal output.
* Reading decrypted data.

This attacker is outside GoVault's practical security boundary.

GoVault must clearly document that a fully compromised host can defeat application-level protections.

---

# 7. Threat Scenarios

---

# 7.1 Stolen Vault Database

## Scenario

An attacker obtains:

```text
~/.local/share/govault/vault.db
```

## Security Requirement

The attacker must not be able to recover:

* Passwords.
* TOTP secrets.
* Secure notes.
* API secrets.
* SSH private keys.
* Sensitive custom fields.
* Entry history.

without the master password or decrypted key material.

## Mitigations

* Strong password-based key derivation.
* Random per-vault salt.
* High-cost Argon2id parameters.
* Separate Vault Key.
* Authenticated encryption.
* No plaintext secret fields.

---

# 7.2 Stolen Backup

## Scenario

The user stores:

```text
govault-backup.gvault
```

on external media which is later stolen.

## Security Requirement

The attacker must not gain useful plaintext vault content.

The backup must not weaken the cryptographic security of the primary vault.

---

# 7.3 Brute-Force Attack Against Master Password

## Scenario

An attacker has the encrypted vault and attempts password guesses offline.

Because the attacker has a local copy, rate limiting inside GoVault cannot prevent this attack.

## Mitigation

GoVault must use a deliberately expensive password-based key derivation function.

Current design:

```text
Argon2id
```

The parameters should consume substantial memory and computation while remaining practical for legitimate users.

Parameters must be stored in the vault header so they may evolve over time.

---

# 7.4 Weak Master Password

## Scenario

The user selects:

```text
password123
```

as the master password.

Even a strong KDF cannot fully compensate for extremely weak passwords.

## Mitigation

During vault creation, GoVault should:

* Estimate password strength locally.
* Warn against weak master passwords.
* Encourage long passwords or passphrases.

GoVault should avoid arbitrary composition requirements such as:

```text
must contain exactly one uppercase letter
must contain one symbol
```

unless justified.

Length and unpredictability should be emphasized.

---

# 7.5 Vault Tampering

## Scenario

An attacker modifies encrypted bytes in the database.

## Security Requirement

GoVault must detect modification before returning plaintext.

## Mitigation

Use authenticated encryption.

Modified ciphertext must fail authentication.

---

# 7.6 Record Replacement

## Scenario

An attacker replaces the encrypted payload for Entry A with the payload belonging to Entry B.

If encryption authenticates only the ciphertext but not the record identity, GoVault may incorrectly interpret one valid encrypted record as another.

## Mitigation

Encryption should authenticate contextual metadata using AEAD Associated Data.

Potential authenticated context:

```text
record UUID
record type
schema version
vault identifier
crypto format version
```

Exact fields will be defined in `cryptography.md`.

---

# 7.7 Record Deletion

## Scenario

An attacker deletes an encrypted entry from SQLite.

Authenticated encryption cannot detect absence of an independent record by itself.

Potential mitigations include:

* Database-level integrity checks.
* Backup comparison.
* Optional authenticated vault index.
* Audit metadata.

Deletion detection is not guaranteed in the initial design unless explicitly implemented.

This limitation should be documented.

---

# 7.8 Database Rollback

## Scenario

An attacker replaces the current vault database with an older valid copy.

Because the older database contains valid authenticated ciphertext, ordinary AEAD verification may succeed.

Preventing rollback attacks reliably requires trusted external state or monotonic counters protected outside the replaced database.

GoVault's offline, self-contained architecture makes this difficult.

Rollback attacks are therefore not fully prevented in the initial threat model.

Users can mitigate this risk with trusted backups and filesystem protections.

---

# 7.9 Backup Rollback

Similarly, GoVault cannot determine that an authenticated backup is outdated unless the user or external trusted state provides that information.

The UI should clearly display backup creation timestamps.

---

# 7.10 Malicious SQLite Content

## Scenario

An attacker modifies non-secret SQLite fields to malformed values.

## Mitigation

GoVault should:

* Validate all database input.
* Use schema constraints.
* Treat database content as untrusted.
* Avoid unsafe deserialization.
* Reject invalid enum values.
* Bound allocation sizes.

Encrypted payload parsing must also treat decrypted serialized content as potentially malformed due to software bugs or version mismatch.

---

# 7.11 Malicious Backup File

## Scenario

The user attempts to restore a crafted `.gvault` file.

## Mitigation

The parser must:

* Validate magic values.
* Validate format versions.
* Validate length fields.
* Bound allocations.
* Authenticate before trusting protected content.
* Reject trailing or malformed structures.
* Avoid arbitrary file extraction.

Backup parsing should be fuzz-tested.

---

# 7.12 Interrupted Backup Creation

## Scenario

The system loses power during:

```text
govault backup create
```

## Mitigation

Backup creation should follow:

```text
write temporary file
→ flush
→ verify
→ atomic rename
```

Incomplete backup files must not replace valid backups.

---

# 7.13 Interrupted Restore

## Scenario

Power is lost during restore.

## Mitigation

Before restoring:

1. Validate backup.
2. Create local recovery snapshot.
3. Restore into temporary storage.
4. Validate restored database.
5. Atomically replace the existing vault where possible.

The original vault should remain recoverable.

---

# 7.14 Interrupted Master Password Change

## Scenario

The process crashes while changing the master password.

## Security Requirement

The vault must remain unlockable with either the old or successfully committed new configuration, never an inconsistent mixture.

## Mitigation

The operation must be transactional or use atomic replacement.

---

# 7.15 Password Printed to Terminal

## Scenario

A CLI command prints:

```text
MySecretPassword
```

to stdout.

This may leak to:

* Scrollback.
* Terminal recording.
* Shell pipelines.
* Logs.
* Screen sharing.

## Mitigation

GoVault must not print secrets by default.

Preferred:

```bash
govault copy github
```

instead of:

```bash
govault get github --password
```

If explicit plaintext output is supported, it must require deliberate invocation.

---

# 7.16 Secrets in Command Arguments

## Scenario

A user runs:

```bash
govault add --password MySecret123
```

Command arguments may be visible in:

* Shell history.
* Process listings.
* Audit logs.

## Requirement

GoVault should not accept high-value secrets directly as ordinary command-line flags unless there is a carefully documented exceptional mode.

Passwords should instead be entered through:

* Interactive hidden prompt.
* Secure stdin flow where appropriate.

---

# 7.17 Master Password in Command Arguments

The following must never be supported:

```bash
govault unlock --master-password hunter2
```

The master password should be accepted through a hidden interactive input.

Potential automation mechanisms require a separate security design.

---

# 7.18 Secret Leakage Through Environment Variables

Environment variables may leak through:

* Process inspection.
* Crash reporting.
* Debugging.
* Child processes.

GoVault should not recommend storing the master password in environment variables.

---

# 7.19 Secret Leakage Through Logs

Application logs must never contain:

```text
master passwords
passwords
TOTP seeds
API keys
private keys
Vault Keys
decrypted secure notes
sensitive custom fields
```

Structured logging must apply the same restrictions.

---

# 7.20 Secret Leakage Through Errors

Bad:

```text
Failed to encrypt password "MySecret123"
```

Good:

```text
Unable to encrypt entry.
```

Errors must not interpolate sensitive values.

---

# 7.21 Secret Leakage Through Panic Output

Unexpected panics may expose application state.

Sensitive values should not implement unsafe string formatting behavior.

Types holding secrets should avoid convenient `String()` implementations that reveal their content.

---

# 7.22 Clipboard Persistence

## Scenario

The user copies a password.

The password remains in clipboard history.

## Mitigation

GoVault should attempt to clear the clipboard after a configurable timeout.

However:

> GoVault cannot guarantee deletion from third-party clipboard history managers.

This limitation must be documented.

---

# 7.23 Clipboard Race

GoVault may set the clipboard to a secret, wait 15 seconds, then attempt to clear it.

If the user has copied something else in the meantime, GoVault must not erase unrelated clipboard content.

Preferred behavior:

1. Copy secret.
2. Wait.
3. Read clipboard if supported.
4. Clear only if clipboard still contains the value GoVault placed there.

Platform limitations may affect this feature.

---

# 7.24 Secret Reveal

Pressing:

```text
r
```

may reveal a secret.

The reveal should:

* Require explicit user action.
* Be temporary.
* Automatically re-mask on navigation.
* Automatically re-mask after timeout.

---

# 7.25 TOTP Display

TOTP codes are credentials.

GoVault should allow users to configure whether TOTP codes are:

```text
visible by default
```

or:

```text
masked until requested
```

Default should favor masking.

---

# 7.26 Search Metadata Leakage

If entry names and tags are stored plaintext, an attacker obtaining `vault.db` may learn information such as:

```text
Bank Account
Production AWS
Company VPN
Personal Email
```

even without passwords.

This may be sensitive.

The metadata encryption policy must therefore be explicitly defined.

Potential strategies:

### Option A

Encrypt everything.

Advantages:

* Strong confidentiality.

Disadvantages:

* Requires decrypting/indexing after unlock.
* More complex search.

### Option B

Leave selected metadata plaintext.

Advantages:

* Faster indexing.
* Simpler queries.

Disadvantages:

* Metadata leakage.

The recommended direction is to encrypt human-readable identifying metadata unless a strong reason exists not to.

---

# 7.27 Timestamps

Timestamps such as:

```text
created_at
updated_at
```

may leak activity patterns.

Whether timestamps remain plaintext or encrypted should be explicitly documented.

This leakage is lower severity than secret content but remains part of the threat model.

---

# 7.28 Entry Count Leakage

The SQLite database may reveal:

* Number of records.
* Approximate vault size.
* Approximate history size.

Preventing this would significantly complicate storage and is not currently a design goal.

Entry count leakage is acceptable for v1 unless requirements change.

---

# 7.29 Entry Size Leakage

Ciphertext length can reveal approximate plaintext length.

Padding could reduce this leakage but introduces complexity and storage overhead.

GoVault does not initially guarantee concealment of plaintext size.

This should be considered an accepted residual risk.

---

# 7.30 Deleted Secrets

SQLite may retain deleted pages depending on configuration and database behavior.

Because sensitive fields are encrypted before storage, residual deleted database pages should contain ciphertext rather than plaintext.

However, decrypted values must never be written to temporary database tables or plaintext files.

---

# 7.31 SQLite Temporary Files

SQLite may create:

```text
journal files
WAL files
shared-memory files
temporary databases
```

All sensitive content written through SQLite must already be encrypted.

Therefore temporary SQLite artifacts should not expose plaintext secrets.

---

# 7.32 Filesystem Permissions

On Unix-like systems, GoVault should attempt restrictive permissions.

Expected:

```text
vault.db      0600
config.toml   0600
```

Directories should avoid unnecessary access by other users.

GoVault Doctor should report insecure permissions.

---

# 7.33 Symlink Attacks

Sensitive file creation should avoid unsafe handling of attacker-controlled symbolic links where practical.

This is particularly important for:

* Backups.
* Temporary files.
* Recovery snapshots.
* Configuration.
* Vault initialization.

---

# 7.34 Path Traversal

Backup and restore logic should not treat archive-controlled paths as arbitrary filesystem destinations.

A GoVault backup should preferably be a structured container rather than a generic archive that extracts arbitrary paths.

---

# 7.35 Malicious Configuration

Configuration is treated as untrusted local input.

GoVault must validate:

```text
timeouts
paths
numeric values
theme values
generator settings
```

Malformed configuration must not cause unsafe behavior.

---

# 7.36 Insecure Auto-Lock Configuration

Users may configure:

```text
auto_lock = "never"
```

This weakens protection.

If supported, GoVault should make this explicit.

Example:

```text
Warning: automatic locking is disabled.
```

---

# 7.37 Process Memory

While the vault is unlocked, decrypted secrets may exist in process memory.

Go's runtime does not guarantee reliable zeroization of every copy because:

* Values may be copied.
* Garbage collection may move or retain objects.
* Compiler optimizations may affect clearing.
* Strings are immutable.

GoVault should minimize unnecessary secret copies.

Where practical:

* Prefer byte slices over strings for long-lived secrets.
* Clear mutable buffers after use.
* Avoid unnecessary conversions.
* Keep decrypted data lifetime short.

However, GoVault must not claim guaranteed memory erasure.

---

# 7.38 Swap and Hibernation

The operating system may write process memory to:

```text
swap
hibernation images
crash dumps
```

GoVault cannot reliably prevent this across all supported platforms.

Users requiring stronger protection should configure operating-system-level protections such as encrypted swap.

---

# 7.39 Core Dumps

Core dumps may contain decrypted secrets.

GoVault should evaluate whether core dumps can reasonably be disabled for the process on supported systems.

At minimum, documentation should mention the risk.

---

# 7.40 Debuggers

A debugger attached to the GoVault process while unlocked may read sensitive memory.

A process with sufficient debugging privileges is considered equivalent to a compromised host.

---

# 7.41 Malicious Terminal Emulator

A terminal emulator can observe:

* Typed master passwords in some threat scenarios.
* Revealed secrets.
* TUI contents.

A malicious terminal emulator is outside GoVault's security boundary.

---

# 7.42 Keylogger

A keylogger may capture:

```text
master password
search terms
typed secrets
```

Protecting against a compromised input stack is outside GoVault's practical scope.

---

# 7.43 Binary Replacement

An attacker may replace the GoVault executable with a malicious version that records the master password.

This is outside the application's internal security boundary.

Distribution integrity and package signing may mitigate this risk in future releases.

---

# 7.44 Dependency Supply-Chain Attack

A malicious dependency could compromise secret handling.

Mitigations:

* Minimize dependencies.
* Prefer Go standard library.
* Pin dependency versions.
* Review security-sensitive dependencies.
* Use `go.sum`.
* Regularly audit dependencies.
* Avoid unnecessary packages.

The cryptographic core should have a very small dependency surface.

---

# 7.45 Networking Dependency Introduction

A contributor may accidentally introduce:

```go
net/http
```

or another networking client.

This would violate the offline guarantee.

Mitigation:

CI should inspect production imports and fail when unauthorized networking dependencies are introduced.

---

# 7.46 DNS or Socket Activity Through Dependencies

The offline guarantee should apply to behavior, not merely direct imports.

Dependencies that perform:

* Update checks.
* Analytics.
* Remote metadata lookup.
* DNS resolution.

must not be included.

---

# 7.47 Denial of Service Through KDF Parameters

Because Argon2id parameters may be stored in vault metadata, an attacker modifying unauthenticated KDF parameters might specify extreme values and force excessive memory allocation.

Mitigation:

GoVault must validate KDF parameters against safe maximum bounds before allocation.

Example concept:

```text
minimum accepted memory
maximum accepted memory
minimum iterations
maximum iterations
supported parallelism range
```

Exact values belong in `cryptography.md`.

---

# 7.48 Downgrade Attack

An attacker may modify metadata to request:

```text
weaker crypto version
lower KDF parameters
legacy format
```

GoVault must not silently downgrade security.

Cryptographic version information and parameters should be authenticated where technically appropriate.

Unsupported insecure formats must fail closed.

---

# 7.49 Nonce Reuse

AEAD schemes may catastrophically fail if nonces are reused incorrectly.

Nonce generation must follow the requirements of the selected AEAD algorithm.

The exact nonce strategy must be documented and tested.

---

# 7.50 Random Number Generator Failure

All cryptographic randomness must originate from:

```go
crypto/rand
```

GoVault must not fall back to weak PRNGs.

Failure to obtain cryptographically secure randomness should fail the operation.

---

# 8. Backup Threat Model

Backups require special consideration because users may place them in less trusted locations.

Examples:

```text
USB drive
NAS
external disk
cloud-synchronized directory
email attachment
another computer
```

GoVault does not control what the user does with a backup.

Therefore `.gvault` backup files must remain cryptographically protected independently of storage location.

---

# 9. Backup Security Requirements

A backup must provide:

* Confidentiality.
* Integrity.
* Authentication.
* Versioning.
* Corruption detection.

A backup must not contain plaintext secrets.

---

# 10. Backup Password Model

The backup cryptographic design must determine whether backups:

### Model A

Use the vault's existing cryptographic key hierarchy.

or:

### Model B

Use a separate backup password.

or:

### Model C

Support both.

This decision belongs in `cryptography.md`.

For the initial version, using the existing vault key hierarchy is simpler, but portability and independent recovery requirements should be carefully considered.

---

# 11. Restore Threat Model

Restore is a high-risk operation because it can replace trusted state.

Restore must:

1. Parse safely.
2. Authenticate the backup.
3. Validate format version.
4. Validate cryptographic parameters.
5. Validate semantic structure.
6. Create a recovery snapshot.
7. Restore transactionally.
8. Preserve the original vault on failure.

---

# 12. Recovery Snapshot Threat Model

Recovery snapshots contain sensitive data.

Therefore they must:

* Be encrypted.
* Have restrictive permissions.
* Be clearly managed.
* Have a defined retention policy.

A recovery snapshot must not create an unencrypted temporary copy of the vault.

---

# 13. CLI Threat Model

The CLI introduces unique exposure risks.

Avoid:

```bash
govault add --password secret
govault unlock --password master
govault totp --seed ABCDEF
```

Prefer hidden interactive prompts.

---

# 14. Shell History

GoVault must assume shell commands may be persisted.

Documentation should discourage users from placing secrets directly in command lines.

---

# 15. Pipelines

If GoVault eventually allows:

```bash
govault reveal github | some-command
```

the user is explicitly exporting a secret beyond the GoVault security boundary.

Such functionality should:

* Require explicit invocation.
* Be clearly documented.
* Avoid accidental use.

---

# 16. TUI Threat Model

The TUI should avoid persistent plaintext exposure.

Required behavior:

* Passwords masked by default.
* TOTP optionally masked.
* Secrets hidden on navigation.
* Secrets hidden after timeout.
* Vault locks after inactivity.
* Screen is redrawn on lock.

---

# 17. Terminal Scrollback

Redrawing the terminal cannot guarantee that a terminal emulator has erased historical scrollback.

GoVault should not claim that clearing the screen permanently removes previously displayed secrets.

This is another reason to minimize explicit reveal operations.

---

# 18. Auto-Lock Threat Model

Auto-lock protects against unattended terminals.

It does not protect against:

* A malicious process already reading memory.
* A keylogger.
* An attacker controlling the host.

Auto-lock should be based primarily on user inactivity within GoVault rather than assuming accurate global OS activity information.

---

# 19. Lock Behavior

When the vault locks, GoVault should:

* Drop the active Vault Key reference.
* Remove decrypted entries from application state.
* Clear sensitive mutable buffers where practical.
* Mask all secrets.
* Remove sensitive TUI state.
* Attempt clipboard cleanup where appropriate.
* Return to lock screen.

---

# 20. Unlock Failure

Authentication failure should not distinguish among unnecessarily specific causes such as:

```text
wrong password
wrong vault key
invalid wrapped key
```

when doing so could expose useful information.

Preferred:

```text
Unable to unlock vault.
Check the master password and vault integrity.
```

Detailed diagnostics may be available through carefully designed internal or doctor workflows without exposing secrets.

---

# 21. Local Rate Limiting

GoVault may add increasing delays after repeated interactive unlock failures to reduce casual guessing against a running session.

However:

> Local rate limiting does not protect against offline brute force of a stolen vault.

The primary defense remains the KDF and master password strength.

---

# 22. Multi-User Systems

On systems shared by multiple OS users, GoVault should:

* Store data under the user's private application directory.
* Use restrictive permissions.
* Warn when permissions are unsafe.

GoVault does not attempt to protect a user's vault from `root`, Administrator, or equivalent privileged accounts while the vault is unlocked.

---

# 23. Physical Access

If the machine is powered off and the attacker obtains storage, GoVault's encrypted vault should remain protected.

If the machine is unlocked and GoVault is unlocked, physical access may allow secret extraction.

GoVault cannot replace full-disk encryption or operating-system session locking.

---

# 24. Data Confidentiality Classification

GoVault data should be divided conceptually into:

## Class A — Critical Secrets

Always encrypted.

Examples:

```text
password
master-derived keys
Vault Key
TOTP seed
API secret
private SSH key
secure note body
sensitive custom fields
historical passwords
```

## Class B — Sensitive Metadata

Prefer encrypted.

Examples:

```text
entry name
username
URL
tags
alias
notes metadata
```

## Class C — Operational Metadata

May remain plaintext if justified.

Examples:

```text
schema version
crypto format version
KDF salt
Argon2 parameters
encrypted Vault Key
non-secret format identifiers
```

Every persistent field should have an explicit classification.

---

# 25. Metadata Policy

The default design direction is:

> Human-readable vault metadata should be encrypted unless it is required before unlock or required for cryptographic initialization.

Therefore the locked database should ideally not reveal entries such as:

```text
Production AWS
Personal Bank
Company VPN
```

---

# 26. Pre-Unlock Metadata

Some information must remain available before the vault can be decrypted.

Examples may include:

```text
vault format version
KDF algorithm
KDF salt
KDF parameters
wrapped Vault Key
crypto version
```

These fields are not confidential by design.

They must still be validated carefully.

---

# 27. Authentication of Metadata

Where relevant, metadata that influences cryptographic interpretation must be authenticated.

This is important for:

* Crypto version.
* Record identity.
* Schema/serialization version.
* Algorithm parameters.
* Vault identity.

The precise authenticated-data design belongs in `cryptography.md`.

---

# 28. Secret Lifetime

The guiding principle is:

> Decrypt as late as possible and discard as early as possible.

Examples:

Search results should not necessarily require every password to remain decrypted in memory.

The application should distinguish:

```text
searchable metadata
```

from:

```text
secret payload
```

while still respecting the metadata confidentiality requirements.

---

# 29. Full-Vault Decryption vs Per-Record Decryption

GoVault should prefer per-record encrypted payloads rather than decrypting the entire vault into a large in-memory plaintext structure.

Benefits:

* Smaller secret exposure window.
* Lower memory footprint.
* Easier record-level updates.
* Easier history management.

Searchable metadata may require a separate design.

---

# 30. Cache Policy

If decrypted metadata is cached after unlock, the cache must:

* Exist only while unlocked.
* Never be written plaintext to disk.
* Be cleared on lock.
* Avoid containing secret fields unless required.

---

# 31. Temporary Files

GoVault should avoid plaintext temporary files.

If temporary files are unavoidable, they require:

* Restrictive permissions.
* Minimal lifetime.
* Secure cleanup where practical.

The preferred design is to avoid them entirely for decrypted data.

---

# 32. Import Threat Model

Future import functionality may ingest untrusted files.

Potential risks:

* Malformed JSON.
* Malformed CSV.
* Extremely large records.
* Unexpected encodings.
* Path manipulation.
* Memory exhaustion.

Imports should be parsed defensively and fuzz-tested.

---

# 33. Export Threat Model

Plaintext export is inherently dangerous.

If GoVault eventually supports plaintext export:

```text
CSV
JSON
```

it must require strong explicit confirmation.

The UI should clearly state that the resulting file is not encrypted.

Encrypted backup should remain the default export mechanism.

---

# 34. Password Generator Threat Model

The password generator must use cryptographically secure randomness.

Required:

```go
crypto/rand
```

Prohibited:

```go
math/rand
timestamp-derived randomness
custom PRNGs
```

---

# 35. Passphrase Generator Threat Model

Word selection must be uniformly random over the configured wordlist unless the design documents another sound distribution.

The embedded wordlist itself is not secret.

---

# 36. Audit Threat Model

The offline security audit may need to compare passwords.

Plaintext passwords should only exist in memory for the duration necessary.

GoVault should avoid persisting:

```text
password hashes for reuse detection
```

if those hashes could become useful to an attacker obtaining the database.

Reuse detection can occur ephemerally during audit.

---

# 37. Audit Findings

Audit findings themselves may leak sensitive metadata.

Example:

```text
GitHub and AWS use the same password
```

Therefore persistent audit results should either:

* Be encrypted.
* Be recomputed on demand.

Recomputing on demand is preferred initially.

---

# 38. Password Age

Password age is an informational heuristic.

GoVault must not imply that changing passwords frequently is always beneficial.

Age warnings should be configurable and advisory.

---

# 39. TOTP Security

TOTP provides a second factor only when the OTP secret is stored separately from the password.

If both password and TOTP secret are stored in GoVault, compromise of the unlocked vault exposes both.

GoVault should document this tradeoff.

The TOTP feature provides convenience, not independent factor isolation.

---

# 40. Security Score

Any GoVault "security score" is a UX heuristic.

It must not be presented as:

* A formal risk assessment.
* A security certification.
* Proof that an account is secure.

---

# 41. Doctor Threat Model

`govault doctor` may inspect:

* Permissions.
* Database integrity.
* Schema version.
* Crypto format.
* Configuration.
* Clipboard availability.
* Network dependency policy.

It must not print secret values.

---

# 42. `doctor` While Locked

Where possible, diagnostics should function without decrypting vault content.

Checks requiring decryption should clearly indicate that unlock is required.

---

# 43. Crash Safety

Security includes availability.

GoVault must minimize the risk that a crash destroys the vault.

Operations requiring special crash safety include:

```text
vault initialization
schema migration
master password change
backup creation
restore
key rotation
```

---

# 44. Transaction Boundaries

SQLite transactions should be used to ensure logically related changes are atomic.

Example:

Updating an entry and creating its history record should occur consistently.

---

# 45. Migration Threat Model

Migrations may alter encrypted records.

A failed migration must not silently corrupt the only copy of a vault.

Potential strategy:

```text
validate
→ snapshot
→ migrate
→ verify
→ commit
```

---

# 46. Legacy Crypto

Once cryptographic formats evolve, GoVault may need to read legacy vaults.

Rules:

* Never silently downgrade newly written data.
* Legacy formats should be clearly versioned.
* Insecure legacy formats may require migration before normal use.
* Unsupported formats must fail safely.

---

# 47. Key Rotation

Future versions may support rotating the Vault Key.

Rotation should:

1. Generate a new random Vault Key.
2. Re-encrypt records safely.
3. Preserve old state until successful completion.
4. Verify all records.
5. Commit atomically where practical.

Key rotation is not required for MVP unless the cryptographic design requires it.

---

# 48. Master Password Recovery

GoVault does not know the master password.

Therefore:

> GoVault cannot recover a forgotten master password.

No hidden recovery key should exist by default.

Users should be warned during vault creation.

---

# 49. Recovery Codes for GoVault

A future optional recovery mechanism would materially change the threat model.

Examples:

```text
recovery key
emergency key
secondary unlock credential
```

Such a feature must not be added without a dedicated security design.

---

# 50. No Backdoors

GoVault must not contain:

* Developer master keys.
* Hidden recovery passwords.
* Remote unlocking mechanisms.
* Escrowed encryption keys.
* Telemetry-based recovery.

---

# 51. Random Vault Identifier

Each vault should have a random identifier used where appropriate for domain separation and authenticated context.

The identifier does not need to be secret.

---

# 52. Cross-Vault Record Substitution

An attacker should not be able to copy a valid encrypted record from Vault A into Vault B and have it decrypt successfully.

The cryptographic design should bind encrypted records to their vault identity.

---

# 53. Cross-Record Substitution

Similarly, ciphertext belonging to one record should not be valid when substituted into a different record identity.

Associated authenticated data should bind ciphertext to its intended context.

---

# 54. Backup/Vault Domain Separation

Cryptographic operations for:

```text
vault records
backup payloads
key wrapping
history records
```

should use explicit domain separation where appropriate.

The same raw key should not be casually reused across cryptographic purposes without a reviewed design.

---

# 55. Key Hierarchy Direction

The target conceptual design is:

```text
Master Password
       │
       ▼
    Argon2id
       │
       ▼
Key Encryption Key
       │
       ▼
Encrypted Vault Key
       │
       ▼
    Vault Key
```

Additional derived keys may be generated from the Vault Key for separate purposes.

Potential example:

```text
Vault Key
├── Entry Encryption Key
├── Backup Key
└── Future specialized keys
```

The exact hierarchy will be defined in `cryptography.md`.

---

# 56. Why Not Encrypt Directly With the Master Password?

Separating the password-derived key from the Vault Key provides important operational benefits.

Changing the master password requires rewrapping the Vault Key rather than re-encrypting every entry.

This also separates:

```text
password-based key derivation
```

from:

```text
record encryption
```

---

# 57. Master Password Verification

GoVault should avoid storing a direct password verifier unless necessary.

Successful decryption/authentication of wrapped key material may serve as password verification.

The exact design belongs in `cryptography.md`.

---

# 58. Offline Guarantee Threat Model

The offline guarantee is itself a security property.

GoVault production code must not require:

```text
network sockets
HTTP clients
DNS lookup
remote services
telemetry SDKs
analytics SDKs
cloud SDKs
```

---

# 59. Allowed OS Interactions

GoVault may interact with local operating-system facilities such as:

```text
filesystem
clipboard
terminal
secure randomness
environment/config directories
```

These do not violate the offline guarantee.

---

# 60. Future Features Requiring Threat Model Review

The following features must trigger a security review before implementation:

* Browser integration.
* Plugin systems.
* IPC daemon.
* SSH agent integration.
* Unix sockets.
* Shared vaults.
* Secret injection into child processes.
* Hardware security keys.
* OS keychain integration.
* Recovery keys.
* Automatic import.
* Plaintext export.
* External editors.
* Extension scripting.

---

# 61. External Editor Risk

Opening a secure note using:

```text
$EDITOR
```

could write plaintext to:

* Temporary files.
* Swap.
* Editor backups.
* Undo history.

Therefore external editor support should not be part of MVP without a deliberate security design.

---

# 62. Child Process Secret Injection

A future feature such as:

```bash
govault run -- command
```

could inject secrets into child processes.

This creates new risks involving:

* Environment leakage.
* Process inspection.
* Child process trust.
* Crash reporting.

It requires a separate threat model extension.

---

# 63. Plugin Systems

An in-process plugin system would give plugins access to decrypted secrets.

GoVault should not implement arbitrary third-party plugins without a strong capability/security model.

Plugins are explicitly out of scope for the initial design.

---

# 64. Security Assumptions

GoVault assumes:

1. The cryptographic primitives used are secure when used correctly.
2. `crypto/rand` provides secure randomness.
3. The Go runtime behaves according to its documented guarantees.
4. The user can keep a sufficiently strong master password.
5. The operating system provides basic filesystem isolation between users.
6. The GoVault executable has not been maliciously replaced.
7. The host is not fully compromised while secrets are being used.

---

# 65. Accepted Risks

The following risks are accepted for the initial GoVault design:

* A compromised unlocked host can expose secrets.
* Clipboard managers may retain copied values.
* Terminal scrollback may retain revealed secrets.
* Ciphertext length may leak approximate plaintext size.
* Database size may reveal approximate entry count.
* Rollback attacks are not fully prevented.
* Perfect memory zeroization cannot be guaranteed in Go.
* Root/Administrator can potentially inspect the unlocked process.
* Filesystem timestamps may reveal usage patterns.
* TOTP stored alongside passwords does not provide independent factor isolation.

---

# 66. Security Non-Claims

GoVault must never claim:

```text
military-grade encryption
unhackable
zero-risk
perfectly secure
memory-safe secret erasure
protection against a fully compromised OS
```

Security documentation should use precise language.

---

# 67. Security Invariants

The following are mandatory invariants.

## INV-001

The master password is never persisted.

## INV-002

The Vault Key is never persisted in plaintext.

## INV-003

Passwords are never persisted in plaintext.

## INV-004

TOTP seeds are never persisted in plaintext.

## INV-005

Secure notes are never persisted in plaintext.

## INV-006

Encrypted data uses authenticated encryption.

## INV-007

Cryptographic randomness uses `crypto/rand`.

## INV-008

Password-based derivation uses Argon2id.

## INV-009

Secrets are not written to logs.

## INV-010

Secrets are not printed by default.

## INV-011

Backups never contain plaintext vault secrets.

## INV-012

Backup authentication is verified before restore.

## INV-013

Restore failure must not destroy the current valid vault.

## INV-014

The application must not silently downgrade cryptographic formats.

## INV-015

Networking is not part of the GoVault production architecture.

---

# 68. Security Review Checklist

Any security-sensitive pull request should consider:

```text
[ ] Does this persist new data?

[ ] Is the data sensitive?

[ ] Should it be encrypted?

[ ] Can it appear in logs?

[ ] Can it appear in terminal output?

[ ] Can it appear in command arguments?

[ ] Can it appear in environment variables?

[ ] Does it create temporary files?

[ ] Does it add a dependency?

[ ] Does the dependency perform network activity?

[ ] Does it alter the vault format?

[ ] Does it alter cryptographic context?

[ ] Does it alter backup compatibility?

[ ] Does it require migration?

[ ] Does failure preserve existing data?

[ ] Can attacker-controlled input trigger large allocations?

[ ] Does it introduce a new secret lifetime?

[ ] Does it require updating this threat model?
```

---

# 69. Testing Requirements Derived From This Threat Model

GoVault security testing should include:

## Cryptographic tests

* Valid encryption/decryption.
* Wrong key.
* Modified ciphertext.
* Modified authentication tag.
* Modified associated data.
* Nonce behavior.

## Vault tests

* Wrong master password.
* Corrupted wrapped Vault Key.
* Invalid KDF parameters.
* Excessive KDF parameters.
* Unsupported crypto version.
* Cross-record substitution.
* Cross-vault substitution.

## Backup tests

* Corrupted header.
* Corrupted ciphertext.
* Truncated backup.
* Unsupported version.
* Invalid lengths.
* Modified authentication data.
* Failed restore.
* Interrupted restore.

## CLI tests

Confirm secrets are not accidentally printed.

## Logging tests

Confirm secrets never appear in logs.

## Storage tests

Confirm no plaintext known-secret marker exists inside:

```text
vault.db
WAL
journal
backup files
```

---

# 70. Known-Secret Storage Test

Automated integration tests should create entries containing recognizable markers such as:

```text
GOVAULT_TEST_PASSWORD_XYZ123
GOVAULT_TEST_TOTP_SECRET_ABC456
GOVAULT_TEST_SECURE_NOTE_DEF789
```

After database operations, tests should scan persistent artifacts and ensure these values do not occur in plaintext.

This should include:

```text
vault.db
vault.db-wal
vault.db-journal
encrypted backups
recovery snapshots
```

---

# 71. Network Isolation Test

CI should run GoVault tests in an environment without network connectivity where practical.

Additionally, static checks should detect prohibited networking dependencies.

---

# 72. Fuzzing Targets

Security-sensitive fuzz targets should include:

```text
backup parser
encrypted record parser
vault header parser
KDF parameter parser
migration input
import formats
configuration parser
```

---

# 73. Security Documentation Requirements

The following security documentation should remain synchronized:

```text
docs/threat-model.md
docs/cryptography.md
docs/vault-format.md
docs/backup-format.md
SECURITY.md
```

A cryptographic-format change should require reviewing all relevant documents.

---

# 74. Threat Model Change Process

Changes that materially alter the threat model should be explicitly documented.

Examples:

```text
adding network functionality
adding plugins
adding a daemon
adding browser integration
adding OS keychain support
adding cloud synchronization
adding plaintext export
adding recovery credentials
```

Such changes should not be merged as ordinary feature work without security review.

---

# 75. MVP Threat Model Summary

For GoVault v0.1, the system primarily protects against:

```text
stolen vault databases
stolen encrypted backups
offline password guessing
unauthorized ciphertext modification
accidental terminal disclosure
unsafe clipboard persistence
unsafe filesystem permissions
corruption during backup or restore
```

GoVault does not attempt to protect against:

```text
fully compromised operating systems
malicious kernels
root/Administrator while unlocked
keyloggers
malicious terminal emulators
physical attackers using an already unlocked vault
```

---

# 76. Security Model in One Sentence

> GoVault protects secrets at rest against offline compromise while minimizing accidental exposure during local terminal use, but it does not claim to protect secrets from an attacker who already controls the running machine.

---

# 77. Decisions Required Before `cryptography.md`

The following questions must be resolved next:

1. Which AEAD algorithm will GoVault use?
2. How will Vault Keys be generated?
3. How will keys be derived and separated?
4. What exact Argon2id parameters will be used?
5. How will KDF parameters evolve?
6. How will the Vault Key be wrapped?
7. Which metadata is encrypted?
8. Which metadata is authenticated as Associated Data?
9. What is the nonce generation strategy?
10. How are entries serialized before encryption?
11. How are history records encrypted?
12. How are backups encrypted?
13. Do backups reuse vault-derived keys or use independent credentials?
14. How is cross-vault ciphertext substitution prevented?
15. How is cross-record substitution prevented?
16. How are crypto format versions represented?
17. How does master password rotation work atomically?
18. What limits are placed on attacker-controlled KDF parameters?
19. How will migration to future cryptographic versions work?
20. What test vectors will define compatibility?

These questions are the input to the next document:

```text
docs/cryptography.md
```

