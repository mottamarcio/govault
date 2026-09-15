package cipher

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const (
	// NonceSize specifies the required 24-byte (192-bit) nonce size for XChaCha20-Poly1305.
	NonceSize = 24
)

var (
	ErrInvalidNonceSize = errors.New("invalid nonce size: must be exactly 24 bytes")
)

// GenerateNonce generates a cryptographically secure 24-byte nonce using crypto/rand.
func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate random nonce: %w", err)
	}
	return nonce, nil
}

// ValidateNonce ensures the nonce length matches the required 24 bytes.
func ValidateNonce(nonce []byte) error {
	if len(nonce) != NonceSize {
		return fmt.Errorf("%w: got %d bytes", ErrInvalidNonceSize, len(nonce))
	}
	return nil
}
