package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mottamarcio/govault/internal/crypto/cipher"
)

var (
	ErrRecordNotFound           = errors.New("record not found")
	ErrOptimisticLockingConflict = errors.New("record has been modified concurrently (optimistic locking conflict)")
)

// Record represents a persisted encrypted record entity.
type Record struct {
	ID         string
	VaultID    string
	RecordType string
	Version    uint32
	Payload    []byte
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

// RecordRepository manages database operations for records, tags, and history archiving.
type RecordRepository struct {
	db      *DB
	tagRepo *TagRepository
}

// NewRecordRepository creates a new RecordRepository.
func NewRecordRepository(db *DB, tagRepo *TagRepository) *RecordRepository {
	return &RecordRepository{
		db:      db,
		tagRepo: tagRepo,
	}
}

// Create persists a new encrypted record and optionally associates tag IDs.
func (r *RecordRepository) Create(ctx context.Context, record *Record, tagIDs []string) error {
	if record == nil {
		return errors.New("record cannot be nil")
	}
	if len(record.Payload) < cipher.MinPayloadSize {
		return fmt.Errorf("payload too short for encrypted record: %d bytes", len(record.Payload))
	}

	if record.ID == "" {
		record.ID = uuid.New().String()
	}
	if record.Version == 0 {
		record.Version = 1
	}

	now := time.Now().Unix()
	createdAt := record.CreatedAt.Unix()
	if createdAt == 0 {
		createdAt = now
		record.CreatedAt = time.Unix(createdAt, 0)
	}
	updatedAt := record.UpdatedAt.Unix()
	if updatedAt == 0 {
		updatedAt = now
		record.UpdatedAt = time.Unix(updatedAt, 0)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO records (
			id, vault_id, record_type, version, payload, created_at, updated_at, deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, NULL);
	`

	_, err = tx.ExecContext(
		ctx,
		query,
		record.ID,
		record.VaultID,
		record.RecordType,
		record.Version,
		record.Payload,
		createdAt,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}

	if len(tagIDs) > 0 {
		if err := r.tagRepo.SetRecordTagsTx(ctx, tx, record.ID, tagIDs); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByID returns an active record by ID.
func (r *RecordRepository) GetByID(ctx context.Context, vaultID, id string) (*Record, error) {
	query := `
		SELECT id, vault_id, record_type, version, payload, created_at, updated_at, deleted_at
		FROM records
		WHERE vault_id = ? AND id = ? AND deleted_at IS NULL;
	`
	return r.scanRecord(r.db.QueryRowContext(ctx, query, vaultID, id))
}

// List returns records for a vault, optionally filtering soft-deleted ones.
func (r *RecordRepository) List(ctx context.Context, vaultID string, includeDeleted bool) ([]Record, error) {
	query := `
		SELECT id, vault_id, record_type, version, payload, created_at, updated_at, deleted_at
		FROM records
		WHERE vault_id = ?
	`
	if !includeDeleted {
		query += " AND deleted_at IS NULL"
	}
	query += " ORDER BY updated_at DESC;"

	rows, err := r.db.QueryContext(ctx, query, vaultID)
	if err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	var list []Record
	for rows.Next() {
		var (
			rec           Record
			createdAtUnix int64
			updatedAtUnix int64
			deletedAtUnix sql.NullInt64
		)

		err := rows.Scan(
			&rec.ID,
			&rec.VaultID,
			&rec.RecordType,
			&rec.Version,
			&rec.Payload,
			&createdAtUnix,
			&updatedAtUnix,
			&deletedAtUnix,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}

		rec.CreatedAt = time.Unix(createdAtUnix, 0)
		rec.UpdatedAt = time.Unix(updatedAtUnix, 0)
		if deletedAtUnix.Valid {
			t := time.Unix(deletedAtUnix.Int64, 0)
			rec.DeletedAt = &t
		}
		list = append(list, rec)
	}

	return list, rows.Err()
}

// Update updates a record, archives the previous version in record_history, and updates tags.
func (r *RecordRepository) Update(ctx context.Context, record *Record, tagIDs []string) error {
	if record == nil {
		return errors.New("record cannot be nil")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin update transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Fetch current record for optimistic concurrency and history archiving
	var current Record
	var createdAtUnix, updatedAtUnix int64
	var deletedAtUnix sql.NullInt64

	checkQuery := `
		SELECT id, vault_id, record_type, version, payload, created_at, updated_at, deleted_at
		FROM records
		WHERE vault_id = ? AND id = ?;
	`
	err = tx.QueryRowContext(ctx, checkQuery, record.VaultID, record.ID).Scan(
		&current.ID,
		&current.VaultID,
		&current.RecordType,
		&current.Version,
		&current.Payload,
		&createdAtUnix,
		&updatedAtUnix,
		&deletedAtUnix,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrRecordNotFound
	}
	if err != nil {
		return fmt.Errorf("failed to load current record for update: %w", err)
	}

	if current.Version != record.Version {
		return ErrOptimisticLockingConflict
	}

	now := time.Now().Unix()

	// 2. Archive current version to record_history
	historyID := uuid.New().String()
	historyQuery := `
		INSERT INTO record_history (id, record_id, vault_id, version, payload, archived_at)
		VALUES (?, ?, ?, ?, ?, ?);
	`
	_, err = tx.ExecContext(ctx, historyQuery, historyID, current.ID, current.VaultID, current.Version, current.Payload, now)
	if err != nil {
		return fmt.Errorf("failed to archive record history: %w", err)
	}

	// 3. Update records table with incremented version and new payload
	newVersion := current.Version + 1
	updateQuery := `
		UPDATE records
		SET
			record_type = ?,
			version = ?,
			payload = ?,
			updated_at = ?
		WHERE vault_id = ? AND id = ? AND version = ?;
	`
	res, err := tx.ExecContext(
		ctx,
		updateQuery,
		record.RecordType,
		newVersion,
		record.Payload,
		now,
		record.VaultID,
		record.ID,
		current.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil || affected == 0 {
		return ErrOptimisticLockingConflict
	}

	// 4. Update tags
	if err := r.tagRepo.SetRecordTagsTx(ctx, tx, record.ID, tagIDs); err != nil {
		return err
	}

	record.Version = newVersion
	record.UpdatedAt = time.Unix(now, 0)
	return tx.Commit()
}

// SoftDelete marks a record as deleted.
func (r *RecordRepository) SoftDelete(ctx context.Context, vaultID, id string) error {
	now := time.Now().Unix()
	query := `UPDATE records SET deleted_at = ? WHERE vault_id = ? AND id = ? AND deleted_at IS NULL;`
	res, err := r.db.ExecContext(ctx, query, now, vaultID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// Restore restores a soft-deleted record.
func (r *RecordRepository) Restore(ctx context.Context, vaultID, id string) error {
	query := `UPDATE records SET deleted_at = NULL WHERE vault_id = ? AND id = ? AND deleted_at IS NOT NULL;`
	res, err := r.db.ExecContext(ctx, query, vaultID, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil || rows == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// PurgeDeleted permanently deletes all soft-deleted records and cascades to tags and history.
func (r *RecordRepository) PurgeDeleted(ctx context.Context, vaultID string) (int64, error) {
	query := `DELETE FROM records WHERE vault_id = ? AND deleted_at IS NOT NULL;`
	res, err := r.db.ExecContext(ctx, query, vaultID)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *RecordRepository) scanRecord(row *sql.Row) (*Record, error) {
	var (
		rec           Record
		createdAtUnix int64
		updatedAtUnix int64
		deletedAtUnix sql.NullInt64
	)

	err := row.Scan(
		&rec.ID,
		&rec.VaultID,
		&rec.RecordType,
		&rec.Version,
		&rec.Payload,
		&createdAtUnix,
		&updatedAtUnix,
		&deletedAtUnix,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan record: %w", err)
	}

	rec.CreatedAt = time.Unix(createdAtUnix, 0)
	rec.UpdatedAt = time.Unix(updatedAtUnix, 0)
	if deletedAtUnix.Valid {
		t := time.Unix(deletedAtUnix.Int64, 0)
		rec.DeletedAt = &t
	}

	return &rec, nil
}
