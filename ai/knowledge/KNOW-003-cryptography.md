---
id: KNOW-003
type: knowledge
status: active
sources:
  - path: ai/raw/cryptography.md
    fingerprint: sha256:0c2c2ffc31701d113668a656126c02a3c6b4c8810e42f5440b8c4ff48d0addcf
---

# KNOW-003: Cryptographic Architecture and Key Hierarchy

## Summary

Specifies the cryptographic algorithms, parameters, key derivation procedures, envelope encryption scheme, and key hierarchy used throughout GoVault (Crypto Suite v1).

## Known Facts

- Cryptographic Primitives (Crypto Suite v1):
  - Password KDF: Argon2id (RFC 9106 recommended parameters: Time = 3 iterations, Memory = 64 MB / 65536 KiB, Parallelism = 4 threads, Salt = 32 bytes random).
  - Key Derivation & Domain Separation: HKDF-SHA-256 (`crypto/hkdf`).
  - Authenticated Encryption: XChaCha20-Poly1305 (256-bit key, 192-bit / 24-byte extended random nonce, 128-bit / 16-byte authentication tag).
  - Secure Randomness: `crypto/rand`.
- Key Hierarchy:
  - Master Password -> Master Key (derived via Argon2id).
  - Key Encryption Key (KEK) & Master Confirmation Key (MCK) derived from Master Key using HKDF with distinct domain separation info strings.
  - Vault Key (VK): 256-bit high-entropy random key generated upon vault creation.
  - Wrapped Vault Key: Vault Key encrypted with KEK using XChaCha20-Poly1305.
  - Record Keys: Derived per-record or generated per-record and bound using authenticated associated data (AAD).
- Password Rotation:
  - Changing the master password re-derives a new Master Key and KEK, and re-encrypts the Wrapped Vault Key without needing to re-encrypt all individual vault secret records.

## Constraints

- Nonce reuse must be mathematically improbable: 192-bit nonces (XChaCha20-Poly1305) generated from CSPRNG (`crypto/rand`) for every encryption operation.
- Associated Data (AAD) must strictly authenticate record context (record ID, version, record type, vault ID) to prevent ciphertext transplant attacks between records or vaults.
- No roll-your-own cryptography or unauthenticated encryption modes (e.g. no AES-CBC without HMAC, no ECB).

## Unknowns

- Future support for hardware security keys (FIDO2/U2F or PKCS#11/YubiKey) or secondary KDFs if Argon2id parameters need adjustment on low-memory embedded environments.

## Conflicts

- None.

## Provenance

- Primitives, key derivation flows, parameters, and encryption rules derived directly from `ai/raw/cryptography.md`.

## Related Topics

- [KNOW-002](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-002-threat-model.md) — Security Threat Model
- [KNOW-004](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-004-vault-format.md) — Vault Storage Format
- [KNOW-005](file:///home/marciovcm/workspace/golang/govault/ai/knowledge/KNOW-005-backup-format.md) — Backup Format
