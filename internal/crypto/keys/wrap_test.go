package keys

import (
	"bytes"
	"testing"

	"github.com/mottamarcio/govault/internal/crypto/kdf"
)

func fastParams() *kdf.Argon2Params {
	return &kdf.Argon2Params{
		Time:    1,
		Memory:  8 * 1024,
		Threads: 1,
		KeyLen:  32,
		Salt:    []byte("01234567890123456789012345678901"),
	}
}

func TestWrapUnwrap(t *testing.T) {
	vaultID := "vault-uuid-1234"
	password := []byte("correct-horse-battery-staple")
	params := fastParams()

	originalVK, err := GenerateVaultKey()
	if err != nil {
		t.Fatalf("GenerateVaultKey failed: %v", err)
	}

	env, err := WrapVaultKey(originalVK, password, vaultID, params)
	if err != nil {
		t.Fatalf("WrapVaultKey failed: %v", err)
	}

	// 1. Successful unwrap
	unwrappedVK, err := UnwrapVaultKey(env, password, vaultID)
	if err != nil {
		t.Fatalf("UnwrapVaultKey failed: %v", err)
	}
	if !bytes.Equal(originalVK, unwrappedVK) {
		t.Fatalf("unwrapped vault key does not match original")
	}

	// 2. Wrong password rejection
	_, err = UnwrapVaultKey(env, []byte("wrong-password"), vaultID)
	if err == nil {
		t.Fatalf("expected unwrap to fail with wrong password")
	}

	// 3. Wrong vaultID AAD tampering rejection
	_, err = UnwrapVaultKey(env, password, "different-vault-uuid")
	if err == nil {
		t.Fatalf("expected unwrap to fail with mismatched vaultID")
	}
}
