package cipher

import (
	"bytes"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	key := make([]byte, KeySize)
	for i := range key {
		key[i] = byte(i)
	}

	aad := []byte("govault/v1|test-context")

	testCases := [][]byte{
		[]byte(""), // empty plaintext
		[]byte("hello world"),
		[]byte("super-secret-password-123!@#"),
		bytes.Repeat([]byte("A"), 1024),   // 1 KB
		bytes.Repeat([]byte("B"), 64*1024), // 64 KB
	}

	for _, pt := range testCases {
		ciphertext, err := Encrypt(key, pt, aad)
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}

		if len(ciphertext) != len(pt)+NonceSize+TagSize {
			t.Fatalf("expected ciphertext length %d, got %d", len(pt)+NonceSize+TagSize, len(ciphertext))
		}

		decrypted, err := Decrypt(key, ciphertext, aad)
		if err != nil {
			t.Fatalf("Decrypt failed: %v", err)
		}

		if !bytes.Equal(decrypted, pt) {
			t.Fatalf("decrypted plaintext does not match original")
		}
	}
}

func TestTamperDetection(t *testing.T) {
	key := make([]byte, KeySize)
	aad := []byte("govault/v1|test-context")
	plaintext := []byte("sensitive-user-secret")

	ciphertext, err := Encrypt(key, plaintext, aad)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// 1. Bit flip in ciphertext / tag
	tampered := append([]byte(nil), ciphertext...)
	tampered[len(tampered)-1] ^= 0x01

	if _, err := Decrypt(key, tampered, aad); err == nil {
		t.Fatalf("expected decryption failure for tampered ciphertext")
	}

	// 2. Bit flip in nonce
	tamperedNonce := append([]byte(nil), ciphertext...)
	tamperedNonce[0] ^= 0x01

	if _, err := Decrypt(key, tamperedNonce, aad); err == nil {
		t.Fatalf("expected decryption failure for tampered nonce")
	}
}

func TestAADMismatchRejection(t *testing.T) {
	key := make([]byte, KeySize)
	aadRecordA := AADContext{VaultID: "v1", RecordID: "rec-A", RecordType: "login", Version: 1}.Bytes()
	aadRecordB := AADContext{VaultID: "v1", RecordID: "rec-B", RecordType: "login", Version: 1}.Bytes()

	plaintext := []byte("record-a-secret-data")

	ciphertext, err := Encrypt(key, plaintext, aadRecordA)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	// Decrypt with incorrect AAD (Record B context)
	if _, err := Decrypt(key, ciphertext, aadRecordB); err == nil {
		t.Fatalf("expected decryption to fail when AAD context is mismatched")
	}
}

func TestInvalidKeySize(t *testing.T) {
	shortKey := make([]byte, 16)
	if _, err := Encrypt(shortKey, []byte("test"), nil); err == nil {
		t.Fatalf("expected error for invalid key size in Encrypt")
	}
	if _, err := Decrypt(shortKey, make([]byte, 40), nil); err == nil {
		t.Fatalf("expected error for invalid key size in Decrypt")
	}
}
