package kdf

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

// Domain separation info tags.
const (
	InfoKEK = "govault/v1/kek"
	InfoMCK = "govault/v1/mck"

	DerivedKeyLen = 32 // Standard 256-bit key length
)

var (
	ErrInvalidMasterKey = errors.New("invalid master key: must be 32 bytes")
)

// DeriveSubKey derives a key from masterKey using HKDF-SHA-256 with specific domain separation info.
func DeriveSubKey(masterKey []byte, info string, length int) ([]byte, error) {
	if len(masterKey) != 32 {
		return nil, ErrInvalidMasterKey
	}
	if length <= 0 {
		return nil, errors.New("sub-key length must be positive")
	}

	reader := hkdf.New(sha256.New, masterKey, nil, []byte(info))
	subKey := make([]byte, length)
	if _, err := io.ReadFull(reader, subKey); err != nil {
		return nil, fmt.Errorf("failed to derive sub-key: %w", err)
	}

	return subKey, nil
}

// DeriveKEK derives the Key Encryption Key (KEK) from the Master Key.
func DeriveKEK(masterKey []byte) ([]byte, error) {
	return DeriveSubKey(masterKey, InfoKEK, DerivedKeyLen)
}

// DeriveMCK derives the Master Confirmation Key (MCK) from the Master Key.
func DeriveMCK(masterKey []byte) ([]byte, error) {
	return DeriveSubKey(masterKey, InfoMCK, DerivedKeyLen)
}
