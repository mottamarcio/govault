package domain

import "time"

// HistoryEntry represents an archived historical revision of an encrypted record.
type HistoryEntry struct {
	ID         string    `json:"id"`
	RecordID   string    `json:"record_id"`
	VaultID    string    `json:"vault_id"`
	Version    uint32    `json:"version"`
	Payload    any       `json:"payload,omitempty"` // Typed decrypted payload or raw []byte
	ArchivedAt time.Time `json:"archived_at"`
}
