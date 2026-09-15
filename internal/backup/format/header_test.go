package format_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/mottamarcio/govault/internal/backup/format"
	"github.com/mottamarcio/govault/internal/crypto/kdf"
)

func sampleValidHeader(t *testing.T) *format.Header {
	t.Helper()
	params, err := kdf.DefaultArgon2Params()
	if err != nil {
		t.Fatalf("unexpected error creating params: %v", err)
	}

	var bid [16]byte
	copy(bid[:], []byte("1234567890123456"))

	var nonce [24]byte
	copy(nonce[:], []byte("123456789012345678901234"))

	return &format.Header{
		FormatVersion:      format.CurrentBackupFormatVersion,
		CryptoSuiteVersion: format.CurrentCryptoSuiteVersion,
		BackupID:           bid,
		VaultID:            "vault-uuid-1234",
		CreatedAtUnix:      time.Now().Unix(),
		KDFAlgorithm:       format.KDFAlgorithmArgon2id,
		KDFParams:          params,
		WrappedVaultKey:    bytes.Repeat([]byte{0xAA}, 56),
		MCK:                bytes.Repeat([]byte{0xBB}, 32),
		BackupNonce:        nonce,
		CiphertextLength:   1024,
	}
}

func TestHeaderMarshalParseRoundTrip(t *testing.T) {
	h := sampleValidHeader(t)

	data, err := h.MarshalBinary()
	if err != nil {
		t.Fatalf("failed to marshal header: %v", err)
	}
	if len(data) < format.MinHeaderSize {
		t.Fatalf("expected header length >= %d, got %d", format.MinHeaderSize, len(data))
	}

	parsed, err := format.ParseHeader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to parse header: %v", err)
	}

	if parsed.FormatVersion != h.FormatVersion {
		t.Fatalf("FormatVersion mismatch: got %d, want %d", parsed.FormatVersion, h.FormatVersion)
	}
	if parsed.CryptoSuiteVersion != h.CryptoSuiteVersion {
		t.Fatalf("CryptoSuiteVersion mismatch: got %d, want %d", parsed.CryptoSuiteVersion, h.CryptoSuiteVersion)
	}
	if parsed.BackupID != h.BackupID {
		t.Fatalf("BackupID mismatch")
	}
	if parsed.VaultID != h.VaultID {
		t.Fatalf("VaultID mismatch: got %s, want %s", parsed.VaultID, h.VaultID)
	}
	if parsed.CreatedAtUnix != h.CreatedAtUnix {
		t.Fatalf("CreatedAtUnix mismatch")
	}
	if parsed.KDFAlgorithm != h.KDFAlgorithm {
		t.Fatalf("KDFAlgorithm mismatch")
	}
	if parsed.KDFParams.Memory != h.KDFParams.Memory {
		t.Fatalf("KDF Memory mismatch")
	}
	if parsed.KDFParams.Time != h.KDFParams.Time {
		t.Fatalf("KDF Time mismatch")
	}
	if parsed.KDFParams.Threads != h.KDFParams.Threads {
		t.Fatalf("KDF Threads mismatch")
	}
	if !bytes.Equal(parsed.KDFParams.Salt, h.KDFParams.Salt) {
		t.Fatalf("KDF Salt mismatch")
	}
	if !bytes.Equal(parsed.WrappedVaultKey, h.WrappedVaultKey) {
		t.Fatalf("WrappedVaultKey mismatch")
	}
	if !bytes.Equal(parsed.MCK, h.MCK) {
		t.Fatalf("MCK mismatch")
	}
	if parsed.BackupNonce != h.BackupNonce {
		t.Fatalf("BackupNonce mismatch")
	}
	if parsed.CiphertextLength != h.CiphertextLength {
		t.Fatalf("CiphertextLength mismatch")
	}
}

func TestInspectHeader(t *testing.T) {
	h := sampleValidHeader(t)
	data, err := h.MarshalBinary()
	if err != nil {
		t.Fatalf("failed to marshal header: %v", err)
	}

	inspection, err := format.InspectHeader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("failed to inspect header: %v", err)
	}
	if inspection.BackupID == "" {
		t.Fatalf("expected non-empty BackupID")
	}
	if inspection.VaultID != "vault-uuid-1234" {
		t.Fatalf("expected vault-uuid-1234, got %s", inspection.VaultID)
	}
	if inspection.FormatVersion != 1 {
		t.Fatalf("expected version 1, got %d", inspection.FormatVersion)
	}
	if inspection.KDFMemoryMB != 64 {
		t.Fatalf("expected 64MB, got %d", inspection.KDFMemoryMB)
	}
	if inspection.CiphertextLength != 1024 {
		t.Fatalf("expected 1024, got %d", inspection.CiphertextLength)
	}
}

func TestHeaderValidationErrors(t *testing.T) {
	t.Run("Invalid magic", func(t *testing.T) {
		data := []byte("INVALIDM\x00\x01\x00\x01\x00\x00\x00\x80")
		_, err := format.ParseHeader(bytes.NewReader(data))
		if !errors.Is(err, format.ErrInvalidMagic) {
			t.Fatalf("expected ErrInvalidMagic, got %v", err)
		}
	})

	t.Run("Unsupported version", func(t *testing.T) {
		h := sampleValidHeader(t)
		h.FormatVersion = 99
		data, err := h.MarshalBinary()
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}

		_, err = format.ParseHeader(bytes.NewReader(data))
		if !errors.Is(err, format.ErrUnsupportedVersion) {
			t.Fatalf("expected ErrUnsupportedVersion, got %v", err)
		}
	})

	t.Run("Truncated header", func(t *testing.T) {
		h := sampleValidHeader(t)
		data, err := h.MarshalBinary()
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}

		_, err = format.ParseHeader(bytes.NewReader(data[:50]))
		if !errors.Is(err, format.ErrCorruptHeader) {
			t.Fatalf("expected ErrCorruptHeader, got %v", err)
		}
	})
}
