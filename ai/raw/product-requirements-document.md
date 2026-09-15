# GoVault

## Product Requirements Document

**Product:** GoVault
**Category:** Offline Password & Secrets Manager
**Platform:** Terminal / CLI / TUI
**Primary Language:** Go
**Document Status:** Draft
**Target Initial Release:** v0.1.0
**License:** TBD

---

# 1. Executive Summary

GoVault is a local-first, terminal-native password and secrets manager written in Go.

It is designed for developers, system administrators, security-conscious users, and terminal enthusiasts who want a fast and ergonomic password manager without accounts, cloud services, telemetry, analytics, or network connectivity.

GoVault provides two complementary interfaces:

* A rich interactive Terminal User Interface (TUI).
* A scriptable Command Line Interface (CLI).

The application stores all vault information locally using SQLite. Sensitive data is encrypted before being persisted.

GoVault must operate entirely offline.

Network connectivity is intentionally outside the product's architecture.

The core product promise is:

> **Your secrets. Your machine. No network required.**

GoVault does not provide cloud synchronization, automatic remote backups, telemetry, analytics, breach APIs, favicon downloads, or any other network-dependent functionality.

Users may manually create encrypted backup files and store or transfer those files wherever they choose.

---

# 2. Product Vision

GoVault aims to combine the security characteristics expected from a password manager with the speed and composability expected from a Unix command-line tool.

The product should feel equally natural when used interactively:

```bash
govault
```

or through individual commands:

```bash
govault search github
govault copy github
govault generate --length 32
govault backup ~/backups/personal.gvault
```

The TUI should not merely wrap CLI commands.

Likewise, the CLI should not be an afterthought.

Both interfaces should operate on the same application and domain layers.

---

# 3. Product Principles

## 3.1 Offline by Design

GoVault must not require or attempt network connectivity.

The application must contain:

* No telemetry.
* No analytics.
* No cloud synchronization.
* No remote API calls.
* No automatic update checks.
* No remote password breach lookup.
* No remote favicon or metadata fetching.
* No account system.

The project should avoid networking dependencies in the core application.

Where practical, CI should verify that network packages such as `net/http` have not accidentally been introduced into production code.

---

## 3.2 Secure by Default

Users should not need to discover configuration options before GoVault behaves safely.

Examples:

* Passwords are hidden by default.
* Secrets are never printed to stdout unless explicitly requested.
* Clipboard contents are cleared automatically.
* Vaults automatically lock after inactivity.
* Database and configuration permissions are validated.
* Backup files are always encrypted.
* Destructive operations require explicit confirmation.

---

## 3.3 Keyboard First

Every important GoVault action must be accessible without a mouse.

The TUI should optimize for:

* Keyboard navigation.
* Fuzzy search.
* Shortcuts.
* Command palette.
* Minimal navigation depth.

---

## 3.4 Search First

Finding an item should be faster than navigating a hierarchy.

From almost anywhere in the TUI:

```text
/
```

opens vault search.

Similarly:

```text
Ctrl+K
```

opens the command palette.

---

## 3.5 Unix Friendly

GoVault should integrate naturally into terminal workflows.

Commands should be predictable, composable, and suitable for scripting where doing so does not unnecessarily expose secrets.

Example:

```bash
govault list --json
```

Sensitive fields must be excluded from normal structured output.

---

## 3.6 Recoverable Operations

Whenever reasonably possible, user mistakes should be reversible.

Examples:

* Entry history.
* Backup verification.
* Recovery snapshots before restore.
* Confirmation for destructive operations.

---

# 4. Target Users

## 4.1 Developers

Developers who spend significant time in the terminal and need to manage:

* Website credentials.
* API keys.
* Database credentials.
* SSH keys.
* Development secrets.
* TOTP credentials.
* Secure notes.

---

## 4.2 System Administrators

Users managing multiple systems, environments, servers, databases, and infrastructure credentials.

---

## 4.3 Security-Conscious Users

Users who prefer:

* Local storage.
* Minimal attack surface.
* No cloud account.
* No telemetry.
* Auditable open-source software.
* Explicit control over backups.

---

## 4.4 Terminal Enthusiasts

Users who prefer keyboard-driven workflows and terminal applications over desktop or browser interfaces.

---

# 5. Non-Goals

The following features are explicitly outside the initial product vision:

* Cloud synchronization.
* GoVault-hosted servers.
* Web interface.
* Browser extension.
* Mobile application.
* Online accounts.
* Remote password breach APIs.
* Automatic remote backup.
* Automatic software update checks.
* Team password sharing.
* Enterprise identity management.
* Browser autofill.

Some may eventually be evaluated as separate companion projects, but they must not compromise the offline guarantees of the core application.

---

# 6. Core Architecture

GoVault should use a layered architecture.

```text
CLI ───────────────┐
                   │
TUI ───────────────┤
                   ▼
             Application
                   │
                   ▼
                Domain
             ┌─────┼─────┐
             ▼     ▼     ▼
           Vault Crypto Storage
             │           │
             └─────┬─────┘
                   ▼
                 SQLite
```

Neither Cobra nor Bubble Tea should contain business logic.

Both should call application services.

---

# 7. Technology Stack

## Language

Go

Go is selected because of:

* Strong standard library.
* Simple deployment.
* Static binaries.
* Excellent concurrency model.
* Cross-platform support.
* Strong terminal tooling ecosystem.
* Good cryptographic ecosystem.

---

## CLI

**Cobra**

Responsibilities:

* Command parsing.
* Flags.
* Help output.
* Shell completion.
* CLI command hierarchy.

Example command tree:

```text
govault
├── init
├── list
├── search
├── show
├── add
├── edit
├── delete
├── copy
├── generate
├── totp
├── audit
├── backup
│   ├── create
│   ├── verify
│   └── inspect
├── restore
├── doctor
├── passwd
├── lock
└── version
```

---

## TUI

**Bubble Tea**

Responsible for:

* Application state.
* Event loop.
* Keyboard events.
* Views.
* Navigation.

Supporting Charm ecosystem libraries may be used where appropriate.

---

## Styling

**Lip Gloss**

Responsible for:

* Layout.
* Borders.
* Typography.
* Colors.
* Status indicators.
* Responsive terminal formatting.

The application must remain usable when terminal color support is limited.

---

## Persistence

**SQLite**

SQLite stores:

* Encrypted entries.
* Encrypted history.
* Vault metadata.
* Tags/index metadata where permitted by the security model.
* Configuration references.
* Schema versions.

Sensitive information must be encrypted before being written to SQLite.

SQLite itself must not be treated as an encryption boundary.

---

## Cryptography

**golang.org/x/crypto**

Primary cryptographic requirements include:

* Argon2id for password-based key derivation.
* Cryptographically secure randomness using `crypto/rand`.
* Authenticated encryption using a carefully selected standard AEAD construction.

The exact AEAD and serialized cryptographic format must be specified and reviewed before v1.

Custom cryptographic algorithms are prohibited.

---

# 8. Vault Cryptographic Model

GoVault should separate the user's master password from the key used to encrypt individual records.

Conceptually:

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
Decrypt encrypted Vault Key
       │
       ▼
    Vault Key
       │
       ├── encrypt Entry A
       ├── encrypt Entry B
       ├── encrypt History
       └── encrypt sensitive metadata
```

The master password must never be stored.

---

# 9. Master Password Changes

Changing the master password should not require decrypting and re-encrypting every vault item.

Instead:

1. Derive the old key-encryption key.
2. Decrypt the Vault Key.
3. Generate new KDF parameters as appropriate.
4. Derive a new key-encryption key.
5. Re-encrypt the Vault Key.
6. Atomically persist the new key material.

Command:

```bash
govault passwd
```

---

# 10. Vault Lifecycle

Possible states:

```text
UNINITIALIZED
LOCKED
UNLOCKED
```

The TUI must clearly communicate the current state.

When unlocked, sensitive key material exists only for as long as required by the application.

When locking:

* Vault key references must be discarded.
* Sensitive temporary buffers should be cleared where practical.
* Clipboard secrets should be cleared when possible.
* Sensitive TUI contents should disappear.
* The terminal should be redrawn.

---

# 11. Vault Entry Model

GoVault supports multiple entry types.

## Login

Fields:

```text
name
username
password
website
totp
tags
notes
custom fields
```

---

## Secure Note

```text
title
content
tags
```

---

## API Credential

```text
name
key
secret
environment
notes
custom fields
```

---

## Database Credential

```text
name
host
port
database
username
password
environment
notes
```

GoVault does not connect to the database.

These fields are informational secrets only.

---

## SSH Key

Potential later release:

```text
name
private key
public key
passphrase
comment
tags
```

---

## Custom Entry

Users may define arbitrary fields.

Each custom field should support a sensitivity flag:

```text
Field: Account Number
Sensitive: yes
```

Sensitive fields remain masked by default.

---

# 12. Entry Metadata

Entries should support:

* UUID.
* Name.
* Type.
* Tags.
* Favorite status.
* Created timestamp.
* Updated timestamp.
* History.
* Optional aliases.

Metadata encryption strategy must be defined by the threat model.

---

# 13. Aliases

Users can create short aliases for frequently accessed entries.

Example:

```bash
govault alias add gh github-personal
```

Then:

```bash
govault copy gh
```

Example aliases:

```text
gh       → GitHub Personal
ghwork   → GitHub Work
pgdev    → PostgreSQL Development
awsdev   → AWS Development
```

Aliases must not resolve ambiguously.

---

# 14. Search

Search is a core interaction.

Requirements:

* Fast fuzzy search.
* Search by name.
* Search by alias.
* Search by tag.
* Filter by entry type.
* Filter favorites.

Potential CLI:

```bash
govault search github
govault search --tag work
govault search --type login
govault search --favorites
```

Search indexes must respect the project's metadata threat model.

---

# 15. Password Generator

GoVault includes an offline cryptographically secure password generator.

Example:

```bash
govault generate
```

Options:

```bash
govault generate --length 32
govault generate --symbols
govault generate --no-ambiguous
```

Generator configuration:

* Length.
* Uppercase.
* Lowercase.
* Numbers.
* Symbols.
* Minimum numbers.
* Minimum symbols.
* Exclude ambiguous characters.
* Custom allowed characters.

Generation must use `crypto/rand`.

---

# 16. Passphrase Generator

GoVault should support passphrase generation.

Example:

```bash
govault generate --passphrase
```

Example output:

```text
meteor-cactus-library-raven-42
```

Configuration:

* Word count.
* Separator.
* Capitalization.
* Number inclusion.

The wordlist must be embedded in the application or distributed locally with GoVault.

No remote wordlist retrieval is permitted.

---

# 17. Password Strength

GoVault should provide a local strength estimate.

The UI may show:

```text
Strength  ████████████████████████████████████  Strong
Entropy   ~143 bits
```

The score must not claim more certainty than the underlying estimator provides.

---

# 18. Clipboard

GoVault should integrate with the system clipboard where supported.

Example:

```bash
govault copy github
```

Default:

```text
Password copied.
Clipboard will be cleared in 15 seconds.
```

Configuration:

```text
clipboard.timeout = 15s
```

Possible fields:

```bash
govault copy github --field password
govault copy github --field username
govault copy github --field totp
```

GoVault should avoid echoing copied secrets.

Clipboard behavior and limitations should be documented per operating system.

---

# 19. Auto-Lock

GoVault should automatically lock after inactivity.

Example configuration:

```text
vault.auto_lock = 5m
```

Possible values:

```text
1m
5m
10m
30m
never
```

`never` should require explicit user configuration.

---

# 20. TOTP

GoVault should provide offline TOTP generation.

No remote services are required.

Example:

```bash
govault totp github
```

TUI:

```text
GitHub Personal
482 193                    █████████████░░░   21s
```

TOTP secrets must be encrypted like passwords.

TOTP codes may optionally be masked until requested.

---

# 21. Security Audit

GoVault includes a completely offline vault audit.

Command:

```bash
govault audit
```

Checks may include:

* Weak passwords.
* Reused passwords.
* Old passwords.
* Empty passwords.
* Duplicate entries.
* Missing usernames where relevant.
* Entries with TOTP configured.

Example:

```text
Security Audit

142 credentials analyzed

✓ 119 strong and unique
⚠   7 reused passwords
⚠   4 weak passwords
⚠  12 old passwords

Security score: 82/100
```

The score is informational and should not be presented as a formal security guarantee.

---

# 22. Password Reuse Detection

Reuse detection should occur locally.

GoVault should avoid storing reusable plaintext-equivalent password fingerprints in the database.

Temporary comparison material should be held only while the audit runs and discarded afterward.

---

# 23. Entry History

GoVault should maintain encrypted history for significant entry changes.

Example:

```text
Sep 11, 2026
Password changed

Sep 10, 2026
TOTP added

Aug 03, 2026
Username changed

Jun 14, 2026
Entry created
```

Users may inspect previous versions and restore a previous version.

Retention should eventually be configurable.

---

# 24. Backup

GoVault provides manual encrypted backups.

It must never upload backups automatically.

Example:

```bash
govault backup create ~/backups/personal.gvault
```

Backup contents may include:

* Entries.
* History.
* TOTP secrets.
* Configuration necessary for restoration.
* Vault metadata.

The backup format should be versioned.

Example:

```text
GoVault Backup
Format: v1
```

---

# 25. Backup Verification

Users can verify a backup without restoring it.

```bash
govault backup verify ~/backups/personal.gvault
```

Checks:

* File structure.
* Format version.
* Authenticated encryption.
* Integrity.
* Required metadata.
* Truncation/corruption detection.

Example:

```text
✓ Valid GoVault backup
✓ Integrity verified
✓ Encryption authenticated
✓ Format supported
```

---

# 26. Backup Inspection

Where possible without compromising confidentiality:

```bash
govault backup inspect backup.gvault
```

After appropriate authentication, the application may display:

```text
Created        Sep 11, 2026 10:42
Format         GoVault Backup v1
Entries        196
Logins         142
Secure Notes    38
TOTP            12
SSH Keys         4
```

---

# 27. Restore

Example:

```bash
govault restore ~/backups/personal.gvault
```

Before restoration:

1. Validate backup.
2. Authenticate backup.
3. Verify integrity.
4. Display backup metadata.
5. Request confirmation.
6. Create a local recovery snapshot of the current vault.
7. Perform restore atomically where possible.

If restore fails, GoVault should preserve the existing vault.

---

# 28. GoVault Doctor

GoVault includes a local diagnostic system.

Command:

```bash
govault doctor
```

Potential checks:

```text
✓ Vault database
✓ Database permissions
✓ Config permissions
✓ Vault encryption format
✓ KDF parameters
✓ Random source
✓ Database integrity
✓ Clipboard integration
✓ Backup directory
✓ Schema version
✓ No unsupported migrations
✓ Network dependencies: none
```

`doctor` must never require network access.

---

# 29. CLI Security Rules

Secrets must not appear in normal output.

For example:

```bash
govault show github
```

should return:

```text
Name       GitHub Personal
Username   john@example.com
Password   ••••••••••••••
TOTP       ••• •••
Tags       personal, development
```

An explicit reveal mechanism may exist, but must be intentionally invoked and clearly documented as potentially unsafe because terminal output may be captured by shell history, logs, multiplexers, recordings, or other processes.

---

# 30. Structured Output

Non-secret commands may support:

```bash
govault list --json
```

Example:

```json
[
  {
    "id": "...",
    "name": "GitHub Personal",
    "type": "login",
    "tags": ["personal", "development"]
  }
]
```

Password, TOTP secret, API secret, private key, and other sensitive fields must not appear by default.

---

# 31. TUI Design Language

GoVault should have a recognizable but minimal visual identity.

Proposed mark:

```text
◆ GoVault
```

Core status indicators:

```text
🔒 LOCKED
🔓 UNLOCKED
◉ SECRET VISIBLE
```

The application should not depend on emoji rendering for usability; text equivalents must remain understandable.

---

# 32. TUI — Lock Screen

```text
╭──────────────────────────────────────────────────────────────────────────╮
│                                                                          │
│                              ◆                                           │
│                           GoVault                                        │
│                                                                          │
│                        VAULT LOCKED                                      │
│                                                                          │
│              ┌──────────────────────────────┐                            │
│              │ Master password              │                            │
│              │ ••••••••••••••••••          │                            │
│              └──────────────────────────────┘                            │
│                                                                          │
│                        [ Unlock ]                                        │
│                                                                          │
│              196 encrypted items · 100% offline                         │
│                                                                          │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

# 33. TUI — Dashboard

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault                                            🔓 VAULT UNLOCKED │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   VAULT                                                                  │
│                                                                          │
│   ┌────────────────┐  ┌────────────────┐  ┌────────────────┐             │
│   │      142       │  │       38       │  │       12       │             │
│   │     Logins     │  │  Secure Notes  │  │      TOTP      │             │
│   └────────────────┘  └────────────────┘  └────────────────┘             │
│                                                                          │
│   SECURITY                                                               │
│                                                                          │
│   ████████████████████████████████░░░░░░  82%                            │
│                                                                          │
│   ⚠  7 reused passwords                                                 │
│   ⚠  4 weak passwords                                                   │
│   ⚠  3 old passwords                                                    │
│                                                                          │
│   RECENT                                                                 │
│                                                                          │
│   GitHub Personal                                      2 minutes ago     │
│   PostgreSQL Dev                                       yesterday         │
│   AWS Development                                      3 days ago        │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  / Search    n New    g Generate    a Audit    b Backup    ? Help        │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

# 34. TUI — Quick Search

Search should be globally accessible using `/`.

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault                                                        🔓    │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Search vault                                                           │
│   ❯ git_                                                                 │
│                                                                          │
│   3 results                                                              │
│                                                                          │
│   ❯ ★ GitHub Personal                                                    │
│       john@example.com                     login · personal · dev        │
│                                                                          │
│     GitLab Work                                                          │
│       john@company.com                     login · work                  │
│                                                                          │
│     Gitea Local                                                          │
│       admin                                login · homelab               │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  ↑↓ navigate    enter open    ctrl+p copy password    esc close          │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

# 35. TUI — Credential Detail

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault  /  Logins  /  GitHub Personal                         🔓    │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   GitHub Personal                                              ★        │
│   ──────────────────────────────────────────────────────────────────     │
│                                                                          │
│   Username                                                               │
│   john@example.com                                             [copy]    │
│                                                                          │
│   Password                                                               │
│   ••••••••••••••••••••••••                                    [copy]    │
│                                                                          │
│   TOTP                                                                   │
│   ••• •••                                                     21s        │
│                                                                          │
│   Website                                                                │
│   github.com                                                             │
│                                                                          │
│   Tags                                                                   │
│   [personal] [development] [git]                                         │
│                                                                          │
│   Notes                                                                  │
│   Recovery codes stored in secure note "GitHub Recovery".                │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│ p copy password  y copy user  r reveal  t TOTP  e edit  h history  esc  │
╰──────────────────────────────────────────────────────────────────────────╯
```

Copying a secret must not automatically reveal it.

---

# 36. TUI — New Item Type

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault  /  New Item                                                  │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   What would you like to store?                                          │
│                                                                          │
│   ❯ 🔑 Login                                                             │
│                                                                          │
│     📝 Secure Note                                                       │
│                                                                          │
│     🔧 API Credential                                                    │
│                                                                          │
│     🗄  Database                                                         │
│                                                                          │
│     🔐 SSH Key                                                           │
│                                                                          │
│     ◇ Custom                                                             │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  ↑↓ navigate                  enter select                  esc cancel    │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

# 37. TUI — New Login

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault  /  New Login                                                 │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Name                                                                   │
│   ┌──────────────────────────────────────────────────────────────────┐   │
│   │ GitHub Personal                                                  │   │
│   └──────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│   Username                                                               │
│   ┌──────────────────────────────────────────────────────────────────┐   │
│   │ john@example.com                                                 │   │
│   └──────────────────────────────────────────────────────────────────┘   │
│                                                                          │
│   Password                                                               │
│   ┌──────────────────────────────────────────────────────────────────┐   │
│   │ ••••••••••••••••••••                                             │   │
│   └──────────────────────────────────────────────────────────────────┘   │
│   Strength  ██████████████████████████████████████  Strong               │
│                                                                          │
│   [ Generate password ]                                                  │
│                                                                          │
│   Website   github.com                                                   │
│   Tags      personal, development                                        │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  tab next field       ctrl+s save       ctrl+g generate       esc cancel │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

# 38. TUI — Password Generator

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault  /  Password Generator                                        │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   PASSWORD                                                               │
│                                                                          │
│   vK7#xL9@pQ2!mN8$zR4&wT6*                                               │
│                                                                          │
│   Strength  ████████████████████████████████████████████  Excellent      │
│   Entropy   ~154 bits                                                    │
│                                                                          │
│   OPTIONS                                                                │
│                                                                          │
│   Length                 24        ◀━━━━━━━━━━━━●━━━━━━━━▶                │
│                                                                          │
│   [✓] Uppercase          ABC                                             │
│   [✓] Lowercase          abc                                             │
│   [✓] Numbers            123                                             │
│   [✓] Symbols            !@#                                             │
│   [✓] Avoid ambiguous    0OIl                                            │
│                                                                          │
│   Minimum numbers        2                                                │
│   Minimum symbols        2                                                │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  r regenerate        c copy        s save as new login        esc close  │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

# 39. TUI — TOTP Authenticator

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault  /  Authenticator                                       🔓    │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Search  ❯ _                                                            │
│                                                                          │
│   GitHub Personal                                                        │
│   482 193                                      █████████████░░░   21s    │
│                                                                          │
│   AWS Development                                                        │
│   730 814                                      ████████░░░░░░░    14s    │
│                                                                          │
│   GitLab Work                                                            │
│   193 025                                      ████░░░░░░░░░░░     7s    │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  ↑↓ navigate       enter copy       / search       r reveal       esc    │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

# 40. TUI — Security Audit

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault  /  Security Audit                                            │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   VAULT HEALTH                                                           │
│                                                                          │
│   82 / 100     █████████████████████████████████░░░░░░░                  │
│                                                                          │
│   ✓ 119 strong & unique                                                  │
│   ⚠   7 reused                                                           │
│   ⚠   4 weak                                                             │
│   ⚠  12 old                                                              │
│                                                                          │
│   REUSED PASSWORDS                                                       │
│                                                                          │
│   ❯ GitHub Personal      ─┐                                               │
│     GitLab Personal      ─┼─ same password                               │
│     Gitea                ─┘                                               │
│                                                                          │
│     Netflix              ─┐                                               │
│     Spotify              ─┘ same password                                │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  1 weak    2 reused    3 old    enter inspect    e edit    esc back      │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

# 41. TUI — Backup

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault  /  Backup                                                    │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Create encrypted backup                                                │
│                                                                          │
│   Destination                                                            │
│   /home/user/backups/govault-2026-09-11.gvault                           │
│                                                                          │
│   Backup contents                                                        │
│                                                                          │
│       142  logins                                                        │
│        38  secure notes                                                  │
│        12  TOTP secrets                                                  │
│         4  SSH keys                                                      │
│                                                                          │
│   Encryption       Authenticated                                         │
│   Format           GoVault Backup v1                                     │
│                                                                          │
│                    [ Create Backup ]                                      │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  tab navigate                                      esc cancel             │
╰──────────────────────────────────────────────────────────────────────────╯
```

The UI should derive the displayed encryption algorithm from the actual vault format rather than hard-code marketing text.

---

# 42. TUI — Restore

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault  /  Restore                                                   │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Backup                                                                 │
│   ~/backups/govault-2026-08-20.gvault                                    │
│                                                                          │
│   ✓ Valid GoVault backup                                                 │
│   ✓ Integrity verified                                                   │
│   ✓ Encryption authenticated                                             │
│                                                                          │
│   Created       Aug 20, 2026 19:42                                       │
│   Entries       183                                                      │
│   Format        v1                                                       │
│                                                                          │
│   ⚠ Restoring will replace your current vault.                           │
│                                                                          │
│   A recovery snapshot will be created first.                             │
│                                                                          │
│                    [ Restore Backup ]                                     │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  enter restore                                              esc cancel   │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

# 43. TUI — Doctor

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault  /  Doctor                                                    │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Checking GoVault...                                                     │
│                                                                          │
│   ✓ Vault database             healthy                                   │
│   ✓ Database permissions       0600                                      │
│   ✓ Config permissions         0600                                      │
│   ✓ Vault encryption           v1                                        │
│   ✓ KDF                        Argon2id                                   │
│   ✓ Random source              available                                 │
│   ✓ Database integrity         OK                                        │
│   ✓ Clipboard                  available                                 │
│   ✓ Network dependencies       none                                      │
│                                                                          │
│   Everything looks good.                                                 │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  r run again                                           esc back           │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

# 44. TUI — Entry History

```text
╭──────────────────────────────────────────────────────────────────────────╮
│  ◆ GoVault  /  GitHub Personal  /  History                               │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│   Sep 11, 2026  09:42                                                    │
│   ● Password changed                                                     │
│   │                                                                      │
│   │ Sep 11, 2026  09:40                                                  │
│   ● TOTP added                                                           │
│   │                                                                      │
│   │ Aug 03, 2026  17:21                                                  │
│   ● Username changed                                                     │
│   │                                                                      │
│   │ Jun 14, 2026  11:03                                                  │
│   ● Entry created                                                        │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  ↑↓ navigate       enter inspect version       r restore       esc back  │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

# 45. TUI — Command Palette

Accessible using:

```text
Ctrl+K
```

Example:

```text
╭──────────────────────────────────────────────────────────────────────────╮
│                                                                          │
│   > backup_                                                              │
│                                                                          │
│   ❯ Create encrypted backup                         Backup               │
│     Restore backup                                  Backup               │
│     Verify backup                                   Backup               │
│                                                                          │
│     ────────────────────────────────────────────────────────────────     │
│                                                                          │
│     ↑↓ navigate   ↵ execute   esc close                                  │
│                                                                          │
╰──────────────────────────────────────────────────────────────────────────╯
```

Searchable actions include:

```text
New Login
New Secure Note
Generate Password
Generate Passphrase
Search Vault
Security Audit
Create Backup
Verify Backup
Restore Backup
Lock Vault
Change Master Password
Settings
Doctor
Quit
```

---

# 46. Keyboard Model

Global shortcuts:

```text
/             Search
Ctrl+K        Command palette
n             New entry
g             Generator
a             Security audit
b             Backup
l             Lock
?             Help
Esc           Back/close
q             Quit where appropriate
```

Credential shortcuts:

```text
p             Copy password
y             Copy username
t             Copy TOTP
r             Reveal
e             Edit
h             History
d             Delete
```

Destructive shortcuts must never execute immediately.

---

# 47. Secret Reveal Behavior

Secrets are masked by default.

Example:

```text
Password
••••••••••••••••••••
```

Pressing:

```text
r
```

reveals the selected secret temporarily.

The UI should display:

```text
◉ SECRET VISIBLE
```

Secrets should automatically become masked again after a configurable timeout or navigation event.

---

# 48. Settings

Potential configuration:

```text
vault.auto_lock
clipboard.timeout
secret.reveal_timeout
totp.mask_codes
generator.default_length
generator.symbols
generator.avoid_ambiguous
ui.startup_screen
ui.theme
history.enabled
history.retention
```

Configuration must not contain secret material.

---

# 49. Configuration File

Potential location:

```text
~/.config/govault/config.toml
```

Example:

```toml
[vault]
auto_lock = "5m"

[clipboard]
timeout = "15s"

[secret]
reveal_timeout = "10s"

[totp]
mask_codes = true

[generator]
default_length = 24
symbols = true
avoid_ambiguous = true

[ui]
startup_screen = "search"
```

OS-specific configuration directories should follow platform conventions.

---

# 50. Local Data Layout

Conceptual layout:

```text
~/.local/share/govault/
├── vault.db
└── state/

~/.config/govault/
└── config.toml
```

Exact paths should use operating-system conventions rather than assuming Unix paths everywhere.

---

# 51. Proposed Go Project Structure

```text
govault/
│
├── cmd/
│   ├── root.go
│   ├── init.go
│   ├── add.go
│   ├── edit.go
│   ├── delete.go
│   ├── list.go
│   ├── search.go
│   ├── copy.go
│   ├── generate.go
│   ├── totp.go
│   ├── audit.go
│   ├── backup.go
│   ├── restore.go
│   ├── doctor.go
│   └── passwd.go
│
├── internal/
│   ├── app/
│   ├── domain/
│   ├── vault/
│   ├── crypto/
│   ├── storage/
│   │   └── sqlite/
│   ├── backup/
│   ├── generator/
│   ├── clipboard/
│   ├── totp/
│   ├── audit/
│   ├── config/
│   └── tui/
│       ├── model/
│       ├── screens/
│       ├── components/
│       ├── keymap/
│       └── styles/
│
├── migrations/
├── docs/
├── testdata/
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

# 52. Domain Layer

The domain layer should know nothing about:

* Cobra.
* Bubble Tea.
* Lip Gloss.
* SQLite implementation details.
* Terminal rendering.

Potential entities:

```text
Vault
Entry
EntryVersion
Tag
Alias
TOTPSecret
BackupManifest
AuditFinding
```

---

# 53. Application Services

Potential application services:

```text
VaultService
EntryService
SearchService
GeneratorService
ClipboardService
TOTPService
AuditService
BackupService
RestoreService
DoctorService
HistoryService
```

Both CLI and TUI should consume these services.

---

# 54. SQLite Schema Direction

Conceptually:

```text
vault_metadata

entries
    id
    type
    encrypted_payload
    nonce
    created_at
    updated_at

entry_history
    id
    entry_id
    encrypted_payload
    nonce
    created_at

aliases
    alias
    entry_id

schema_migrations
```

This is not the final schema.

Whether names, types, tags, aliases, timestamps, or other metadata remain plaintext must be decided during threat modeling.

---

# 55. Threat Model

A formal threat model is required before stabilizing the storage format.

GoVault should consider at minimum:

### Attacker obtains `vault.db`

Goal:

Prevent recovery of secrets without the master password.

### Attacker obtains an encrypted backup

Goal:

Same confidentiality expectations as the primary vault.

### Attacker reads configuration

Goal:

Configuration reveals no credentials.

### Terminal shoulder surfing

Mitigations:

* Secrets masked by default.
* Temporary reveal.
* Clear sensitive screens on lock.

### Clipboard monitoring

Mitigations:

* Clipboard timeout.
* Avoid clipboard unless explicitly requested.
* Document OS limitations.

### Terminal logging

Mitigations:

* Do not print secrets by default.
* Explicit warnings for reveal-to-stdout operations.

### Memory inspection

Go cannot guarantee perfect secret erasure because of runtime and memory-management behavior.

GoVault must document this limitation rather than promise impossible memory-security guarantees.

### Compromised host

If an attacker controls the machine while the vault is unlocked, GoVault cannot guarantee protection from keyloggers, process inspection, clipboard interception, or malicious kernel-level software.

This limitation must be documented.

---

# 56. Security Invariants

The following should become explicit engineering invariants:

1. Master passwords are never persisted.
2. Plaintext vault secrets are never persisted.
3. Backups are never plaintext.
4. Secret fields are never logged.
5. Secret fields are excluded from normal JSON output.
6. Random cryptographic values use `crypto/rand`.
7. Password-derived keys use Argon2id.
8. Encryption must provide authentication.
9. Backup restore requires integrity verification.
10. Network access is not part of the application architecture.

---

# 57. Logging

Logging must never include:

* Passwords.
* Master passwords.
* Vault keys.
* TOTP seeds.
* API secrets.
* Private keys.
* Sensitive custom fields.

Debug logging must follow the same rule.

---

# 58. Database Migrations

Database migrations must be:

* Versioned.
* Deterministic.
* Tested.
* Transactional where possible.
* Backward-aware.

Before destructive migrations, GoVault should create a local recovery snapshot where practical.

---

# 59. Backup Format Versioning

Backup files should include an authenticated format header/manifest conceptually containing:

```text
magic
format_version
created_at
crypto_version
kdf_parameters
encrypted_payload
authentication_data
```

The exact serialization format should be documented.

GoVault must reject unsupported or malformed backup versions safely.

---

# 60. Atomic File Operations

Critical writes should follow safe patterns such as:

```text
write temporary file
→ flush
→ validate
→ rename atomically where supported
```

This is particularly important for:

* Backup creation.
* Restore.
* Vault initialization.
* Key material updates.
* Master password changes.

---

# 61. Error Handling

Errors should be actionable.

Bad:

```text
Error: operation failed
```

Better:

```text
Unable to create backup.

Destination:
~/backups/personal.gvault

Reason:
Permission denied.

The vault was not modified.
```

Security-sensitive errors should avoid revealing unnecessary cryptographic details.

---

# 62. Accessibility and Terminal Compatibility

GoVault should support:

* Narrow terminals.
* Large terminals.
* No-color environments.
* Common terminal emulators.
* Keyboard-only operation.

Important states must never be represented by color alone.

For example:

```text
✓ Healthy
⚠ Warning
✗ Error
```

rather than relying exclusively on green/yellow/red.

---

# 63. Performance Requirements

Target behavior for normal personal vaults:

Startup:

```text
< 100 ms target before authentication UI
```

Search:

```text
< 50 ms perceived response for typical vault sizes
```

The application should comfortably handle thousands of entries.

Cryptographic operations intentionally affected by Argon2id may exceed these latency targets.

---

# 64. Testing Strategy

Testing should include:

### Unit Tests

* Encryption/decryption.
* Key derivation.
* Password generation.
* Passphrase generation.
* TOTP.
* Search.
* Audit rules.
* Backup parsing.
* Configuration.

### Integration Tests

* Vault creation.
* Unlock.
* CRUD.
* Backup.
* Restore.
* Master password change.
* Database migrations.

### Property/Fuzz Testing

Strong candidates:

* Backup parser.
* Vault format parser.
* Import parser.
* Crypto serialization.
* Search parser.
* Custom fields.

### TUI Tests

Bubble Tea state transitions should be tested independently from terminal rendering where possible.

---

# 65. Corruption Testing

Tests should intentionally:

* Truncate vault databases.
* Truncate backup files.
* Modify encrypted bytes.
* Modify authentication tags.
* Use unsupported format versions.
* Provide malformed metadata.
* Simulate interrupted writes.

GoVault should fail safely.

---

# 66. Offline Guarantee Testing

CI should contain a check that detects unintended networking dependencies.

Potential policies:

* Detect imports of `net/http` in application packages.
* Detect known network client dependencies.
* Document allowed exceptions if Go internals indirectly reference networking packages.
* Run core integration tests in an isolated environment.

The guarantee should be enforced through architecture rather than only documentation.

---

# 67. Initial CLI Experience

First execution:

```bash
govault
```

If no vault exists:

```text
◆ GoVault

No vault found.

Create a new local vault?

> Create Vault
  Exit
```

The setup flow asks for:

```text
Master password
Confirm master password
Auto-lock timeout
Clipboard timeout
```

Then:

```text
✓ Vault created.

Your vault is stored locally.

GoVault cannot recover your master password.

Create an encrypted backup after adding your first credentials.
```

---

# 68. MVP — v0.1

The first release should deliberately remain focused.

Required:

* Vault initialization.
* Master password.
* Argon2id.
* Authenticated encryption.
* SQLite persistence.
* Login entries.
* Secure notes.
* CRUD.
* Fuzzy search.
* Tags.
* Password generator.
* Clipboard integration.
* Clipboard auto-clear.
* Auto-lock.
* Cobra CLI.
* Bubble Tea TUI.
* Lip Gloss UI.
* Manual encrypted backup.
* Backup verification.
* Restore.
* `doctor`.
* No network dependencies.

---

# 69. v0.2

Potential scope:

* Custom fields.
* Favorites.
* Entry history.
* Passphrase generator.
* Aliases.
* Improved backup inspection.
* Security audit.
* Improved search filters.

---

# 70. v0.3

Potential scope:

* TOTP.
* API credential type.
* Database credential type.
* Command palette.
* Advanced audit.
* Import/export framework.
* More configurable keyboard shortcuts.

---

# 71. v0.4

Potential scope:

* SSH keys.
* Advanced history.
* Custom entry templates.
* Themes.
* Advanced backup management.
* Additional local import formats.

---

# 72. v1.0 Criteria

GoVault should not reach v1.0 merely because it has many features.

v1.0 should indicate confidence in:

* Cryptographic format.
* Vault format.
* Backup format.
* Migration strategy.
* Threat model.
* Recovery behavior.
* CLI compatibility.
* TUI stability.
* Cross-platform behavior.
* Documentation.
* Security review.

Once v1 is released, vault-format compatibility becomes a serious long-term commitment.

---

# 73. Success Metrics

Because GoVault has no telemetry, product metrics must not depend on user tracking.

Project health can instead be evaluated using voluntarily public information such as:

* GitHub stars.
* Contributors.
* Issues.
* Pull requests.
* Releases.
* Community feedback.
* Security reports.
* Package downloads where externally available.

No metrics should be collected from the GoVault application itself.

---

# 74. Open-Source Security

The project should provide:

```text
SECURITY.md
```

containing:

* Supported versions.
* Vulnerability reporting process.
* Disclosure expectations.
* Security scope.

Security issues should not require public disclosure before maintainers have an opportunity to investigate.

---

# 75. Documentation

Initial documentation should include:

```text
README.md
SECURITY.md
CONTRIBUTING.md
docs/
    architecture.md
    threat-model.md
    cryptography.md
    vault-format.md
    backup-format.md
    cli.md
    tui.md
```

The cryptographic design should be documented rather than hidden behind implementation details.

---

# 76. README Positioning

Suggested opening:

```text
◆ GoVault

A local-first password and secrets manager for your terminal.

100% offline.
No accounts.
No telemetry.
No cloud.
```

Followed by:

```bash
govault
```

for interactive usage and:

```bash
govault search github
govault copy github
govault generate --length 32
```

for CLI workflows.

---

# 77. Product Personality

GoVault should feel:

* Quiet.
* Fast.
* Predictable.
* Technical.
* Trustworthy.
* Minimal.

It should avoid excessive animations, decorative output, or unnecessary confirmations for safe operations.

Security-sensitive actions, however, should be unmistakable.

---

# 78. UX Philosophy

Two global interactions should define the TUI:

```text
/        Find anything in the vault.
Ctrl+K   Do anything in GoVault.
```

This minimizes deep menu navigation.

The ideal experienced-user workflow is:

```text
unlock
→ search
→ copy
→ continue working
```

often within seconds.

---

# 79. Critical Pre-Implementation Decisions

Before implementing the persistent vault format, the project must define:

1. Threat model.
2. Vault encryption format.
3. AEAD algorithm.
4. Key hierarchy.
5. Argon2id parameters and migration strategy.
6. Metadata encryption policy.
7. Nonce strategy.
8. Backup cryptographic format.
9. Key rotation behavior.
10. Master password change behavior.
11. Database schema.
12. Crash-safe write behavior.
13. Secret handling boundaries.
14. Clipboard security behavior.
15. Cross-platform filesystem permissions.

These decisions are more expensive to change than the CLI or TUI.

---

# 80. Recommended Implementation Sequence

### Phase 1 — Security Foundation

Implement:

```text
crypto
vault format
key hierarchy
Argon2id
vault initialization
unlock/lock
```

No TUI required yet.

### Phase 2 — Storage

Implement:

```text
SQLite
migrations
encrypted entries
CRUD
history foundation
```

### Phase 3 — Application Layer

Implement:

```text
VaultService
EntryService
SearchService
GeneratorService
```

### Phase 4 — CLI

Implement:

```text
govault init
govault add
govault list
govault search
govault copy
govault generate
```

### Phase 5 — Backup & Recovery

Implement:

```text
backup
verify
restore
recovery snapshots
doctor
```

### Phase 6 — TUI

Implement:

```text
lock screen
search
entry details
editor
generator
backup
doctor
```

### Phase 7 — Security Features

Implement:

```text
audit
history
aliases
TOTP
```

### Phase 8 — Hardening

Perform:

```text
fuzzing
corruption tests
migration tests
cross-platform tests
security review
documentation review
```

---

# 81. Definition of Done for v0.1

GoVault v0.1 is considered complete when a user can:

```text
install GoVault
↓
create a vault
↓
unlock it
↓
create credentials
↓
search credentials
↓
copy secrets safely
↓
generate passwords
↓
edit/delete entries
↓
create an encrypted backup
↓
verify that backup
↓
restore that backup
↓
run diagnostics
↓
lock the vault
```

without GoVault requiring or initiating any network communication.

---

# 82. Final Product Statement

GoVault is not intended to compete by integrating with every service.

Its differentiation is the opposite.

GoVault deliberately does less outside the user's machine.

It provides a focused environment for storing, retrieving, generating, auditing, and backing up secrets using a terminal-native workflow.

The core promise should remain understandable in one sentence:

> **GoVault is a fast, encrypted, offline password and secrets manager built for the terminal.**
