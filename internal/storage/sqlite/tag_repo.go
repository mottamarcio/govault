package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Tag represents a categorization tag entity.
type Tag struct {
	ID        string
	VaultID   string
	Name      string
	CreatedAt time.Time
}

// TagRepository manages persistence of tags and record_tags associations.
type TagRepository struct {
	db *DB
}

// NewTagRepository creates a new TagRepository.
func NewTagRepository(db *DB) *TagRepository {
	return &TagRepository{db: db}
}

// Create persists a new tag or returns existing tag if name already exists in vault.
func (r *TagRepository) Create(ctx context.Context, vaultID, name string) (*Tag, error) {
	if name == "" {
		return nil, errors.New("tag name cannot be empty")
	}

	id := uuid.New().String()
	now := time.Now().Unix()

	query := `
		INSERT INTO tags (id, vault_id, name, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(vault_id, name) DO UPDATE SET name=name
		RETURNING id, vault_id, name, created_at;
	`

	var tag Tag
	var createdAtUnix int64
	err := r.db.QueryRowContext(ctx, query, id, vaultID, name, now).Scan(
		&tag.ID,
		&tag.VaultID,
		&tag.Name,
		&createdAtUnix,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	tag.CreatedAt = time.Unix(createdAtUnix, 0)
	return &tag, nil
}

// List returns all tags for a given vault.
func (r *TagRepository) List(ctx context.Context, vaultID string) ([]Tag, error) {
	query := `SELECT id, vault_id, name, created_at FROM tags WHERE vault_id = ? ORDER BY name ASC;`
	rows, err := r.db.QueryContext(ctx, query, vaultID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		var createdAtUnix int64
		if err := rows.Scan(&tag.ID, &tag.VaultID, &tag.Name, &createdAtUnix); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tag.CreatedAt = time.Unix(createdAtUnix, 0)
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// Delete removes a tag by ID.
func (r *TagRepository) Delete(ctx context.Context, vaultID, tagID string) error {
	query := `DELETE FROM tags WHERE vault_id = ? AND id = ?;`
	_, err := r.db.ExecContext(ctx, query, vaultID, tagID)
	return err
}

// GetRecordTags returns all tag IDs associated with a record.
func (r *TagRepository) GetRecordTags(ctx context.Context, recordID string) ([]string, error) {
	query := `SELECT tag_id FROM record_tags WHERE record_id = ?;`
	rows, err := r.db.QueryContext(ctx, query, recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to query record tags: %w", err)
	}
	defer rows.Close()

	var tagIDs []string
	for rows.Next() {
		var tid string
		if err := rows.Scan(&tid); err != nil {
			return nil, err
		}
		tagIDs = append(tagIDs, tid)
	}
	return tagIDs, rows.Err()
}

// SetRecordTagsTx associates tag IDs with a record inside a transaction.
func (r *TagRepository) SetRecordTagsTx(ctx context.Context, tx *sql.Tx, recordID string, tagIDs []string) error {
	if _, err := tx.ExecContext(ctx, "DELETE FROM record_tags WHERE record_id = ?;", recordID); err != nil {
		return fmt.Errorf("failed to clear old record tags: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT OR IGNORE INTO record_tags (record_id, tag_id) VALUES (?, ?);")
	if err != nil {
		return fmt.Errorf("failed to prepare tag association statement: %w", err)
	}
	defer stmt.Close()

	for _, tid := range tagIDs {
		if _, err := stmt.ExecContext(ctx, recordID, tid); err != nil {
			return fmt.Errorf("failed to associate tag %s: %w", tid, err)
		}
	}

	return nil
}
