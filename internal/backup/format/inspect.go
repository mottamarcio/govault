package format

import (
	"encoding/hex"
	"io"
	"time"
)

// HeaderInspection contains non-secret structural metadata extracted from a backup header.
type HeaderInspection struct {
	BackupID           string
	VaultID            string
	CreatedAt          time.Time
	FormatVersion      uint16
	CryptoSuiteVersion uint16
	KDFAlgorithm       uint16
	KDFMemoryMB        uint32
	KDFTime            uint32
	KDFThreads         uint8
	CiphertextLength   uint64
}

// InspectHeader reads and structurally validates the header without requiring password input.
func InspectHeader(r io.Reader) (*HeaderInspection, error) {
	h, err := ParseHeader(r)
	if err != nil {
		return nil, err
	}

	return &HeaderInspection{
		BackupID:           hex.EncodeToString(h.BackupID[:]),
		VaultID:            h.VaultID,
		CreatedAt:          time.Unix(h.CreatedAtUnix, 0).UTC(),
		FormatVersion:      h.FormatVersion,
		CryptoSuiteVersion: h.CryptoSuiteVersion,
		KDFAlgorithm:       h.KDFAlgorithm,
		KDFMemoryMB:        h.KDFParams.Memory / 1024,
		KDFTime:            h.KDFParams.Time,
		KDFThreads:         h.KDFParams.Threads,
		CiphertextLength:   h.CiphertextLength,
	}, nil
}
