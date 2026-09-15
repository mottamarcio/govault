package keys

import (
	"errors"
	"fmt"

	"github.com/mottamarcio/govault/internal/crypto/cipher"
	"github.com/mottamarcio/govault/internal/crypto/kdf"
)

var (
	ErrAuthenticationFailed = errors.New("authentication failed: incorrect master password or corrupt envelope")
)

// VaultKeyAAD constructs the canonical AAD for wrapping/unwrapping the Vault Key.
func VaultKeyAAD(vaultID string) []byte {
	return []byte(fmt.Sprintf("govault/v1|vault_key:%s", vaultID))
}

// WrapVaultKey derives the Master Key, KEK, and MCK from the master password,
// and encrypts the Vault Key using XChaCha20-Poly1305 bound to the vault ID.
func WrapVaultKey(vaultKey []byte, password []byte, vaultID string, params *kdf.Argon2Params) (*WrappedKeyEnvelope, error) {
	if len(vaultKey) != VaultKeySize {
		return nil, fmt.Errorf("invalid vault key size: got %d bytes, need %d", len(vaultKey), VaultKeySize)
	}
	if len(vaultID) == 0 {
		return nil, errors.New("vault ID cannot be empty")
	}

	var err error
	if params == nil {
		params, err = kdf.DefaultArgon2Params()
		if err != nil {
			return nil, fmt.Errorf("failed to generate default KDF params: %w", err)
		}
	} else if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid KDF params: %w", err)
	}

	// 1. Derive Master Key
	masterKey, err := kdf.DeriveMasterKey(password, params)
	if err != nil {
		return nil, fmt.Errorf("failed to derive master key: %w", err)
	}
	defer kdf.Zeroize(masterKey)

	// 2. Derive KEK and MCK
	kek, err := kdf.DeriveKEK(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive KEK: %w", err)
	}
	defer kdf.Zeroize(kek)

	mck, err := kdf.DeriveMCK(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive MCK: %w", err)
	}

	// 3. Encrypt Vault Key with KEK and VaultKey AAD
	aad := VaultKeyAAD(vaultID)
	wrappedKeyPayload, err := cipher.Encrypt(kek, vaultKey, aad)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt vault key: %w", err)
	}

	return &WrappedKeyEnvelope{
		KDFParams:  params,
		WrappedKey: wrappedKeyPayload,
		MCK:        mck,
	}, nil
}

// UnwrapVaultKey derives candidate keys and verifies the password before decrypting the Vault Key.
func UnwrapVaultKey(env *WrappedKeyEnvelope, password []byte, vaultID string) ([]byte, error) {
	if err := env.Validate(); err != nil {
		return nil, fmt.Errorf("invalid envelope: %w", err)
	}
	if len(vaultID) == 0 {
		return nil, errors.New("vault ID cannot be empty")
	}

	// 1. Derive candidate Master Key
	masterKey, err := kdf.DeriveMasterKey(password, env.KDFParams)
	if err != nil {
		return nil, ErrAuthenticationFailed
	}
	defer kdf.Zeroize(masterKey)

	// 2. Derive candidate MCK and verify in constant time
	candidateMCK, err := kdf.DeriveMCK(masterKey)
	if err != nil {
		return nil, ErrAuthenticationFailed
	}
	defer kdf.Zeroize(candidateMCK)

	if !kdf.VerifyMCK(candidateMCK, env.MCK) {
		return nil, ErrAuthenticationFailed
	}

	// 3. Derive KEK and decrypt WrappedKey
	kek, err := kdf.DeriveKEK(masterKey)
	if err != nil {
		return nil, ErrAuthenticationFailed
	}
	defer kdf.Zeroize(kek)

	aad := VaultKeyAAD(vaultID)
	vaultKey, err := cipher.Decrypt(kek, env.WrappedKey, aad)
	if err != nil {
		return nil, ErrAuthenticationFailed
	}

	if len(vaultKey) != VaultKeySize {
		kdf.Zeroize(vaultKey)
		return nil, ErrAuthenticationFailed
	}

	return vaultKey, nil
}
