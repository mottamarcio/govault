package keys

import (
	"fmt"

	"github.com/mottamarcio/govault/internal/crypto/kdf"
)

// RotateMasterPassword re-wraps the underlying Vault Key using a new master password and fresh KDF parameters
// without requiring re-encryption of the records protected by the Vault Key.
func RotateMasterPassword(env *WrappedKeyEnvelope, oldPassword, newPassword []byte, vaultID string, newParams *kdf.Argon2Params) (*WrappedKeyEnvelope, error) {
	// 1. Unwrap the underlying Vault Key using the old password
	vaultKey, err := UnwrapVaultKey(env, oldPassword, vaultID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate old master password: %w", err)
	}
	defer kdf.Zeroize(vaultKey)

	// 2. Generate new parameters with a fresh salt if not explicitly provided
	if newParams == nil {
		newParams, err = kdf.DefaultArgon2Params()
		if err != nil {
			return nil, fmt.Errorf("failed to generate new KDF params: %w", err)
		}
	}

	// 3. Re-wrap the Vault Key with the new password
	newEnv, err := WrapVaultKey(vaultKey, newPassword, vaultID, newParams)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap vault key with new password: %w", err)
	}

	return newEnv, nil
}
