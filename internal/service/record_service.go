package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mottamarcio/govault/internal/crypto/cipher"
	"github.com/mottamarcio/govault/internal/crypto/kdf"
	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
)

var (
	ErrDecryptionFailed = errors.New("service: record decryption or authentication failed")
)

// RecordInput represents input parameters for creating or updating a record.
type RecordInput struct {
	ID      string
	Title   string
	Type    domain.RecordType
	Tags    []string
	Payload any
	Version uint32
}

// RecordService orchestrates record encryption, storage, retrieval, history, trash, and search.
type RecordService struct {
	session     *Session
	recordRepo  *sqlite.RecordRepository
	tagRepo     *sqlite.TagRepository
	historyRepo *sqlite.HistoryRepository
}

// NewRecordService creates a new RecordService instance.
func NewRecordService(
	session *Session,
	recordRepo *sqlite.RecordRepository,
	tagRepo *sqlite.TagRepository,
	historyRepo *sqlite.HistoryRepository,
) *RecordService {
	return &RecordService{
		session:     session,
		recordRepo:  recordRepo,
		tagRepo:     tagRepo,
		historyRepo: historyRepo,
	}
}

// deriveRecordKey derives a record-specific 32-byte key from the decrypted VaultKey.
func (s *RecordService) deriveRecordKey(vaultKey []byte, recordID string) ([]byte, error) {
	info := fmt.Sprintf("govault/v1/record/%s", recordID)
	return kdf.DeriveSubKey(vaultKey, info, 32)
}

// Create encrypts and persists a new domain secret record.
func (s *RecordService) Create(ctx context.Context, input RecordInput) (*domain.Record, error) {
	vaultKey, err := s.session.VaultKey()
	if err != nil {
		return nil, ErrVaultLocked
	}
	defer kdf.Zeroize(vaultKey)

	vaultID, err := s.session.VaultID()
	if err != nil {
		return nil, ErrVaultLocked
	}

	recordID := strings.TrimSpace(input.ID)
	if recordID == "" {
		recordID = uuid.New().String()
	}

	// 1. Serialize typed payload
	serializedPayload, err := domain.SerializePayload(input.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload: %w", err)
	}

	// 2. Derive record subkey
	kRecord, err := s.deriveRecordKey(vaultKey, recordID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive record key: %w", err)
	}
	defer kdf.Zeroize(kRecord)

	// 3. Encrypt payload with authenticated AAD
	aad := cipher.AADContext{
		VaultID:    vaultID,
		RecordID:   recordID,
		RecordType: input.Type.String(),
		Version:    1,
	}

	ciphertextPayload, err := cipher.Encrypt(kRecord, serializedPayload, aad.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt payload: %w", err)
	}

	// 4. Resolve tag IDs (create missing tags if necessary)
	tagIDs, err := s.resolveOrCreateTags(ctx, vaultID, input.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tags: %w", err)
	}

	now := time.Now()
	storageRecord := &sqlite.Record{
		ID:         recordID,
		VaultID:    vaultID,
		RecordType: input.Type.String(),
		Version:    1,
		Payload:    ciphertextPayload,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.recordRepo.Create(ctx, storageRecord, tagIDs); err != nil {
		return nil, fmt.Errorf("failed to save record: %w", err)
	}

	// 5. Construct domain record
	domainRecord := &domain.Record{
		ID:        recordID,
		VaultID:   vaultID,
		Title:     input.Title,
		Type:      input.Type,
		Tags:      input.Tags,
		Version:   1,
		Payload:   input.Payload,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return domainRecord, nil
}

// GetByID fetches and decrypts a secret record by ID.
func (s *RecordService) GetByID(ctx context.Context, id string) (*domain.Record, error) {
	vaultKey, err := s.session.VaultKey()
	if err != nil {
		return nil, ErrVaultLocked
	}
	defer kdf.Zeroize(vaultKey)

	vaultID, err := s.session.VaultID()
	if err != nil {
		return nil, ErrVaultLocked
	}

	storageRecord, err := s.recordRepo.GetByID(ctx, vaultID, id)
	if err != nil {
		return nil, err
	}

	return s.decryptRecord(ctx, vaultKey, storageRecord)
}

// List returns decrypted records (active or including trash).
func (s *RecordService) List(ctx context.Context, includeDeleted bool) ([]*domain.Record, error) {
	vaultKey, err := s.session.VaultKey()
	if err != nil {
		return nil, ErrVaultLocked
	}
	defer kdf.Zeroize(vaultKey)

	vaultID, err := s.session.VaultID()
	if err != nil {
		return nil, ErrVaultLocked
	}

	records, err := s.recordRepo.List(ctx, vaultID, includeDeleted)
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}

	var domainRecords []*domain.Record
	for _, rec := range records {
		recCopy := rec
		dRec, err := s.decryptRecord(ctx, vaultKey, &recCopy)
		if err != nil {
			return nil, err
		}
		domainRecords = append(domainRecords, dRec)
	}

	return domainRecords, nil
}

// Update updates an existing record, archiving previous revision into history.
func (s *RecordService) Update(ctx context.Context, id string, input RecordInput) (*domain.Record, error) {
	vaultKey, err := s.session.VaultKey()
	if err != nil {
		return nil, ErrVaultLocked
	}
	defer kdf.Zeroize(vaultKey)

	vaultID, err := s.session.VaultID()
	if err != nil {
		return nil, ErrVaultLocked
	}

	current, err := s.recordRepo.GetByID(ctx, vaultID, id)
	if err != nil {
		return nil, err
	}

	nextVersion := current.Version + 1

	// 1. Serialize new payload
	serializedPayload, err := domain.SerializePayload(input.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize payload: %w", err)
	}

	// 2. Derive record subkey
	kRecord, err := s.deriveRecordKey(vaultKey, id)
	if err != nil {
		return nil, fmt.Errorf("failed to derive record key: %w", err)
	}
	defer kdf.Zeroize(kRecord)

	// 3. Encrypt payload with new AADContext (version nextVersion)
	aad := cipher.AADContext{
		VaultID:    vaultID,
		RecordID:   id,
		RecordType: input.Type.String(),
		Version:    nextVersion,
	}

	ciphertextPayload, err := cipher.Encrypt(kRecord, serializedPayload, aad.Bytes())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt payload: %w", err)
	}

	// 4. Resolve tag IDs
	tagIDs, err := s.resolveOrCreateTags(ctx, vaultID, input.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tags: %w", err)
	}

	now := time.Now()
	storageRecord := &sqlite.Record{
		ID:         id,
		VaultID:    vaultID,
		RecordType: input.Type.String(),
		Version:    current.Version, // Expected version for optimistic concurrency
		Payload:    ciphertextPayload,
		UpdatedAt:  now,
	}

	if err := s.recordRepo.Update(ctx, storageRecord, tagIDs); err != nil {
		return nil, err
	}

	return &domain.Record{
		ID:        id,
		VaultID:   vaultID,
		Title:     input.Title,
		Type:      input.Type,
		Tags:      input.Tags,
		Version:   nextVersion,
		Payload:   input.Payload,
		CreatedAt: current.CreatedAt,
		UpdatedAt: now,
	}, nil
}

// Delete soft-deletes a record.
func (s *RecordService) Delete(ctx context.Context, id string) error {
	vaultID, err := s.session.VaultID()
	if err != nil {
		return ErrVaultLocked
	}
	return s.recordRepo.SoftDelete(ctx, vaultID, id)
}

// Restore restores a soft-deleted record.
func (s *RecordService) Restore(ctx context.Context, id string) error {
	vaultID, err := s.session.VaultID()
	if err != nil {
		return ErrVaultLocked
	}
	return s.recordRepo.Restore(ctx, vaultID, id)
}

// ListTrash returns all soft-deleted records.
func (s *RecordService) ListTrash(ctx context.Context) ([]*domain.Record, error) {
	all, err := s.List(ctx, true)
	if err != nil {
		return nil, err
	}

	var trash []*domain.Record
	for _, r := range all {
		if r.IsDeleted() {
			trash = append(trash, r)
		}
	}
	return trash, nil
}

// PurgeTrash permanently removes all soft-deleted records and their history.
func (s *RecordService) PurgeTrash(ctx context.Context) error {
	vaultID, err := s.session.VaultID()
	if err != nil {
		return ErrVaultLocked
	}
	_, err = s.recordRepo.PurgeDeleted(ctx, vaultID)
	return err
}

// ListHistory returns metadata for historical revisions of a record.
func (s *RecordService) ListHistory(ctx context.Context, recordID string) ([]*domain.HistoryEntry, error) {
	vaultKey, err := s.session.VaultKey()
	if err != nil {
		return nil, ErrVaultLocked
	}
	defer kdf.Zeroize(vaultKey)

	vaultID, err := s.session.VaultID()
	if err != nil {
		return nil, ErrVaultLocked
	}

	entries, err := s.historyRepo.ListByRecordID(ctx, recordID)
	if err != nil {
		return nil, err
	}

	kRecord, err := s.deriveRecordKey(vaultKey, recordID)
	if err != nil {
		return nil, err
	}
	defer kdf.Zeroize(kRecord)

	var domainEntries []*domain.HistoryEntry
	for _, entry := range entries {
		// Decrypt payload
		aad := cipher.AADContext{
			VaultID:    vaultID,
			RecordID:   recordID,
			RecordType: "",
			Version:    entry.Version,
		}
		// Fetch record type from current record
		rec, _ := s.recordRepo.GetByID(ctx, vaultID, recordID)
		if rec != nil {
			aad.RecordType = rec.RecordType
		}

		plaintext, err := cipher.Decrypt(kRecord, entry.Payload, aad.Bytes())
		if err != nil {
			return nil, ErrDecryptionFailed
		}

		var typedPayload any
		if rec != nil {
			typedPayload, _ = domain.DeserializePayload(domain.RecordType(rec.RecordType), plaintext)
		}

		domainEntries = append(domainEntries, &domain.HistoryEntry{
			ID:         entry.ID,
			RecordID:   entry.RecordID,
			VaultID:    entry.VaultID,
			Version:    entry.Version,
			Payload:    typedPayload,
			ArchivedAt: entry.ArchivedAt,
		})
	}

	return domainEntries, nil
}

// GetHistoryRevision fetches and decrypts a specific historical revision.
func (s *RecordService) GetHistoryRevision(ctx context.Context, recordID string, version uint32) (*domain.Record, error) {
	vaultKey, err := s.session.VaultKey()
	if err != nil {
		return nil, ErrVaultLocked
	}
	defer kdf.Zeroize(vaultKey)

	vaultID, err := s.session.VaultID()
	if err != nil {
		return nil, ErrVaultLocked
	}

	entry, err := s.historyRepo.GetVersion(ctx, recordID, version)
	if err != nil {
		return nil, err
	}

	current, err := s.recordRepo.GetByID(ctx, vaultID, recordID)
	if err != nil {
		return nil, err
	}

	kRecord, err := s.deriveRecordKey(vaultKey, recordID)
	if err != nil {
		return nil, err
	}
	defer kdf.Zeroize(kRecord)

	aad := cipher.AADContext{
		VaultID:    vaultID,
		RecordID:   recordID,
		RecordType: current.RecordType,
		Version:    version,
	}

	plaintext, err := cipher.Decrypt(kRecord, entry.Payload, aad.Bytes())
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	typedPayload, err := domain.DeserializePayload(domain.RecordType(current.RecordType), plaintext)
	if err != nil {
		return nil, err
	}

	return &domain.Record{
		ID:        recordID,
		VaultID:   vaultID,
		Type:      domain.RecordType(current.RecordType),
		Version:   version,
		Payload:   typedPayload,
		CreatedAt: current.CreatedAt,
		UpdatedAt: entry.ArchivedAt,
	}, nil
}

// Search performs in-memory searching and ranking across decrypted records.
func (s *RecordService) Search(ctx context.Context, filter SearchFilter) ([]*domain.Record, error) {
	all, err := s.List(ctx, filter.IncludeTrash)
	if err != nil {
		return nil, err
	}

	return FilterAndRankRecords(all, filter), nil
}

func (s *RecordService) decryptRecord(ctx context.Context, vaultKey []byte, storageRecord *sqlite.Record) (*domain.Record, error) {
	kRecord, err := s.deriveRecordKey(vaultKey, storageRecord.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to derive record key: %w", err)
	}
	defer kdf.Zeroize(kRecord)

	aad := cipher.AADContext{
		VaultID:    storageRecord.VaultID,
		RecordID:   storageRecord.ID,
		RecordType: storageRecord.RecordType,
		Version:    storageRecord.Version,
	}

	plaintext, err := cipher.Decrypt(kRecord, storageRecord.Payload, aad.Bytes())
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	typedPayload, err := domain.DeserializePayload(domain.RecordType(storageRecord.RecordType), plaintext)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize payload: %w", err)
	}

	// Fetch tag IDs and resolve tag names
	tagIDs, err := s.tagRepo.GetRecordTags(ctx, storageRecord.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch record tags: %w", err)
	}

	allVaultTags, err := s.tagRepo.List(ctx, storageRecord.VaultID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch vault tags: %w", err)
	}

	tagMap := make(map[string]string, len(allVaultTags))
	for _, t := range allVaultTags {
		tagMap[t.ID] = t.Name
	}

	var tagNames []string
	for _, tid := range tagIDs {
		if name, exists := tagMap[tid]; exists {
			tagNames = append(tagNames, name)
		}
	}

	// Title inference
	title := storageRecord.ID
	switch p := typedPayload.(type) {
	case domain.LoginPayload:
		if p.URI != "" {
			title = p.URI
		} else if p.Username != "" {
			title = p.Username
		}
	case domain.NotePayload:
		lines := strings.Split(strings.TrimSpace(p.Content), "\n")
		if len(lines) > 0 && lines[0] != "" {
			title = lines[0]
		}
	case domain.APIKeyPayload:
		if p.Service != "" {
			title = p.Service
		}
	case domain.CustomPayload:
		if p.Notes != "" {
			title = p.Notes
		}
	}

	return &domain.Record{
		ID:        storageRecord.ID,
		VaultID:   storageRecord.VaultID,
		Title:     title,
		Type:      domain.RecordType(storageRecord.RecordType),
		Tags:      tagNames,
		Version:   storageRecord.Version,
		Payload:   typedPayload,
		CreatedAt: storageRecord.CreatedAt,
		UpdatedAt: storageRecord.UpdatedAt,
		DeletedAt: storageRecord.DeletedAt,
	}, nil
}

func (s *RecordService) resolveOrCreateTags(ctx context.Context, vaultID string, tagNames []string) ([]string, error) {
	if len(tagNames) == 0 {
		return nil, nil
	}

	existingTags, err := s.tagRepo.List(ctx, vaultID)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string]string, len(existingTags))
	for _, t := range existingTags {
		tagMap[strings.ToLower(t.Name)] = t.ID
	}

	var tagIDs []string
	for _, name := range tagNames {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if id, exists := tagMap[lower]; exists {
			tagIDs = append(tagIDs, id)
		} else {
			// Create tag
			tag, err := s.tagRepo.Create(ctx, vaultID, trimmed)
			if err != nil {
				return nil, err
			}
			tagMap[lower] = tag.ID
			tagIDs = append(tagIDs, tag.ID)
		}
	}

	return tagIDs, nil
}
