package kdf

import (
	"fmt"

	"golang.org/x/crypto/argon2"
)

// DeriveMasterKey derives a cryptographic master key from the master password and Argon2id parameters.
func DeriveMasterKey(password []byte, params *Argon2Params) ([]byte, error) {
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("cannot derive key with invalid params: %w", err)
	}

	key := argon2.IDKey(
		password,
		params.Salt,
		params.Time,
		params.Memory,
		params.Threads,
		params.KeyLen,
	)

	return key, nil
}
