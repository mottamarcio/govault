package sqlite

import (
	"context"
	"fmt"
	"time"
)

// HistoryEntry represents an archived previous version of an encrypted record.
type HistoryEntry struct {
	ID         string
	RecordID   string
	VaultID    string
	Version    uint32
	Payload    []byte
	ArchivedAt time.Time
}

// HistoryRepository handles querying previous versions from record_history.
type HistoryRepository struct {
	db *DB
}

// NewHistoryRepository creates a new HistoryRepository.
func NewHistoryRepository(db *DB) *HistoryRepository {
	return &HistoryRepository{db: db}
}

// ListByRecordID returns all archived history entries for a record in descending version order.
func (r *HistoryRepository) ListByRecordID(ctx context.Context, recordID string) ([]HistoryEntry, error) {
	query := `
		SELECT id, record_id, vault_id, version, payload, archived_at
		FROM record_history
		WHERE record_id = ?
		ORDER BY version DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to query record history: %w", err)
	}
	defer rows.Close()

	var entries []HistoryEntry
	for rows.Next() {
		var e HistoryEntry
		var archivedAtUnix int64
		err := rows.Scan(&e.ID, &e.RecordID, &e.VaultID, &e.Version, &e.Payload, &archivedAtUnix)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history entry: %w", err)
		}
		e.ArchivedAt = time.Unix(archivedAtUnix, 0)
		entries = append(entries, e)
	}

	return entries, rows.Err()
}

// GetVersion returns a specific archived version of a record.
func (r *HistoryRepository) GetVersion(ctx context.Context, recordID string, version uint32) (*HistoryEntry, error) {
	query := `
		SELECT id, record_id, vault_id, version, payload, archived_at
		FROM record_history
		WHERE record_id = ? AND version = ?;
	`
	var e HistoryEntry
	var archivedAtUnix int64
	err := r.db.QueryRowContext(ctx, query, recordID, version).Scan(
		&e.ID,
		&e.RecordID,
		&e.VaultID,
		&e.Version,
		&e.Payload,
		&archivedAtUnix,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query history version %d: %w", version, err)
	}

	e.ArchivedAt = time.Unix(archivedAtUnix, 0)
	return &e, nil
}
