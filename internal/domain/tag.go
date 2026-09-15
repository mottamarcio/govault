package domain

import "time"

// Tag represents a categorization label within a vault.
type Tag struct {
	ID        string    `json:"id"`
	VaultID   string    `json:"vault_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
