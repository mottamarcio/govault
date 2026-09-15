package format

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/mottamarcio/govault/internal/domain"
)

const (
	CurrentPayloadVersion = 1
)

// BackupManifest contains metadata about the backed-up vault.
type BackupManifest struct {
	SourceVaultVersion uint16 `json:"source_vault_version"`
	SourceSchemaVersion uint16 `json:"source_schema_version"`
	AppVersion          string `json:"app_version"`
	EntryCount          int    `json:"entry_count"`
	HistoryCount        int    `json:"history_count"`
	TagCount            int    `json:"tag_count"`
	CreatedAtUnix       int64  `json:"created_at_unix"`
}

// BackupEntry represents a logical decrypted vault record in the backup stream.
type BackupEntry struct {
	ID            string            `json:"id"`
	Type          domain.RecordType `json:"type"`
	Title         string            `json:"title"`
	Payload       []byte            `json:"payload"`
	Version       int               `json:"version"`
	CreatedAtUnix int64             `json:"created_at_unix"`
	UpdatedAtUnix int64             `json:"updated_at_unix"`
	Tags          []string          `json:"tags"`
	IsDeleted     bool              `json:"is_deleted"`
}

// BackupHistory represents a historical revision of a vault record.
type BackupHistory struct {
	HistoryID     string `json:"history_id"`
	RecordID      string `json:"record_id"`
	Version       int    `json:"version"`
	Payload       []byte `json:"payload"`
	CreatedAtUnix int64  `json:"created_at_unix"`
}

// BackupTag represents a tag definition in the vault.
type BackupTag struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	CreatedAtUnix int64  `json:"created_at_unix"`
}

// BackupPayload is the root plaintext structure that is canonically encoded before encryption.
type BackupPayload struct {
	Version   uint16          `json:"version"`
	Manifest  BackupManifest  `json:"manifest"`
	Entries   []BackupEntry   `json:"entries"`
	Histories []BackupHistory `json:"histories"`
	Tags      []BackupTag     `json:"tags"`
}

// SortNormalizes sorts entries, histories, and tags deterministically.
func (p *BackupPayload) SortNormalizes() {
	sort.Slice(p.Entries, func(i, j int) bool {
		return p.Entries[i].ID < p.Entries[j].ID
	})
	sort.Slice(p.Histories, func(i, j int) bool {
		if p.Histories[i].RecordID == p.Histories[j].RecordID {
			return p.Histories[i].Version < p.Histories[j].Version
		}
		return p.Histories[i].RecordID < p.Histories[j].RecordID
	})
	sort.Slice(p.Tags, func(i, j int) bool {
		return p.Tags[i].Name < p.Tags[j].Name
	})
}

// Marshal encodes the BackupPayload into canonical JSON bytes.
func (p *BackupPayload) Marshal() ([]byte, error) {
	if p.Version == 0 {
		p.Version = CurrentPayloadVersion
	}
	p.Manifest.EntryCount = len(p.Entries)
	p.Manifest.HistoryCount = len(p.Histories)
	p.Manifest.TagCount = len(p.Tags)

	p.SortNormalizes()

	return json.Marshal(p)
}

// UnmarshalPayload decodes bytes into BackupPayload and performs logical validation.
func UnmarshalPayload(data []byte) (*BackupPayload, error) {
	var payload BackupPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptPayload, err)
	}

	if payload.Version != CurrentPayloadVersion {
		return nil, fmt.Errorf("%w: unsupported payload version %d", ErrCorruptPayload, payload.Version)
	}

	if payload.Manifest.EntryCount != len(payload.Entries) {
		return nil, fmt.Errorf("%w: manifest entry count mismatch: declared %d, actual %d", ErrCorruptPayload, payload.Manifest.EntryCount, len(payload.Entries))
	}
	if payload.Manifest.HistoryCount != len(payload.Histories) {
		return nil, fmt.Errorf("%w: manifest history count mismatch: declared %d, actual %d", ErrCorruptPayload, payload.Manifest.HistoryCount, len(payload.Histories))
	}
	if payload.Manifest.TagCount != len(payload.Tags) {
		return nil, fmt.Errorf("%w: manifest tag count mismatch: declared %d, actual %d", ErrCorruptPayload, payload.Manifest.TagCount, len(payload.Tags))
	}

	// Validate duplicate Entry IDs
	entryMap := make(map[string]bool, len(payload.Entries))
	for _, e := range payload.Entries {
		if entryMap[e.ID] {
			return nil, fmt.Errorf("%w: duplicate entry ID %s", ErrCorruptPayload, e.ID)
		}
		entryMap[e.ID] = true
	}

	// Validate duplicate History IDs and broken references
	historyMap := make(map[string]bool, len(payload.Histories))
	for _, h := range payload.Histories {
		if historyMap[h.HistoryID] {
			return nil, fmt.Errorf("%w: duplicate history ID %s", ErrCorruptPayload, h.HistoryID)
		}
		historyMap[h.HistoryID] = true
		if !entryMap[h.RecordID] {
			return nil, fmt.Errorf("%w: orphaned history record referencing missing entry %s", ErrCorruptPayload, h.RecordID)
		}
	}

	return &payload, nil
}
