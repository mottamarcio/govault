package cipher

import (
	"errors"
	"fmt"
)

const (
	TagSize        = 16
	MinPayloadSize = NonceSize + TagSize // 40 bytes minimum (empty plaintext)
)

var (
	ErrPayloadTooShort = errors.New("payload too short to contain nonce and authentication tag")
)

// Envelope represents an unpacked encrypted ciphertext envelope.
type Envelope struct {
	Nonce      []byte
	Ciphertext []byte
	AuthTag    []byte
}

// PackPayload serializes the nonce and ciphertext+tag into a contiguous byte slice.
func PackPayload(nonce, ciphertextWithTag []byte) ([]byte, error) {
	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("%w: got %d bytes", ErrInvalidNonceSize, len(nonce))
	}
	if len(ciphertextWithTag) < TagSize {
		return nil, errors.New("ciphertext must include at least the authentication tag")
	}

	packed := make([]byte, len(nonce)+len(ciphertextWithTag))
	copy(packed[:NonceSize], nonce)
	copy(packed[NonceSize:], ciphertextWithTag)
	return packed, nil
}

// UnpackPayload parses a packed byte slice into its nonce and ciphertextWithTag components.
func UnpackPayload(packed []byte) (nonce []byte, ciphertextWithTag []byte, err error) {
	if len(packed) < MinPayloadSize {
		return nil, nil, fmt.Errorf("%w: got %d bytes, need at least %d", ErrPayloadTooShort, len(packed), MinPayloadSize)
	}

	nonce = make([]byte, NonceSize)
	copy(nonce, packed[:NonceSize])

	ciphertextWithTag = make([]byte, len(packed)-NonceSize)
	copy(ciphertextWithTag, packed[NonceSize:])

	return nonce, ciphertextWithTag, nil
}
