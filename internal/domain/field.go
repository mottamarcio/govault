package domain

import "strings"

// Field represents an arbitrary key-value custom attribute on a secret record.
type Field struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Masked bool   `json:"masked"`
}

// Validate checks the validity of a Field.
func (f Field) Validate() error {
	if strings.TrimSpace(f.Key) == "" {
		return ErrEmptyKey
	}
	return nil
}
