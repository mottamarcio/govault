package cipher

import (
	"bytes"
	"testing"
)

func TestNonceGeneration(t *testing.T) {
	n1, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce failed: %v", err)
	}
	if len(n1) != NonceSize {
		t.Fatalf("expected nonce size %d, got %d", NonceSize, len(n1))
	}

	n2, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce failed: %v", err)
	}
	if bytes.Equal(n1, n2) {
		t.Fatalf("expected consecutive nonces to be random and distinct")
	}

	if err := ValidateNonce(n1); err != nil {
		t.Fatalf("ValidateNonce failed for valid nonce: %v", err)
	}
}

func TestValidateNonceInvalid(t *testing.T) {
	tests := [][]byte{
		nil,
		{},
		make([]byte, 12), // standard 96-bit ChaCha20 nonce (must be rejected)
		make([]byte, 32),
	}

	for _, n := range tests {
		if err := ValidateNonce(n); err == nil {
			t.Errorf("expected error for invalid nonce length %d, got nil", len(n))
		}
	}
}
