package cipher

import (
	"bytes"
	"testing"
)

func TestAADContextSerialization(t *testing.T) {
	ctx1 := AADContext{
		VaultID:    "vault-123",
		RecordID:   "rec-456",
		RecordType: "login",
		Version:    1,
	}

	b1 := ctx1.Bytes()
	expected := []byte("govault/v1|v:vault-123|r:rec-456|t:login|ver:1")
	if !bytes.Equal(b1, expected) {
		t.Fatalf("expected AAD bytes %q, got %q", expected, b1)
	}

	// Changing version must yield different AAD
	ctx2 := ctx1
	ctx2.Version = 2
	if bytes.Equal(ctx1.Bytes(), ctx2.Bytes()) {
		t.Fatalf("different versions must yield different AAD bytes")
	}

	// SimpleAAD check
	simple := SimpleAAD("backup_key")
	if !bytes.Equal(simple, []byte("govault/v1|label:backup_key")) {
		t.Fatalf("unexpected SimpleAAD output: %q", simple)
	}
}
