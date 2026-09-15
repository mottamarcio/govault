package keys

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/mottamarcio/govault/internal/crypto/cipher"
	"github.com/mottamarcio/govault/internal/crypto/kdf"
)

// WrappedKeyEnvelope holds the serialized Argon2 parameters, the encrypted Vault Key payload,
// and the Master Confirmation Key (MCK) authentication tag.
type WrappedKeyEnvelope struct {
	KDFParams  *kdf.Argon2Params `json:"kdf_params"`
	WrappedKey []byte            `json:"wrapped_key"`
	MCK        []byte            `json:"mck"`
}

// Validate checks whether the envelope fields are well-formed.
func (e *WrappedKeyEnvelope) Validate() error {
	if e == nil {
		return errors.New("envelope cannot be nil")
	}
	if err := e.KDFParams.Validate(); err != nil {
		return fmt.Errorf("invalid KDF params in envelope: %w", err)
	}
	if len(e.WrappedKey) < cipher.MinPayloadSize {
		return fmt.Errorf("wrapped key payload too short: got %d bytes, need at least %d", len(e.WrappedKey), cipher.MinPayloadSize)
	}
	if len(e.MCK) != 32 {
		return fmt.Errorf("invalid MCK length in envelope: got %d bytes, expected 32", len(e.MCK))
	}
	return nil
}

// Marshal encodes the envelope to JSON format.
func (e *WrappedKeyEnvelope) Marshal() ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(e)
}

// UnmarshalEnvelope decodes an envelope from JSON data.
func UnmarshalEnvelope(data []byte) (*WrappedKeyEnvelope, error) {
	var env WrappedKeyEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wrapped key envelope: %w", err)
	}
	if err := env.Validate(); err != nil {
		return nil, err
	}
	return &env, nil
}
