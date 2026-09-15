package cipher

import (
	"bytes"
	"testing"
)

func TestEnvelopePacking(t *testing.T) {
	nonce, _ := GenerateNonce()
	ciphertextWithTag := []byte("secret-encrypted-data-with-tag-123456")

	packed, err := PackPayload(nonce, ciphertextWithTag)
	if err != nil {
		t.Fatalf("PackPayload failed: %v", err)
	}

	if len(packed) != len(nonce)+len(ciphertextWithTag) {
		t.Fatalf("unexpected packed length: %d", len(packed))
	}

	unpackedNonce, unpackedCiphertext, err := UnpackPayload(packed)
	if err != nil {
		t.Fatalf("UnpackPayload failed: %v", err)
	}

	if !bytes.Equal(nonce, unpackedNonce) {
		t.Errorf("unpacked nonce mismatch")
	}
	if !bytes.Equal(ciphertextWithTag, unpackedCiphertext) {
		t.Errorf("unpacked ciphertext mismatch")
	}
}

func TestUnpackPayloadTooShort(t *testing.T) {
	short := make([]byte, 39) // 1 byte short of MinPayloadSize (40)
	_, _, err := UnpackPayload(short)
	if err == nil {
		t.Fatalf("expected error for payload < 40 bytes")
	}
}
