package keys

import (
	"bytes"
	"testing"

	"github.com/mottamarcio/govault/internal/crypto/cipher"
	"github.com/mottamarcio/govault/internal/crypto/kdf"
)

func TestWrappedKeyEnvelopeSerialization(t *testing.T) {
	params, _ := kdf.DefaultArgon2Params()
	env := &WrappedKeyEnvelope{
		KDFParams:  params,
		WrappedKey: make([]byte, cipher.MinPayloadSize+32),
		MCK:        make([]byte, 32),
	}

	data, err := env.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	restored, err := UnmarshalEnvelope(data)
	if err != nil {
		t.Fatalf("UnmarshalEnvelope failed: %v", err)
	}

	if !bytes.Equal(env.WrappedKey, restored.WrappedKey) {
		t.Errorf("wrapped key mismatch")
	}
	if !bytes.Equal(env.MCK, restored.MCK) {
		t.Errorf("MCK mismatch")
	}
}
