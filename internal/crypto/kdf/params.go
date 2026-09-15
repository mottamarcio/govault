package kdf

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Default RFC 9106 recommended Argon2id parameters.
const (
	DefaultArgon2Time        uint32 = 3
	DefaultArgon2Memory      uint32 = 64 * 1024 // 64 MB in KiB
	DefaultArgon2Threads     uint8  = 4
	DefaultArgon2KeyLen      uint32 = 32
	DefaultArgon2SaltLen     int    = 32
	MinArgon2Time            uint32 = 1
	MinArgon2Memory          uint32 = 8 * 1024 // 8 MB min
	MinArgon2Threads         uint8  = 1
	MinArgon2KeyLen          uint32 = 16
	MinArgon2SaltLen         int    = 16
)

var (
	ErrInvalidParams = errors.New("invalid argon2 parameters")
)

// Argon2Params defines configuration parameters for Argon2id key derivation.
type Argon2Params struct {
	Time    uint32 `json:"time"`
	Memory  uint32 `json:"memory"`
	Threads uint8  `json:"threads"`
	KeyLen  uint32 `json:"key_len"`
	Salt    []byte `json:"salt"`
}

// DefaultArgon2Params creates a standard RFC 9106 Argon2id parameter set with a fresh CSPRNG salt.
func DefaultArgon2Params() (*Argon2Params, error) {
	salt := make([]byte, DefaultArgon2SaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}

	return &Argon2Params{
		Time:    DefaultArgon2Time,
		Memory:  DefaultArgon2Memory,
		Threads: DefaultArgon2Threads,
		KeyLen:  DefaultArgon2KeyLen,
		Salt:    salt,
	}, nil
}

// Validate checks whether the parameters meet the required security baselines.
func (p *Argon2Params) Validate() error {
	if p == nil {
		return fmt.Errorf("%w: params cannot be nil", ErrInvalidParams)
	}
	if p.Time < MinArgon2Time {
		return fmt.Errorf("%w: time iterations %d < %d", ErrInvalidParams, p.Time, MinArgon2Time)
	}
	if p.Memory < MinArgon2Memory {
		return fmt.Errorf("%w: memory %d KiB < %d KiB", ErrInvalidParams, p.Memory, MinArgon2Memory)
	}
	if p.Threads < MinArgon2Threads {
		return fmt.Errorf("%w: threads %d < %d", ErrInvalidParams, p.Threads, MinArgon2Threads)
	}
	if p.KeyLen < MinArgon2KeyLen {
		return fmt.Errorf("%w: key length %d < %d", ErrInvalidParams, p.KeyLen, MinArgon2KeyLen)
	}
	if len(p.Salt) < MinArgon2SaltLen {
		return fmt.Errorf("%w: salt length %d < %d", ErrInvalidParams, len(p.Salt), MinArgon2SaltLen)
	}
	return nil
}

// MarshalJSON encodes parameters to JSON format.
func (p *Argon2Params) Marshal() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(p)
}

// UnmarshalParams decodes parameters from JSON.
func UnmarshalParams(data []byte) (*Argon2Params, error) {
	var p Argon2Params
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("failed to unmarshal argon2 params: %w", err)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}
