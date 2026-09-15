package format

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/mottamarcio/govault/internal/crypto/kdf"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/hkdf"
)

// DeriveBackupKey derives the cryptographic BackupKey using HKDF-SHA256 with vaultID as salt.
func DeriveBackupKey(vaultKey []byte, vaultID string) ([]byte, error) {
	if len(vaultKey) != 32 {
		return nil, fmt.Errorf("invalid vault key size: got %d, expected 32", len(vaultKey))
	}
	if len(vaultID) == 0 {
		return nil, fmt.Errorf("vault_id cannot be empty")
	}

	reader := hkdf.New(sha256.New, vaultKey, []byte(vaultID), []byte("govault:v1:backup"))
	backupKey := make([]byte, 32)
	if _, err := io.ReadFull(reader, backupKey); err != nil {
		return nil, fmt.Errorf("failed to derive backup key: %w", err)
	}

	return backupKey, nil
}

// ConstructBackupAAD generates canonical AAD bytes binding the header metadata.
func ConstructBackupAAD(h *Header) []byte {
	return []byte(fmt.Sprintf(
		"govault/v1/backup|format:%d|suite:%d|bid:%s|vid:%s|created:%d",
		h.FormatVersion,
		h.CryptoSuiteVersion,
		hex.EncodeToString(h.BackupID[:]),
		h.VaultID,
		h.CreatedAtUnix,
	))
}

// EncryptBackup serializes and encrypts the logical payload and writes the complete .gvault stream.
func EncryptBackup(w io.Writer, payload *BackupPayload, vaultKey []byte, header *Header) error {
	payloadBytes, err := payload.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal backup payload: %w", err)
	}

	backupKey, err := DeriveBackupKey(vaultKey, header.VaultID)
	if err != nil {
		return fmt.Errorf("failed to derive backup key: %w", err)
	}
	defer kdf.Zeroize(backupKey)

	// Ensure fresh 24-byte random nonce
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return fmt.Errorf("failed to generate random backup nonce: %w", err)
	}
	header.BackupNonce = nonce

	aad := ConstructBackupAAD(header)

	aead, err := chacha20poly1305.NewX(backupKey)
	if err != nil {
		return fmt.Errorf("failed to initialize XChaCha20-Poly1305: %w", err)
	}

	ciphertextWithTag := aead.Seal(nil, header.BackupNonce[:], payloadBytes, aad)

	header.CiphertextLength = uint64(len(ciphertextWithTag))

	headerBytes, err := header.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal backup header: %w", err)
	}

	if _, err := w.Write(headerBytes); err != nil {
		return fmt.Errorf("failed to write backup header: %w", err)
	}
	if _, err := w.Write(ciphertextWithTag); err != nil {
		return fmt.Errorf("failed to write backup ciphertext: %w", err)
	}

	return nil
}

// DecryptBackup parses the header, verifies AAD, and decrypts the backup payload from an io.Reader.
func DecryptBackup(r io.Reader, vaultKey []byte) (*Header, *BackupPayload, error) {
	header, err := ParseHeader(r)
	if err != nil {
		return nil, nil, err
	}

	if header.CiphertextLength == 0 || header.CiphertextLength > MaxCiphertext {
		return nil, nil, ErrInvalidCiphertext
	}

	ciphertextWithTag := make([]byte, header.CiphertextLength)
	if _, err := io.ReadFull(r, ciphertextWithTag); err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	backupKey, err := DeriveBackupKey(vaultKey, header.VaultID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive backup key: %w", err)
	}
	defer kdf.Zeroize(backupKey)

	aad := ConstructBackupAAD(header)

	aead, err := chacha20poly1305.NewX(backupKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize XChaCha20-Poly1305: %w", err)
	}

	plaintext, err := aead.Open(nil, header.BackupNonce[:], ciphertextWithTag, aad)
	if err != nil {
		return nil, nil, ErrAuthenticationFail
	}

	payload, err := UnmarshalPayload(plaintext)
	if err != nil {
		return nil, nil, err
	}

	return header, payload, nil
}
