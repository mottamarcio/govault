---
type: constitution
---
# Project Constitution: GoVault

## Product Invariants

- GoVault is strictly offline-by-design: it MUST NOT require, request, or initiate network connectivity under any circumstance (no telemetry, no analytics, no cloud synchronization, no remote update checks, no remote breach lookups, and no remote metadata/favicon fetching).
- Parity between interfaces: the CLI and TUI MUST share identical domain and application services; neither interface shall possess capabilities not expressible or available through the shared core.
- Secure by default: secrets, passwords, and sensitive keys MUST NOT be printed to stdout or logged in plaintext unless explicitly requested via dedicated CLI flags.
- Clipboard operations MUST automatically clear sensitive data after a configured expiration window (default 45 seconds).

## Architecture Invariants

- Inward dependency flow (Clean/Hexagonal Architecture): presentation layers (CLI and TUI) depend on Application Services, which depend on the pure Domain model. Presentation layers MUST NOT contain direct database queries or cryptographic business logic.
- Domain models MUST remain pure with zero third-party dependencies.
- Production binaries MUST exclude all networking packages (e.g., `net`, `net/http`); adherence MUST be enforced in CI build and lint verification.
- Cross-platform file path resolution MUST strictly adhere to OS standards (XDG specifications on Linux/Unix, standard Application Support on macOS, AppData on Windows).

## Security Invariants

- Application-level authenticated encryption is mandatory: persistence engines (SQLite) are treated as untrusted storage; sensitive data MUST be encrypted before persistence.
- Cryptographic Suite v1 baseline:
  - Password KDF: Argon2id (RFC 9106 recommended parameters: 3 iterations, 64 MB memory, 4 parallelism threads, 32-byte CSPRNG salt).
  - Key derivation and domain separation: HKDF-SHA-256.
  - Authenticated encryption: XChaCha20-Poly1305 using 256-bit keys and fresh 192-bit nonces generated from a cryptographically secure random number generator (`crypto/rand`).
- Nonce reuse prohibition: 192-bit CSPRNG nonces MUST be generated uniquely for every encryption operation.
- Context authentication: Associated Authenticated Data (AAD) MUST bind record metadata (Record ID, Version, Vault ID, Record Type) to ciphertext to prevent ciphertext relocation or transplant attacks.
- Master passwords and raw key material MUST NEVER be written to disk in plaintext, reversible format, or insecure temporary files.
- Secrets MUST NOT be accepted as command-line positional or flag arguments (to avoid exposure in process lists and shell history); they MUST be ingested via stdin, secure prompts, or environment variables.

## Data Invariants

- Local database transactions and write operations MUST use SQLite in Write-Ahead Logging (`WAL`) mode with foreign key constraints enabled and explicit busy timeout handling.
- Sensitive fields (title/name, credentials, secret notes, custom attributes, TOTP seeds, history payloads) MUST be stored encrypted within encrypted record envelopes; unencrypted SQLite storage is restricted strictly to non-secret routing metadata (UUIDs, timestamps, record types, schema versions).
- In-memory search: secret search indexes MUST only exist in transient memory while the vault is unlocked; plaintext secret search indices MUST NOT be persisted to SQLite.
- File system permissions on database files (`vault.db`, `vault.db-wal`, `vault.db-shm`), config files, and backup archives MUST be restricted to the owner only (`0600` / `0700` directory).

## Integration Invariants

- Portable backups (`.gvault`) MUST be self-contained logical representations independent of SQLite database files and host-specific paths.
- Backup payloads MUST be encrypted using XChaCha20-Poly1305 with headers cryptographically bound via AAD.
- Backup creation MUST be atomic (written to temporary files with restricted permissions and atomically renamed upon success) to prevent corrupted partial archives.

## Quality Requirements

- Code MUST follow SOLID, DRY, KISS, and YAGNI: no speculative
  abstraction, no duplicated logic, no unnecessary complexity ahead of
  a demonstrated need.
- Every change to behavior MUST be covered by tests. Prefer writing
  the test first (TDD) and specifying behavior through concrete
  scenarios before implementation (BDD).
- Cryptographic operations and security boundaries MUST be verified with unit and integration tests, including known-answer test vectors (KAT) and boundary condition tests.
- CI pipeline MUST enforce linting, test coverage, format verification, and offline-code checks (absence of network dependencies).

## Compatibility Requirements

- Format Versioning: Vault metadata schema and backup archives MUST include explicit format and crypto version numbers to ensure backward compatibility and safe migration paths.
- Master Password Rotation: Master password rotation MUST update the wrapped Vault Key (KEK re-encryption) without necessitating full re-encryption of existing record ciphertexts.
