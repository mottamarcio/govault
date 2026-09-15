package format_test

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/mottamarcio/govault/internal/backup/format"
	"github.com/mottamarcio/govault/internal/crypto/kdf"
	"github.com/mottamarcio/govault/internal/domain"
)

func TestBackupCryptoRoundTrip(t *testing.T) {
	vaultKey := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, vaultKey)

	vaultID := "vault-uuid-5678"
	params, err := kdf.DefaultArgon2Params()
	if err != nil {
		t.Fatalf("unexpected error creating KDF params: %v", err)
	}

	var bid [16]byte
	_, _ = io.ReadFull(rand.Reader, bid[:])

	header := &format.Header{
		FormatVersion:      format.CurrentBackupFormatVersion,
		CryptoSuiteVersion: format.CurrentCryptoSuiteVersion,
		BackupID:           bid,
		VaultID:            vaultID,
		CreatedAtUnix:      time.Now().Unix(),
		KDFAlgorithm:       format.KDFAlgorithmArgon2id,
		KDFParams:          params,
		WrappedVaultKey:    bytes.Repeat([]byte{0x11}, 56),
		MCK:                bytes.Repeat([]byte{0x22}, 32),
	}

	loginPayload := domain.LoginPayload{
		Username: "admin",
		Password: "SecurePassword999!",
	}
	loginBytes, _ := domain.SerializePayload(loginPayload)

	payload := &format.BackupPayload{
		Version: 1,
		Manifest: format.BackupManifest{
			SourceVaultVersion:  1,
			SourceSchemaVersion: 1,
			AppVersion:          "0.1.0",
			CreatedAtUnix:       time.Now().Unix(),
		},
		Entries: []format.BackupEntry{
			{
				ID:            "entry-100",
				Type:          domain.RecordTypeLogin,
				Title:         "Admin Portal",
				Payload:       loginBytes,
				Version:       1,
				CreatedAtUnix: time.Now().Unix(),
				UpdatedAtUnix: time.Now().Unix(),
			},
		},
	}

	buf := new(bytes.Buffer)
	err = format.EncryptBackup(buf, payload, vaultKey, header)
	if err != nil {
		t.Fatalf("failed to encrypt backup: %v", err)
	}

	backupBytes := buf.Bytes()
	if len(backupBytes) <= format.MinHeaderSize {
		t.Fatalf("backup bytes too small: %d", len(backupBytes))
	}

	// Successful Decrypt
	parsedHeader, decryptedPayload, err := format.DecryptBackup(bytes.NewReader(backupBytes), vaultKey)
	if err != nil {
		t.Fatalf("failed to decrypt backup: %v", err)
	}

	if parsedHeader.VaultID != vaultID {
		t.Fatalf("vault ID mismatch: got %s, want %s", parsedHeader.VaultID, vaultID)
	}
	if len(decryptedPayload.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(decryptedPayload.Entries))
	}
	if decryptedPayload.Entries[0].Title != "Admin Portal" {
		t.Fatalf("entry title mismatch: %s", decryptedPayload.Entries[0].Title)
	}

	// Failed Decrypt with wrong key
	wrongKey := make([]byte, 32)
	_, _ = io.ReadFull(rand.Reader, wrongKey)
	_, _, err = format.DecryptBackup(bytes.NewReader(backupBytes), wrongKey)
	if !errors.Is(err, format.ErrAuthenticationFail) {
		t.Fatalf("expected ErrAuthenticationFail with wrong key, got %v", err)
	}

	// Tampered Header: modify a byte in vault ID inside the header
	tamperedBytes := make([]byte, len(backupBytes))
	copy(tamperedBytes, backupBytes)
	// Header offset where vault_id starts (around byte 34)
	tamperedBytes[34] ^= 0xFF
	_, _, err = format.DecryptBackup(bytes.NewReader(tamperedBytes), vaultKey)
	if !errors.Is(err, format.ErrAuthenticationFail) && !errors.Is(err, format.ErrCorruptHeader) {
		t.Fatalf("expected authentication failure for tampered header, got %v", err)
	}
}
