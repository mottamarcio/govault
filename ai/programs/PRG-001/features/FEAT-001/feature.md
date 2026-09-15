---
id: FEAT-001
type: feature
status: active
parent: PRG-001
---

# FEAT-001: Core Cryptography Engine & Key Hierarchy

## Capability

Provides cryptographic primitives, password derivation (Argon2id), key derivation & domain separation (HKDF-SHA-256), authenticated envelope encryption (XChaCha20-Poly1305), CSPRNG generation, and key wrapping for secure vault and record protection.

## User Value

Guarantees that all user credentials, secret notes, and metadata are cryptographically protected at rest with modern, high-entropy cipher suites, preventing unauthorized data recovery even if the storage file is stolen.

## Scope

- Implementation of Argon2id KDF with RFC 9106 parameter configuration.
- Implementation of HKDF-SHA-256 for deriving KEK and MCK keys with domain separation.
- Implementation of XChaCha20-Poly1305 authenticated symmetric encryption with 192-bit CSPRNG nonces.
- Key hierarchy management (Master Password -> Master Key -> KEK/MCK -> Vault Key -> Record Keys).
- Master password verification without storing master passwords.
- Master password rotation by re-wrapping the Vault Key (KEK re-encryption).
- Associated Authenticated Data (AAD) generation binding context metadata.

## Non-Goals

- Unauthenticated or legacy ciphers (e.g. AES-CBC, DES).
- Hardware token integration (YubiKey/FIDO2) for v1.
- Multi-user asymmetric public key encryption.

## Constraints

- Strictly offline: relies solely on Go standard library (`crypto/rand`, `crypto/subtle`, `crypto/sha256`) and vetted Go cryptography packages (`golang.org/x/crypto/argon2`, `golang.org/x/crypto/chacha20poly1305`, `golang.org/x/crypto/hkdf`).
- Nonces must never repeat: fresh 192-bit nonce generated for every encryption operation.
- Constant-time comparison for all cryptographic tags and confirmation checks.

## Relevant Knowledge

- [KNOW-002](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-002-threat-model.md) — Security Threat Model and Trust Boundaries
- [KNOW-003](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-003-cryptography.md) — Cryptographic Architecture and Key Hierarchy

## Open Questions

- None.
