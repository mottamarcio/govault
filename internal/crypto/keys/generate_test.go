package keys

import (
	"bytes"
	"testing"
)

func TestGenerateVaultKey(t *testing.T) {
	vk1, err := GenerateVaultKey()
	if err != nil {
		t.Fatalf("GenerateVaultKey failed: %v", err)
	}
	if len(vk1) != VaultKeySize {
		t.Fatalf("expected vault key size %d, got %d", VaultKeySize, len(vk1))
	}

	vk2, err := GenerateVaultKey()
	if err != nil {
		t.Fatalf("GenerateVaultKey failed: %v", err)
	}
	if bytes.Equal(vk1, vk2) {
		t.Fatalf("expected distinct random vault keys")
	}

	ScrubKey(vk1)
	for i, b := range vk1 {
		if b != 0 {
			t.Fatalf("expected scrubbed key byte at %d to be 0, got %d", i, b)
		}
	}
}
