package cipher

import (
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	// KeySize specifies the required 256-bit (32-byte) key length.
	KeySize = 32
)

var (
	ErrInvalidKeySize    = errors.New("invalid key size: must be exactly 32 bytes")
	ErrDecryptionFailed  = errors.New("decryption failed: authentication tag mismatch or malformed payload")
)

// Encrypt encrypts plaintext using XChaCha20-Poly1305 with a freshly generated 192-bit CSPRNG nonce
// and returns the packed payload [24-byte nonce || ciphertext + 16-byte auth tag].
func Encrypt(key []byte, plaintext []byte, aad []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("%w: got %d bytes", ErrInvalidKeySize, len(key))
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize XChaCha20-Poly1305: %w", err)
	}

	nonce, err := GenerateNonce()
	if err != nil {
		return nil, err
	}

	// aead.Seal appends the ciphertext and 16-byte tag to the provided slice
	ciphertextWithTag := aead.Seal(nil, nonce, plaintext, aad)

	return PackPayload(nonce, ciphertextWithTag)
}

// Decrypt unpacks the payload, verifies the authentication tag against the provided AAD,
// and decrypts the ciphertext using XChaCha20-Poly1305.
func Decrypt(key []byte, packedPayload []byte, aad []byte) ([]byte, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("%w: got %d bytes", ErrInvalidKeySize, len(key))
	}

	nonce, ciphertextWithTag, err := UnpackPayload(packedPayload)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize XChaCha20-Poly1305: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertextWithTag, aad)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}
