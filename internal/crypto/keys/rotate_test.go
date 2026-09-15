package keys

import (
	"bytes"
	"testing"

	"github.com/mottamarcio/govault/internal/crypto/cipher"
)

func TestMasterPasswordRotation(t *testing.T) {
	vaultID := "vault-uuid-rotate-test"
	oldPassword := []byte("old-super-secure-password-1")
	newPassword := []byte("new-super-secure-password-2")

	params := fastParams()

	// 1. Generate Vault Key and wrap with old password
	originalVK, err := GenerateVaultKey()
	if err != nil {
		t.Fatalf("GenerateVaultKey failed: %v", err)
	}

	env, err := WrapVaultKey(originalVK, oldPassword, vaultID, params)
	if err != nil {
		t.Fatalf("WrapVaultKey failed: %v", err)
	}

	// 2. Encrypt a mock record payload with the original Vault Key
	recordPlaintext := []byte("my-bank-login-credentials-12345")
	recordAAD := cipher.AADContext{VaultID: vaultID, RecordID: "rec-1", RecordType: "login", Version: 1}.Bytes()
	recordCiphertext, err := cipher.Encrypt(originalVK, recordPlaintext, recordAAD)
	if err != nil {
		t.Fatalf("Encrypt record failed: %v", err)
	}

	// 3. Perform Master Password Rotation
	newParams := fastParams()
	newParams.Salt = []byte("different-salt-9876543210123456")

	newEnv, err := RotateMasterPassword(env, oldPassword, newPassword, vaultID, newParams)
	if err != nil {
		t.Fatalf("RotateMasterPassword failed: %v", err)
	}

	// 4. Old password should fail to unwrap new envelope
	if _, err := UnwrapVaultKey(newEnv, oldPassword, vaultID); err == nil {
		t.Fatalf("expected old password to fail unwrapping rotated envelope")
	}

	// 5. New password unwraps the EXACT same Vault Key
	recoveredVK, err := UnwrapVaultKey(newEnv, newPassword, vaultID)
	if err != nil {
		t.Fatalf("UnwrapVaultKey with new password failed: %v", err)
	}
	if !bytes.Equal(originalVK, recoveredVK) {
		t.Fatalf("recovered vault key after rotation does not match original")
	}

	// 6. Record ciphertext encrypted prior to rotation still decrypts cleanly
	decryptedRecord, err := cipher.Decrypt(recoveredVK, recordCiphertext, recordAAD)
	if err != nil {
		t.Fatalf("failed to decrypt record with recovered vault key: %v", err)
	}
	if !bytes.Equal(recordPlaintext, decryptedRecord) {
		t.Fatalf("decrypted record plaintext does not match original")
	}
}

func TestRotateMasterPasswordWrongOldPassword(t *testing.T) {
	vaultID := "vault-uuid-wrong-old"
	oldPassword := []byte("actual-old-password")
	wrongPassword := []byte("wrong-old-password")
	newPassword := []byte("new-password")

	originalVK, _ := GenerateVaultKey()
	env, _ := WrapVaultKey(originalVK, oldPassword, vaultID, fastParams())

	_, err := RotateMasterPassword(env, wrongPassword, newPassword, vaultID, fastParams())
	if err == nil {
		t.Fatalf("expected rotation to fail with wrong old password")
	}
}
