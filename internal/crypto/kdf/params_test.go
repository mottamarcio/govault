package kdf

import (
	"bytes"
	"testing"
)

func TestDefaultArgon2Params(t *testing.T) {
	p, err := DefaultArgon2Params()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Time != DefaultArgon2Time {
		t.Errorf("expected time %d, got %d", DefaultArgon2Time, p.Time)
	}
	if p.Memory != DefaultArgon2Memory {
		t.Errorf("expected memory %d, got %d", DefaultArgon2Memory, p.Memory)
	}
	if p.Threads != DefaultArgon2Threads {
		t.Errorf("expected threads %d, got %d", DefaultArgon2Threads, p.Threads)
	}
	if p.KeyLen != DefaultArgon2KeyLen {
		t.Errorf("expected key len %d, got %d", DefaultArgon2KeyLen, p.KeyLen)
	}
	if len(p.Salt) != DefaultArgon2SaltLen {
		t.Errorf("expected salt length %d, got %d", DefaultArgon2SaltLen, len(p.Salt))
	}
	if err := p.Validate(); err != nil {
		t.Errorf("validation failed on defaults: %v", err)
	}
}

func TestArgon2ParamsValidation(t *testing.T) {
	valid, _ := DefaultArgon2Params()

	tests := []struct {
		name    string
		mutate  func(p *Argon2Params)
		wantErr bool
	}{
		{
			name:    "nil params",
			mutate:  nil,
			wantErr: true,
		},
		{
			name: "time too low",
			mutate: func(p *Argon2Params) {
				p.Time = 0
			},
			wantErr: true,
		},
		{
			name: "memory too low",
			mutate: func(p *Argon2Params) {
				p.Memory = 1024
			},
			wantErr: true,
		},
		{
			name: "threads zero",
			mutate: func(p *Argon2Params) {
				p.Threads = 0
			},
			wantErr: true,
		},
		{
			name: "key length too short",
			mutate: func(p *Argon2Params) {
				p.KeyLen = 8
			},
			wantErr: true,
		},
		{
			name: "salt too short",
			mutate: func(p *Argon2Params) {
				p.Salt = []byte("short")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mutate == nil {
				var nilP *Argon2Params
				if err := nilP.Validate(); err == nil {
					t.Errorf("expected error for nil params, got nil")
				}
				return
			}
			clone := *valid
			clone.Salt = append([]byte(nil), valid.Salt...)
			tt.mutate(&clone)
			err := clone.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestArgon2ParamsSerialization(t *testing.T) {
	p, err := DefaultArgon2Params()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := p.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal params: %v", err)
	}

	restored, err := UnmarshalParams(data)
	if err != nil {
		t.Fatalf("failed to unmarshal params: %v", err)
	}

	if restored.Time != p.Time || restored.Memory != p.Memory || restored.Threads != p.Threads || restored.KeyLen != p.KeyLen {
		t.Errorf("restored fields do not match original")
	}
	if !bytes.Equal(restored.Salt, p.Salt) {
		t.Errorf("restored salt does not match original")
	}
}
