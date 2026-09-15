package format

import "errors"

var (
	ErrInvalidMagic        = errors.New("backup: invalid magic bytes")
	ErrUnsupportedVersion  = errors.New("backup: unsupported format or crypto suite version")
	ErrInvalidHeaderLength = errors.New("backup: invalid header length")
	ErrInvalidKDFParams    = errors.New("backup: invalid KDF parameters")
	ErrInvalidWrappedKey   = errors.New("backup: invalid wrapped key or MCK")
	ErrInvalidCiphertext   = errors.New("backup: invalid ciphertext length")
	ErrCorruptHeader       = errors.New("backup: corrupt or truncated header")
	ErrAuthenticationFail  = errors.New("backup: authentication failed or header tampered")
	ErrCorruptPayload      = errors.New("backup: corrupt or invalid backup payload")
)
