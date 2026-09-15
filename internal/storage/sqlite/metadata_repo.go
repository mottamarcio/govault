package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mottamarcio/govault/internal/crypto/kdf"
	"github.com/mottamarcio/govault/internal/crypto/keys"
)

var (
	ErrVaultNotInitialized = errors.New("vault is not initialized")
	ErrVaultAlreadyExists  = errors.New("vault metadata already exists: only one vault allowed per database")
)

// VaultMetadata represents the stored metadata and wrapped key material of a vault.
type VaultMetadata struct {
	VaultID       string
	SchemaVersion int
	CryptoSuite   string
	Envelope      *keys.WrappedKeyEnvelope
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// VaultHeaderInfo contains non-sensitive metadata for fast inspection.
type VaultHeaderInfo struct {
	VaultID       string
	SchemaVersion int
	CryptoSuite   string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// MetadataRepository manages database persistence for vault_metadata.
type MetadataRepository struct {
	db *DB
}

// NewMetadataRepository creates a new MetadataRepository instance.
func NewMetadataRepository(db *DB) *MetadataRepository {
	return &MetadataRepository{db: db}
}

// Create persists a new vault metadata row, ensuring single-vault invariant.
func (r *MetadataRepository) Create(ctx context.Context, meta *VaultMetadata) error {
	if meta == nil || meta.Envelope == nil {
		return errors.New("metadata and envelope cannot be nil")
	}
	if err := meta.Envelope.Validate(); err != nil {
		return fmt.Errorf("invalid envelope: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Check single-vault constraint
	var count int
	if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM vault_metadata;").Scan(&count); err != nil {
		return fmt.Errorf("failed to check existing vault metadata: %w", err)
	}
	if count > 0 {
		return ErrVaultAlreadyExists
	}

	// 2. Serialize KDF params
	kdfJSON, err := meta.Envelope.KDFParams.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal KDF params: %w", err)
	}

	now := time.Now().Unix()
	createdAt := meta.CreatedAt.Unix()
	if createdAt == 0 {
		createdAt = now
	}
	updatedAt := meta.UpdatedAt.Unix()
	if updatedAt == 0 {
		updatedAt = now
	}

	query := `
		INSERT INTO vault_metadata (
			vault_id,
			schema_version,
			crypto_suite,
			kdf_params_json,
			wrapped_key,
			mck,
			created_at,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`

	_, err = tx.ExecContext(
		ctx,
		query,
		meta.VaultID,
		meta.SchemaVersion,
		meta.CryptoSuite,
		string(kdfJSON),
		meta.Envelope.WrappedKey,
		meta.Envelope.MCK,
		createdAt,
		updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert vault metadata: %w", err)
	}

	return tx.Commit()
}

// Get queries the full VaultMetadata row.
func (r *MetadataRepository) Get(ctx context.Context) (*VaultMetadata, error) {
	query := `
		SELECT
			vault_id,
			schema_version,
			crypto_suite,
			kdf_params_json,
			wrapped_key,
			mck,
			created_at,
			updated_at
		FROM vault_metadata
		LIMIT 1;
	`

	var (
		vaultID       string
		schemaVersion int
		cryptoSuite   string
		kdfParamsJSON string
		wrappedKey    []byte
		mck           []byte
		createdAtUnix int64
		updatedAtUnix int64
	)

	err := r.db.QueryRowContext(ctx, query).Scan(
		&vaultID,
		&schemaVersion,
		&cryptoSuite,
		&kdfParamsJSON,
		&wrappedKey,
		&mck,
		&createdAtUnix,
		&updatedAtUnix,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVaultNotInitialized
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query vault metadata: %w", err)
	}

	kdfParams, err := kdf.UnmarshalParams([]byte(kdfParamsJSON))
	if err != nil {
		return nil, fmt.Errorf("corrupt KDF params in vault metadata: %w", err)
	}

	envelope := &keys.WrappedKeyEnvelope{
		KDFParams:  kdfParams,
		WrappedKey: wrappedKey,
		MCK:        mck,
	}

	return &VaultMetadata{
		VaultID:       vaultID,
		SchemaVersion: schemaVersion,
		CryptoSuite:   cryptoSuite,
		Envelope:      envelope,
		CreatedAt:     time.Unix(createdAtUnix, 0),
		UpdatedAt:     time.Unix(updatedAtUnix, 0),
	}, nil
}

// Inspect returns header info without unmarshaling the sensitive envelope.
func (r *MetadataRepository) Inspect(ctx context.Context) (*VaultHeaderInfo, error) {
	query := `
		SELECT
			vault_id,
			schema_version,
			crypto_suite,
			created_at,
			updated_at
		FROM vault_metadata
		LIMIT 1;
	`

	var (
		vaultID       string
		schemaVersion int
		cryptoSuite   string
		createdAtUnix int64
		updatedAtUnix int64
	)

	err := r.db.QueryRowContext(ctx, query).Scan(
		&vaultID,
		&schemaVersion,
		&cryptoSuite,
		&createdAtUnix,
		&updatedAtUnix,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrVaultNotInitialized
	}
	if err != nil {
		return nil, fmt.Errorf("failed to inspect vault metadata: %w", err)
	}

	return &VaultHeaderInfo{
		VaultID:       vaultID,
		SchemaVersion: schemaVersion,
		CryptoSuite:   cryptoSuite,
		CreatedAt:     time.Unix(createdAtUnix, 0),
		UpdatedAt:     time.Unix(updatedAtUnix, 0),
	}, nil
}

// UpdateEnvelope updates the WrappedKeyEnvelope and updated_at timestamp atomically.
func (r *MetadataRepository) UpdateEnvelope(ctx context.Context, vaultID string, env *keys.WrappedKeyEnvelope) error {
	if env == nil {
		return errors.New("envelope cannot be nil")
	}
	if err := env.Validate(); err != nil {
		return fmt.Errorf("invalid envelope: %w", err)
	}

	kdfJSON, err := env.KDFParams.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal KDF params: %w", err)
	}

	now := time.Now().Unix()
	query := `
		UPDATE vault_metadata
		SET
			kdf_params_json = ?,
			wrapped_key = ?,
			mck = ?,
			updated_at = ?
		WHERE vault_id = ?;
	`

	res, err := r.db.ExecContext(ctx, query, string(kdfJSON), env.WrappedKey, env.MCK, now, vaultID)
	if err != nil {
		return fmt.Errorf("failed to update vault metadata: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: vault ID %s not found", ErrVaultNotInitialized, vaultID)
	}

	return nil
}
