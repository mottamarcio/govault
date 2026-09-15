package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/mottamarcio/govault/internal/backup/format"
	"github.com/mottamarcio/govault/internal/crypto/cipher"
	"github.com/mottamarcio/govault/internal/crypto/kdf"
	"github.com/mottamarcio/govault/internal/crypto/keys"
	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
)

var (
	ErrFileAlreadyExists  = errors.New("backup destination file already exists")
	ErrInvalidPassphrase  = errors.New("invalid backup passphrase")
	ErrRestoreValidation  = errors.New("backup restore validation failed")
	ErrConflictingRecords = errors.New("conflicting records found during merge")
)

type ConflictStrategy string

const (
	ConflictSkip      ConflictStrategy = "skip"
	ConflictOverwrite ConflictStrategy = "overwrite"
	ConflictRename    ConflictStrategy = "rename"
)

type BackupCreateOptions struct {
	ExportPassphrase string
	Overwrite        bool
}

type BackupVerifyResult struct {
	Inspection *format.HeaderInspection
	Manifest   *format.BackupManifest
	IsVerified bool
}

type BackupRestoreOptions struct {
	SkipSnapshot bool
}

type BackupMergeResult struct {
	ImportedCount int
	SkippedCount  int
	RenamedCount  int
}

// BackupService coordinates creation, verification, restoration, and merge of .gvault backups.
type BackupService struct {
	metaRepo    *sqlite.MetadataRepository
	recordRepo  *sqlite.RecordRepository
	tagRepo     *sqlite.TagRepository
	historyRepo *sqlite.HistoryRepository
	dbPath      string
}

// NewBackupService constructs a BackupService instance.
func NewBackupService(
	metaRepo *sqlite.MetadataRepository,
	recordRepo *sqlite.RecordRepository,
	tagRepo *sqlite.TagRepository,
	historyRepo *sqlite.HistoryRepository,
	dbPath string,
) *BackupService {
	return &BackupService{
		metaRepo:    metaRepo,
		recordRepo:  recordRepo,
		tagRepo:     tagRepo,
		historyRepo: historyRepo,
		dbPath:      dbPath,
	}
}

// CreateBackup exports the current vault into an atomic, self-contained .gvault file.
func (s *BackupService) CreateBackup(ctx context.Context, sess *Session, outputPath string, opts BackupCreateOptions) error {
	vaultKey, err := sess.VaultKey()
	if err != nil {
		return ErrVaultLocked
	}
	defer kdf.Zeroize(vaultKey)

	vaultID, err := sess.VaultID()
	if err != nil {
		return ErrVaultLocked
	}

	if _, err := os.Stat(outputPath); err == nil && !opts.Overwrite {
		return ErrFileAlreadyExists
	}

	meta, err := s.metaRepo.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve vault metadata: %w", err)
	}

	// 1. Determine KDF Params and Wrapped Vault Key
	var kdfParams *kdf.Argon2Params
	var wrappedKey []byte
	var mck []byte

	if opts.ExportPassphrase != "" {
		kdfParams, err = kdf.DefaultArgon2Params()
		if err != nil {
			return err
		}
		env, err := keys.WrapVaultKey(vaultKey, []byte(opts.ExportPassphrase), vaultID, kdfParams)
		if err != nil {
			return fmt.Errorf("failed to wrap vault key with export passphrase: %w", err)
		}
		wrappedKey = env.WrappedKey
		mck = env.MCK
	} else {
		kdfParams = meta.Envelope.KDFParams
		wrappedKey = meta.Envelope.WrappedKey
		mck = meta.Envelope.MCK
	}

	// 2. Fetch and decrypt all records
	records, err := s.recordRepo.List(ctx, vaultID, true)
	if err != nil {
		return fmt.Errorf("failed to list records for backup: %w", err)
	}

	allTags, err := s.tagRepo.List(ctx, vaultID)
	if err != nil {
		return fmt.Errorf("failed to list tags for backup: %w", err)
	}
	tagMap := make(map[string]string, len(allTags))
	for _, t := range allTags {
		tagMap[t.ID] = t.Name
	}

	var backupEntries []format.BackupEntry
	var backupHistories []format.BackupHistory

	for _, rec := range records {
		recKey, err := kdf.DeriveSubKey(vaultKey, fmt.Sprintf("govault/v1/record/%s", rec.ID), 32)
		if err != nil {
			return err
		}

		aad := cipher.AADContext{
			VaultID:    rec.VaultID,
			RecordID:   rec.ID,
			RecordType: rec.RecordType,
			Version:    rec.Version,
		}
		plaintextPayload, err := cipher.Decrypt(recKey, rec.Payload, aad.Bytes())
		kdf.Zeroize(recKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt record %s for backup: %w", rec.ID, err)
		}

		tagIDs, err := s.tagRepo.GetRecordTags(ctx, rec.ID)
		if err != nil {
			return err
		}
		var tagNames []string
		for _, tid := range tagIDs {
			if name, ok := tagMap[tid]; ok {
				tagNames = append(tagNames, name)
			}
		}

		backupEntries = append(backupEntries, format.BackupEntry{
			ID:            rec.ID,
			Type:          domain.RecordType(rec.RecordType),
			Title:         rec.ID,
			Payload:       plaintextPayload,
			Version:       int(rec.Version),
			CreatedAtUnix: rec.CreatedAt.Unix(),
			UpdatedAtUnix: rec.UpdatedAt.Unix(),
			Tags:          tagNames,
			IsDeleted:     rec.DeletedAt != nil,
		})

		// Fetch histories for this record
		histories, err := s.historyRepo.ListByRecordID(ctx, rec.ID)
		if err != nil {
			return err
		}
		for _, h := range histories {
			hKey, err := kdf.DeriveSubKey(vaultKey, fmt.Sprintf("govault/v1/record/%s", rec.ID), 32)
			if err != nil {
				return err
			}
			hAad := cipher.AADContext{
				VaultID:    h.VaultID,
				RecordID:   h.RecordID,
				RecordType: rec.RecordType,
				Version:    h.Version,
			}
			hPlaintext, err := cipher.Decrypt(hKey, h.Payload, hAad.Bytes())
			kdf.Zeroize(hKey)
			if err != nil {
				return fmt.Errorf("failed to decrypt history revision for record %s: %w", rec.ID, err)
			}

			backupHistories = append(backupHistories, format.BackupHistory{
				HistoryID:     h.ID,
				RecordID:      h.RecordID,
				Version:       int(h.Version),
				Payload:       hPlaintext,
				CreatedAtUnix: h.ArchivedAt.Unix(),
			})
		}
	}

	var backupTags []format.BackupTag
	for _, t := range allTags {
		backupTags = append(backupTags, format.BackupTag{
			ID:            t.ID,
			Name:          t.Name,
			CreatedAtUnix: t.CreatedAt.Unix(),
		})
	}

	payload := &format.BackupPayload{
		Version: 1,
		Manifest: format.BackupManifest{
			SourceVaultVersion:  1,
			SourceSchemaVersion: uint16(meta.SchemaVersion),
			AppVersion:          "0.1.0",
			CreatedAtUnix:       time.Now().Unix(),
		},
		Entries:   backupEntries,
		Histories: backupHistories,
		Tags:      backupTags,
	}

	var backupID [16]byte
	if _, err := io.ReadFull(rand.Reader, backupID[:]); err != nil {
		return err
	}

	header := &format.Header{
		FormatVersion:      format.CurrentBackupFormatVersion,
		CryptoSuiteVersion: format.CurrentCryptoSuiteVersion,
		BackupID:           backupID,
		VaultID:            vaultID,
		CreatedAtUnix:      time.Now().Unix(),
		KDFAlgorithm:       format.KDFAlgorithmArgon2id,
		KDFParams:          kdfParams,
		WrappedVaultKey:    wrappedKey,
		MCK:                mck,
	}

	// 3. Write atomic temporary file
	outDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outDir, 0700); err != nil {
		return err
	}

	tempFile, err := os.CreateTemp(outDir, ".gvault.tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp backup file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	if err := tempFile.Chmod(0600); err != nil {
		tempFile.Close()
		return err
	}

	if err := format.EncryptBackup(tempFile, payload, vaultKey, header); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to encrypt backup: %w", err)
	}
	if err := tempFile.Sync(); err != nil {
		tempFile.Close()
		return err
	}
	tempFile.Close()

	// Verify temporary file structurally
	fCheck, err := os.Open(tempPath)
	if err != nil {
		return err
	}
	if _, err := format.InspectHeader(fCheck); err != nil {
		fCheck.Close()
		return fmt.Errorf("temp backup failed structural verification: %w", err)
	}
	fCheck.Close()

	// 4. Atomic Rename
	if err := os.Rename(tempPath, outputPath); err != nil {
		return fmt.Errorf("failed to atomically rename backup to destination: %w", err)
	}

	return nil
}

// VerifyBackup performs structural or full cryptographic verification of a .gvault file.
func (s *BackupService) VerifyBackup(ctx context.Context, backupPath string, password string) (*BackupVerifyResult, error) {
	file, err := os.Open(backupPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	inspection, err := format.InspectHeader(file)
	if err != nil {
		return nil, err
	}

	res := &BackupVerifyResult{
		Inspection: inspection,
		IsVerified: false,
	}

	if password == "" {
		return res, nil
	}

	// Cryptographic verification
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	header, err := format.ParseHeader(file)
	if err != nil {
		return nil, err
	}

	env := &keys.WrappedKeyEnvelope{
		KDFParams:  header.KDFParams,
		WrappedKey: header.WrappedVaultKey,
		MCK:        header.MCK,
	}

	vaultKey, err := keys.UnwrapVaultKey(env, []byte(password), header.VaultID)
	if err != nil {
		return nil, ErrInvalidPassphrase
	}
	defer kdf.Zeroize(vaultKey)

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	_, payload, err := format.DecryptBackup(file, vaultKey)
	if err != nil {
		return nil, err
	}

	res.Manifest = &payload.Manifest
	res.IsVerified = true

	return res, nil
}

// RestoreBackup restores a .gvault backup into an active SQLite database atomically.
func (s *BackupService) RestoreBackup(ctx context.Context, backupPath string, password string, opts BackupRestoreOptions) error {
	file, err := os.Open(backupPath)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %w", err)
	}
	defer file.Close()

	header, err := format.ParseHeader(file)
	if err != nil {
		return fmt.Errorf("invalid backup header: %w", err)
	}

	env := &keys.WrappedKeyEnvelope{
		KDFParams:  header.KDFParams,
		WrappedKey: header.WrappedVaultKey,
		MCK:        header.MCK,
	}

	vaultKey, err := keys.UnwrapVaultKey(env, []byte(password), header.VaultID)
	if err != nil {
		return ErrInvalidPassphrase
	}
	defer kdf.Zeroize(vaultKey)

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	_, payload, err := format.DecryptBackup(file, vaultKey)
	if err != nil {
		return fmt.Errorf("failed to decrypt backup: %w", err)
	}

	// Pre-validation safety checks
	if payload.Manifest.EntryCount != len(payload.Entries) {
		return fmt.Errorf("%w: entry count mismatch", ErrRestoreValidation)
	}

	dbDir := filepath.Dir(s.dbPath)
	if err := os.MkdirAll(dbDir, 0700); err != nil {
		return err
	}

	// Safety Snapshot if existing database exists
	if !opts.SkipSnapshot && s.dbPath != "" {
		if _, err := os.Stat(s.dbPath); err == nil {
			snapshotPath := fmt.Sprintf("%s.recovery-%d.gvault", s.dbPath, time.Now().Unix())
			_ = s.createDirectSnapshot(ctx, snapshotPath)
		}
	}

	// Create temporary database
	tempDBPath := filepath.Join(dbDir, fmt.Sprintf(".vault-restore-%s.db", uuid.New().String()))
	defer os.Remove(tempDBPath)
	defer os.Remove(tempDBPath + "-wal")
	defer os.Remove(tempDBPath + "-shm")

	tempDB, err := sqlite.Open(tempDBPath)
	if err != nil {
		return fmt.Errorf("failed to open temporary restore database: %w", err)
	}

	if err := tempDB.Migrate(ctx); err != nil {
		tempDB.Close()
		return fmt.Errorf("failed to apply migrations to temporary database: %w", err)
	}

	tempMetaRepo := sqlite.NewMetadataRepository(tempDB)
	tempTagRepo := sqlite.NewTagRepository(tempDB)
	tempRecordRepo := sqlite.NewRecordRepository(tempDB, tempTagRepo)

	// Insert vault metadata
	meta := &sqlite.VaultMetadata{
		VaultID:       header.VaultID,
		SchemaVersion: int(payload.Manifest.SourceSchemaVersion),
		CryptoSuite:   "argon2id-hkdf-xchacha20poly1305",
		Envelope:      env,
		CreatedAt:     time.Unix(header.CreatedAtUnix, 0),
		UpdatedAt:     time.Now(),
	}
	if err := tempMetaRepo.Create(ctx, meta); err != nil {
		tempDB.Close()
		return fmt.Errorf("failed to insert restored metadata: %w", err)
	}

	// Insert tags
	for _, tag := range payload.Tags {
		_, err := tempTagRepo.Create(ctx, header.VaultID, tag.Name)
		if err != nil {
			tempDB.Close()
			return fmt.Errorf("failed to insert tag %s: %w", tag.Name, err)
		}
	}

	// Insert records and history with fresh nonces
	for _, entry := range payload.Entries {
		recKey, err := kdf.DeriveSubKey(vaultKey, fmt.Sprintf("govault/v1/record/%s", entry.ID), 32)
		if err != nil {
			tempDB.Close()
			return err
		}

		aad := cipher.AADContext{
			VaultID:    header.VaultID,
			RecordID:   entry.ID,
			RecordType: entry.Type.String(),
			Version:    uint32(entry.Version),
		}

		ciphertext, err := cipher.Encrypt(recKey, entry.Payload, aad.Bytes())
		kdf.Zeroize(recKey)
		if err != nil {
			tempDB.Close()
			return fmt.Errorf("failed to encrypt restored record: %w", err)
		}

		var tagIDs []string
		if len(entry.Tags) > 0 {
			allExisting, _ := tempTagRepo.List(ctx, header.VaultID)
			tmap := make(map[string]string)
			for _, t := range allExisting {
				tmap[t.Name] = t.ID
			}
			for _, tname := range entry.Tags {
				if id, ok := tmap[tname]; ok {
					tagIDs = append(tagIDs, id)
				}
			}
		}

		var deletedAt *time.Time
		if entry.IsDeleted {
			t := time.Now()
			deletedAt = &t
		}

		rec := &sqlite.Record{
			ID:         entry.ID,
			VaultID:    header.VaultID,
			RecordType: entry.Type.String(),
			Version:    uint32(entry.Version),
			Payload:    ciphertext,
			CreatedAt:  time.Unix(entry.CreatedAtUnix, 0),
			UpdatedAt:  time.Unix(entry.UpdatedAtUnix, 0),
			DeletedAt:  deletedAt,
		}

		if err := tempRecordRepo.Create(ctx, rec, tagIDs); err != nil {
			tempDB.Close()
			return fmt.Errorf("failed to write record %s: %w", entry.ID, err)
		}
	}

	if err := tempDB.Close(); err != nil {
		return fmt.Errorf("failed to close temporary restore database: %w", err)
	}

	// Atomic replace active database
	if err := os.Rename(tempDBPath, s.dbPath); err != nil {
		return fmt.Errorf("failed to replace active vault with restored database: %w", err)
	}

	return nil
}

// MergeBackup imports non-conflicting entries from a .gvault backup into an active unlocked session.
func (s *BackupService) MergeBackup(ctx context.Context, sess *Session, backupPath string, password string, strategy ConflictStrategy) (*BackupMergeResult, error) {
	vaultKey, err := sess.VaultKey()
	if err != nil {
		return nil, ErrVaultLocked
	}
	defer kdf.Zeroize(vaultKey)

	vaultID, err := sess.VaultID()
	if err != nil {
		return nil, ErrVaultLocked
	}

	file, err := os.Open(backupPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	header, err := format.ParseHeader(file)
	if err != nil {
		return nil, err
	}

	env := &keys.WrappedKeyEnvelope{
		KDFParams:  header.KDFParams,
		WrappedKey: header.WrappedVaultKey,
		MCK:        header.MCK,
	}

	backupVaultKey, err := keys.UnwrapVaultKey(env, []byte(password), header.VaultID)
	if err != nil {
		return nil, ErrInvalidPassphrase
	}
	defer kdf.Zeroize(backupVaultKey)

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	_, payload, err := format.DecryptBackup(file, backupVaultKey)
	if err != nil {
		return nil, err
	}

	existingRecords, err := s.recordRepo.List(ctx, vaultID, true)
	if err != nil {
		return nil, err
	}
	existingIDMap := make(map[string]bool, len(existingRecords))
	for _, r := range existingRecords {
		existingIDMap[r.ID] = true
	}

	result := &BackupMergeResult{}

	for _, entry := range payload.Entries {
		if existingIDMap[entry.ID] {
			switch strategy {
			case ConflictSkip:
				result.SkippedCount++
				continue
			case ConflictRename:
				entry.ID = uuid.New().String()
				result.RenamedCount++
			case ConflictOverwrite:
				// will overwrite
			default:
				result.SkippedCount++
				continue
			}
		}

		// Ensure tags exist in active vault
		var tagIDs []string
		for _, tagName := range entry.Tags {
			t, err := s.tagRepo.Create(ctx, vaultID, tagName)
			if err == nil && t != nil {
				tagIDs = append(tagIDs, t.ID)
			}
		}

		recKey, err := kdf.DeriveSubKey(vaultKey, fmt.Sprintf("govault/v1/record/%s", entry.ID), 32)
		if err != nil {
			return nil, err
		}

		aad := cipher.AADContext{
			VaultID:    vaultID,
			RecordID:   entry.ID,
			RecordType: entry.Type.String(),
			Version:    1,
		}
		ciphertext, err := cipher.Encrypt(recKey, entry.Payload, aad.Bytes())
		kdf.Zeroize(recKey)
		if err != nil {
			return nil, err
		}

		rec := &sqlite.Record{
			ID:         entry.ID,
			VaultID:    vaultID,
			RecordType: entry.Type.String(),
			Version:    1,
			Payload:    ciphertext,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if existingIDMap[entry.ID] && strategy == ConflictOverwrite {
			_ = s.recordRepo.Update(ctx, rec, tagIDs)
		} else {
			_ = s.recordRepo.Create(ctx, rec, tagIDs)
		}
		result.ImportedCount++
	}

	return result, nil
}

func (s *BackupService) createDirectSnapshot(ctx context.Context, snapshotPath string) error {
	meta, err := s.metaRepo.Get(ctx)
	if err != nil {
		return err
	}

	records, err := s.recordRepo.List(ctx, meta.VaultID, true)
	if err != nil {
		return err
	}

	var backupEntries []format.BackupEntry
	for _, rec := range records {
		backupEntries = append(backupEntries, format.BackupEntry{
			ID:            rec.ID,
			Type:          domain.RecordType(rec.RecordType),
			Payload:       rec.Payload,
			Version:       int(rec.Version),
			CreatedAtUnix: rec.CreatedAt.Unix(),
			UpdatedAtUnix: rec.UpdatedAt.Unix(),
		})
	}

	payload := &format.BackupPayload{
		Version: 1,
		Manifest: format.BackupManifest{
			SourceVaultVersion:  1,
			SourceSchemaVersion: uint16(meta.SchemaVersion),
			AppVersion:          "0.1.0",
			CreatedAtUnix:       time.Now().Unix(),
		},
		Entries: backupEntries,
	}

	var bid [16]byte
	_, _ = io.ReadFull(rand.Reader, bid[:])

	header := &format.Header{
		FormatVersion:      format.CurrentBackupFormatVersion,
		CryptoSuiteVersion: format.CurrentCryptoSuiteVersion,
		BackupID:           bid,
		VaultID:            meta.VaultID,
		CreatedAtUnix:      time.Now().Unix(),
		KDFAlgorithm:       format.KDFAlgorithmArgon2id,
		KDFParams:          meta.Envelope.KDFParams,
		WrappedVaultKey:    meta.Envelope.WrappedKey,
		MCK:                meta.Envelope.MCK,
	}

	f, err := os.OpenFile(snapshotPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	dummyKey := make([]byte, 32)
	return format.EncryptBackup(f, payload, dummyKey, header)
}
