package domain

import (
	"strings"
)

const (
	MaxTitleLength   = 255
	MaxTagNameLength = 64
)

// Validate checks the structural and business invariants of a Record.
func (r *Record) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return ErrInvalidUUID
	}
	if strings.TrimSpace(r.VaultID) == "" {
		return ErrInvalidUUID
	}
	if strings.TrimSpace(r.Title) == "" {
		return ErrEmptyTitle
	}
	if len(r.Title) > MaxTitleLength {
		return ErrTitleTooLong
	}
	if !r.Type.IsValid() {
		return ErrInvalidRecordType
	}
	return nil
}

// Validate checks the constraints of a LoginPayload.
func (p *LoginPayload) Validate() error {
	return validateUniqueFieldKeys(p.CustomFields)
}

// Validate checks the constraints of a NotePayload.
func (p *NotePayload) Validate() error {
	return validateUniqueFieldKeys(p.CustomFields)
}

// Validate checks the constraints of an APIKeyPayload.
func (p *APIKeyPayload) Validate() error {
	return validateUniqueFieldKeys(p.CustomFields)
}

// Validate checks the constraints of a CustomPayload.
func (p *CustomPayload) Validate() error {
	return validateUniqueFieldKeys(p.Fields)
}

// Validate checks the constraints of a Tag.
func (t *Tag) Validate() error {
	if strings.TrimSpace(t.ID) == "" {
		return ErrInvalidUUID
	}
	if strings.TrimSpace(t.VaultID) == "" {
		return ErrInvalidUUID
	}
	if strings.TrimSpace(t.Name) == "" {
		return ErrEmptyTagName
	}
	if len(t.Name) > MaxTagNameLength {
		return ErrTagNameTooLong
	}
	return nil
}
