package kdf

import (
	"crypto/subtle"
)

// VerifyMCK performs a constant-time comparison between candidateMCK and expectedMCK.
// Returns true if both keys are exactly 32 bytes and match byte-for-byte.
func VerifyMCK(candidateMCK, expectedMCK []byte) bool {
	if len(candidateMCK) != 32 || len(expectedMCK) != 32 {
		return false
	}
	return subtle.ConstantTimeCompare(candidateMCK, expectedMCK) == 1
}
