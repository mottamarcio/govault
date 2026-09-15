package domain

import "time"

// Record represents a secret entry within a vault.
type Record struct {
	ID        string     `json:"id"`
	VaultID   string     `json:"vault_id"`
	Title     string     `json:"title"`
	Type      RecordType `json:"type"`
	Tags      []string   `json:"tags,omitempty"`
	Version   uint32     `json:"version"`
	Payload   any        `json:"payload,omitempty"` // Typed payload (LoginPayload, NotePayload, APIKeyPayload, CustomPayload)
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// IsDeleted returns true if the record is currently soft-deleted.
func (r *Record) IsDeleted() bool {
	return r.DeletedAt != nil
}
