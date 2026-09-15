---
id: SPEC-002
type: spec
status: ready
parent: FEAT-001
depends_on:
  - SPEC-001
---

# SPEC-002: Authenticated Symmetric Encryption (XChaCha20-Poly1305) & AAD Binding

## Intent

Specify the authenticated encryption and decryption engine using XChaCha20-Poly1305 with random 192-bit CSPRNG nonces and strict Associated Authenticated Data (AAD) context binding.

## Requirements

- **R1: Authenticated Cipher Implementation:** All payload and record encryption must use XChaCha20-Poly1305 (`golang.org/x/crypto/chacha20poly1305.NewX`) with 256-bit keys and 16-byte Poly1305 authentication tags.
- **R2: Nonce Generation and Freshness:** A unique 24-byte (192-bit) nonce must be generated using `crypto/rand` for every single encryption call. Nonce reuse must be prevented.
- **R3: Associated Authenticated Data (AAD) Binding:** Every record encryption operation must include canonical AAD constructed from contextual metadata (e.g. `vault_id`, `record_id`, `record_type`, `version`). The decryption engine must reject any payload where AAD does not match.
- **R4: Envelope Packing/Unpacking:** The cryptographic layer must support packing/unpacking the envelope components (nonce, ciphertext, authentication tag) into standard byte slices or structured envelopes for database and archive storage.

## Acceptance Scenarios

- **Scenario 1: Successful Encrypt and Decrypt Round-Trip**
  - *Given* a 32-byte key, a plaintext payload, and context AAD
  - *When* the payload is encrypted and then decrypted with the same key and AAD
  - *Then* the decrypted plaintext exactly matches the original input.
- **Scenario 2: Tamper Detection (Integrity Failure)**
  - *Given* a valid ciphertext envelope
  - *When* a single bit of the ciphertext or authentication tag is modified
  - *Then* decryption fails with an explicit integrity error and returns no plaintext.
- **Scenario 3: AAD Mismatch / Relocation Attack Prevention**
  - *Given* a valid ciphertext created with AAD for Record `A`
  - *When* attempting to decrypt the ciphertext while supplying AAD for Record `B`
  - *Then* decryption fails authentication.

## Edge Cases

- Encrypting empty (0-byte) payloads must be supported and authenticate properly.
- Handling malformed ciphertext envelopes where total length is less than the required nonce + overhead length (24 + 16 = 40 bytes) must return an explicit error without panicking.

## Constraints

- Never use unauthenticated ciphers or short nonces (e.g., standard ChaCha20 96-bit nonces are forbidden; XChaCha20 192-bit nonces are required).
- Nonce must be generated directly from `crypto/rand.Reader`.

## Non-Goals

- Key exchange protocols or asymmetric cryptography.
