package keys

import (
	"crypto/rand"
	"fmt"
	"io"

	"github.com/mottamarcio/govault/internal/crypto/kdf"
)

const (
	// VaultKeySize specifies the 256-bit (32-byte) length for the master Vault Key.
	VaultKeySize = 32
)

// GenerateVaultKey generates a high-entropy 32-byte Vault Key using CSPRNG.
func GenerateVaultKey() ([]byte, error) {
	key := make([]byte, VaultKeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate random vault key: %w", err)
	}
	return key, nil
}

// ScrubKey zeroes out key material from memory.
func ScrubKey(key []byte) {
	kdf.Zeroize(key)
}
