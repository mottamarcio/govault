package format

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/mottamarcio/govault/internal/crypto/cipher"
	"github.com/mottamarcio/govault/internal/crypto/kdf"
)

const (
	MagicString                = "GOVAULTB"
	CurrentBackupFormatVersion = 1
	CurrentCryptoSuiteVersion  = 1
	KDFAlgorithmArgon2id       = 1

	MinHeaderSize = 138 // Minimum static size with expected wrapped key & MCK
	MaxHeaderSize = 64 * 1024
	MaxCiphertext = 8 * 1024 * 1024 * 1024 // 8 GiB
)

// Header holds all unencrypted metadata required to authenticate and decrypt a backup file.
type Header struct {
	FormatVersion      uint16
	CryptoSuiteVersion uint16
	BackupID           [16]byte
	VaultID            string
	CreatedAtUnix      int64
	KDFAlgorithm       uint16
	KDFParams          *kdf.Argon2Params
	WrappedVaultKey    []byte
	MCK                []byte
	BackupNonce        [24]byte
	CiphertextLength   uint64
}

// MarshalBinary encodes the header into canonical big-endian binary representation.
func (h *Header) MarshalBinary() ([]byte, error) {
	if h.FormatVersion == 0 {
		h.FormatVersion = CurrentBackupFormatVersion
	}
	if h.CryptoSuiteVersion == 0 {
		h.CryptoSuiteVersion = CurrentCryptoSuiteVersion
	}
	if h.KDFAlgorithm == 0 {
		h.KDFAlgorithm = KDFAlgorithmArgon2id
	}
	if h.KDFParams == nil {
		return nil, ErrInvalidKDFParams
	}
	if err := h.KDFParams.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKDFParams, err)
	}
	if len(h.WrappedVaultKey) < cipher.MinPayloadSize {
		return nil, ErrInvalidWrappedKey
	}
	if len(h.MCK) != 32 {
		return nil, ErrInvalidWrappedKey
	}

	buf := new(bytes.Buffer)

	// 1. Magic (8 bytes)
	buf.WriteString(MagicString)

	// 2. Format & Crypto Suite versions (2 + 2 = 4 bytes)
	if err := binary.Write(buf, binary.BigEndian, h.FormatVersion); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.CryptoSuiteVersion); err != nil {
		return nil, err
	}

	// 3. Placeholder for HeaderLength (4 bytes, uint32)
	headerLenOffset := buf.Len()
	if err := binary.Write(buf, binary.BigEndian, uint32(0)); err != nil {
		return nil, err
	}

	// 4. Backup ID (16 bytes)
	buf.Write(h.BackupID[:])

	// 5. Vault ID Length (2 bytes) + Vault ID bytes
	vaultIDBytes := []byte(h.VaultID)
	if len(vaultIDBytes) > 255 {
		return nil, fmt.Errorf("vault_id too long: %d", len(vaultIDBytes))
	}
	if err := binary.Write(buf, binary.BigEndian, uint16(len(vaultIDBytes))); err != nil {
		return nil, err
	}
	buf.Write(vaultIDBytes)

	// 6. CreatedAt Unix timestamp (8 bytes)
	if err := binary.Write(buf, binary.BigEndian, h.CreatedAtUnix); err != nil {
		return nil, err
	}

	// 7. KDF Algorithm & Reserved (2 + 2 = 4 bytes)
	if err := binary.Write(buf, binary.BigEndian, h.KDFAlgorithm); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil { // reserved
		return nil, err
	}

	// 8. KDF Parameters: Memory, Time, Threads (4 + 4 + 1 = 9 bytes) + 3 bytes padding
	if err := binary.Write(buf, binary.BigEndian, h.KDFParams.Memory); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.KDFParams.Time); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.KDFParams.Threads); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint8(h.KDFParams.KeyLen)); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, uint16(len(h.KDFParams.Salt))); err != nil {
		return nil, err
	}
	buf.Write(h.KDFParams.Salt)

	// 9. Wrapped Vault Key & MCK
	if err := binary.Write(buf, binary.BigEndian, uint16(len(h.WrappedVaultKey))); err != nil {
		return nil, err
	}
	buf.Write(h.WrappedVaultKey)
	buf.Write(h.MCK)

	// 10. Backup Nonce (24 bytes)
	buf.Write(h.BackupNonce[:])

	// 11. Ciphertext Length (8 bytes)
	if err := binary.Write(buf, binary.BigEndian, h.CiphertextLength); err != nil {
		return nil, err
	}

	res := buf.Bytes()
	headerLen := uint32(len(res))
	if headerLen > MaxHeaderSize {
		return nil, ErrInvalidHeaderLength
	}

	// Update header_length field
	binary.BigEndian.PutUint32(res[headerLenOffset:headerLenOffset+4], headerLen)

	return res, nil
}

// ParseHeader reads and decodes the Header from an io.Reader.
func ParseHeader(r io.Reader) (*Header, error) {
	magic := make([]byte, 8)
	if _, err := io.ReadFull(r, magic); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}
	if string(magic) != MagicString {
		return nil, ErrInvalidMagic
	}

	var formatVer, cryptoVer uint16
	if err := binary.Read(r, binary.BigEndian, &formatVer); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}
	if err := binary.Read(r, binary.BigEndian, &cryptoVer); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	if formatVer != CurrentBackupFormatVersion || cryptoVer != CurrentCryptoSuiteVersion {
		return nil, ErrUnsupportedVersion
	}

	var headerLen uint32
	if err := binary.Read(r, binary.BigEndian, &headerLen); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}
	if headerLen < MinHeaderSize || headerLen > MaxHeaderSize {
		return nil, ErrInvalidHeaderLength
	}

	// Read remaining header bytes
	remainingLen := int(headerLen) - 16 // 8 magic + 2 + 2 + 4 = 16
	headerBody := make([]byte, remainingLen)
	if _, err := io.ReadFull(r, headerBody); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	reader := bytes.NewReader(headerBody)

	var backupID [16]byte
	if _, err := io.ReadFull(reader, backupID[:]); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	var vaultIDLen uint16
	if err := binary.Read(reader, binary.BigEndian, &vaultIDLen); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}
	vaultIDBytes := make([]byte, vaultIDLen)
	if _, err := io.ReadFull(reader, vaultIDBytes); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	var createdAtUnix int64
	if err := binary.Read(reader, binary.BigEndian, &createdAtUnix); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	var kdfAlg, reserved uint16
	if err := binary.Read(reader, binary.BigEndian, &kdfAlg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}
	if err := binary.Read(reader, binary.BigEndian, &reserved); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	var mem, timeCount uint32
	var threads, keyLen uint8
	var saltLen uint16

	if err := binary.Read(reader, binary.BigEndian, &mem); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}
	if err := binary.Read(reader, binary.BigEndian, &timeCount); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}
	if err := binary.Read(reader, binary.BigEndian, &threads); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}
	if err := binary.Read(reader, binary.BigEndian, &keyLen); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}
	if err := binary.Read(reader, binary.BigEndian, &saltLen); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	if saltLen < 16 || saltLen > 256 {
		return nil, ErrInvalidKDFParams
	}
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(reader, salt); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	kdfParams := &kdf.Argon2Params{
		Memory:  mem,
		Time:    timeCount,
		Threads: threads,
		KeyLen:  uint32(keyLen),
		Salt:    salt,
	}
	if err := kdfParams.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKDFParams, err)
	}

	var wrappedKeyLen uint16
	if err := binary.Read(reader, binary.BigEndian, &wrappedKeyLen); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}
	if wrappedKeyLen < cipher.MinPayloadSize || wrappedKeyLen > 1024 {
		return nil, ErrInvalidWrappedKey
	}

	wrappedKey := make([]byte, wrappedKeyLen)
	if _, err := io.ReadFull(reader, wrappedKey); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	mck := make([]byte, 32)
	if _, err := io.ReadFull(reader, mck); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	var backupNonce [24]byte
	if _, err := io.ReadFull(reader, backupNonce[:]); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}

	var ciphertextLen uint64
	if err := binary.Read(reader, binary.BigEndian, &ciphertextLen); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCorruptHeader, err)
	}
	if ciphertextLen > MaxCiphertext {
		return nil, ErrInvalidCiphertext
	}

	return &Header{
		FormatVersion:      formatVer,
		CryptoSuiteVersion: cryptoVer,
		BackupID:           backupID,
		VaultID:            string(vaultIDBytes),
		CreatedAtUnix:      createdAtUnix,
		KDFAlgorithm:       kdfAlg,
		KDFParams:          kdfParams,
		WrappedVaultKey:    wrappedKey,
		MCK:                mck,
		BackupNonce:        backupNonce,
		CiphertextLength:   ciphertextLen,
	}, nil
}
