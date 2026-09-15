---
type: plan
for: SPEC-002
status: ready
---

# Implementation Plan: Authenticated Symmetric Encryption (XChaCha20-Poly1305) & AAD Binding

## Summary

Design and implement the authenticated symmetric encryption package (`internal/crypto/cipher`) utilizing XChaCha20-Poly1305 AEAD with fresh 192-bit CSPRNG nonces per operation, canonical Associated Authenticated Data (AAD) generation, and structured ciphertext envelope packing/unpacking.

## Repository Context

The repository now has Go module `github.com/mottamarcio/govault` and `internal/crypto/kdf` implemented and validated under `SPEC-001`. The cryptographic cipher package will reside at `internal/crypto/cipher`, depending on `golang.org/x/crypto/chacha20poly1305` and Go standard library `crypto/rand` / `crypto/subtle`.

## Requirement Coverage

- **R1 → Authenticated Cipher Implementation:**
  - Implement `NewAEAD(key []byte) (cipher.AEAD, error)` wrapping `chacha20poly1305.NewX(key)`.
  - Validate that key length is strictly 32 bytes (256-bit).
  - Encrypt with 16-byte Poly1305 authentication tag appended automatically.
- **R2 → Nonce Generation and Freshness:**
  - Generate a random 24-byte (192-bit) nonce using `crypto/rand.Reader` for every encryption invocation via `GenerateNonce() ([]byte, error)`.
  - Enforce that standard ChaCha20 96-bit nonces are rejected and only 192-bit nonces are used.
- **R3 → Associated Authenticated Data (AAD) Binding:**
  - Define `AADContext` struct capturing `VaultID`, `RecordID`, `RecordType`, and `Version`.
  - Implement canonical byte serialization `(c AADContext) Bytes() []byte` (e.g. `govault/v1|vault:<id>|record:<id>|type:<type>|v:<version>`) ensuring deterministic context authentication without field delimiters ambiguity.
- **R4 → Envelope Packing/Unpacking:**
  - Define `Envelope` struct: `Nonce []byte` (24 bytes), `Ciphertext []byte`, `AuthTag []byte` (16 bytes).
  - Provide binary packing helper `(e Envelope) MarshalBinary() []byte` (`[24 bytes Nonce || Ciphertext + Tag]`) and unpacker `UnmarshalBinary(data []byte) (*Envelope, error)`.
  - Provide high-level encrypt/decrypt functions: `Encrypt(key []byte, plaintext []byte, aad []byte) ([]byte, error)` and `Decrypt(key []byte, packed []byte, aad []byte) ([]byte, error)`.

## Architecture

- Pure Go package at `internal/crypto/cipher`.
- Implements authenticated encryption with zero network dependencies.
- Defensive error handling: generic decryption failure messages preventing cryptographic padding or oracle leaks.

## Components Affected

- `internal/crypto/cipher/cipher.go`: Primary AEAD wrapper, Encrypt and Decrypt functions.
- `internal/crypto/cipher/nonce.go`: Nonce generation and validation.
- `internal/crypto/cipher/aad.go`: Canonical AAD context structuring and serialization.
- `internal/crypto/cipher/envelope.go`: Structured envelope representations and byte encoding.
- `internal/crypto/cipher/*_test.go`: Unit tests, corruption/tampering tests, and KAT vectors.

## Data Changes

- Standard binary payload wire format:
  `[ 24-byte Nonce | (N-byte Ciphertext + 16-byte Poly1305 Tag) ]` (total minimum size = 40 bytes for 0-byte plaintext).

## API Changes

```go
package cipher

const (
    KeySize   = 32
    NonceSize = 24
    TagSize   = 16
    MinPayloadSize = NonceSize + TagSize // 40 bytes
)

type AADContext struct {
    VaultID    string
    RecordID   string
    RecordType string
    Version    uint32
}

func (c AADContext) Bytes() []byte
func GenerateNonce() ([]byte, error)
func Encrypt(key []byte, plaintext []byte, aad []byte) ([]byte, error)
func Decrypt(key []byte, packed []byte, aad []byte) ([]byte, error)
```

## Integration Changes

- Ingested by `SPEC-003` (Vault Key wrapping) and `FEAT-002` (SQLite record envelope encryption).

## Implementation Sequence

1. Implement `nonce.go` with 192-bit CSPRNG generation.
2. Implement `aad.go` with canonical AAD structuring and serialization.
3. Implement `cipher.go` with XChaCha20-Poly1305 `Encrypt` and `Decrypt`.
4. Implement `envelope.go` for structured envelope packing/unpacking.
5. Create comprehensive test suite in `cipher_test.go` verifying round-trips, bit-flip tamper detection, AAD mismatch rejection, and malformed payload handling.

## Test Strategy

- **Round-Trip Test:** Verify encryption and decryption on varying payload sizes (0 bytes, 16 bytes, 1 KB, 64 KB).
- **Tamper Resistance:** Modify individual bits in nonce, ciphertext, and tag; assert decryption returns authentication failure.
- **AAD Context Binding:** Encrypt with context A and verify decryption fails when supplied context B.
- **Boundary Conditions:** Verify rejection of truncated buffers (< 40 bytes) and invalid key lengths (!= 32 bytes).

## Risks

- None.

## Assumptions

- CSPRNG `crypto/rand` is available and reliable on all supported deployment OS platforms.
