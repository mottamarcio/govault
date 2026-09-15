package kdf

import (
	"bytes"
	"testing"
)

func TestDeriveMasterKey(t *testing.T) {
	// Lightweight params for fast unit testing
	params := &Argon2Params{
		Time:    1,
		Memory:  8 * 1024,
		Threads: 1,
		KeyLen:  32,
		Salt:    []byte("01234567890123456789012345678901"),
	}

	password := []byte("correct-horse-battery-staple")

	key1, err := DeriveMasterKey(password, params)
	if err != nil {
		t.Fatalf("DeriveMasterKey failed: %v", err)
	}
	if len(key1) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key1))
	}

	// Determinism check
	key2, err := DeriveMasterKey(password, params)
	if err != nil {
		t.Fatalf("DeriveMasterKey failed: %v", err)
	}
	if !bytes.Equal(key1, key2) {
		t.Fatalf("expected deterministic output for identical inputs")
	}

	// Different password -> different key
	key3, err := DeriveMasterKey([]byte("different-password"), params)
	if err != nil {
		t.Fatalf("DeriveMasterKey failed: %v", err)
	}
	if bytes.Equal(key1, key3) {
		t.Fatalf("expected different keys for different passwords")
	}
}

func TestDeriveMasterKeyInvalidParams(t *testing.T) {
	_, err := DeriveMasterKey([]byte("password"), nil)
	if err == nil {
		t.Fatalf("expected error for nil params")
	}
}
