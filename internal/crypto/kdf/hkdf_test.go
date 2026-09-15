package kdf

import (
	"bytes"
	"testing"
)

func TestDeriveSubKeys(t *testing.T) {
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = byte(i)
	}

	kek, err := DeriveKEK(masterKey)
	if err != nil {
		t.Fatalf("DeriveKEK failed: %v", err)
	}
	if len(kek) != 32 {
		t.Fatalf("expected 32-byte KEK, got %d", len(kek))
	}

	mck, err := DeriveMCK(masterKey)
	if err != nil {
		t.Fatalf("DeriveMCK failed: %v", err)
	}
	if len(mck) != 32 {
		t.Fatalf("expected 32-byte MCK, got %d", len(mck))
	}

	// KEK and MCK must be distinct (domain separation)
	if bytes.Equal(kek, mck) {
		t.Fatalf("KEK and MCK must not be identical")
	}
	if bytes.Equal(kek, masterKey) || bytes.Equal(mck, masterKey) {
		t.Fatalf("derived keys must not match master key")
	}

	// Determinism
	kek2, _ := DeriveKEK(masterKey)
	if !bytes.Equal(kek, kek2) {
		t.Fatalf("DeriveKEK not deterministic")
	}
}

func TestDeriveSubKeysInvalidMasterKey(t *testing.T) {
	shortKey := []byte("too-short")
	if _, err := DeriveKEK(shortKey); err == nil {
		t.Errorf("expected error for short master key in DeriveKEK")
	}
	if _, err := DeriveMCK(shortKey); err == nil {
		t.Errorf("expected error for short master key in DeriveMCK")
	}
}
