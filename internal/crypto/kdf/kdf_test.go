package kdf

import (
	"encoding/hex"
	"testing"
)

func TestVerifyMCK(t *testing.T) {
	mck1 := make([]byte, 32)
	for i := range mck1 {
		mck1[i] = byte(i)
	}

	mck2 := make([]byte, 32)
	copy(mck2, mck1)

	if !VerifyMCK(mck1, mck2) {
		t.Fatalf("expected identical MCKs to verify as true")
	}

	// Tampered byte
	mck2[15] ^= 0xFF
	if VerifyMCK(mck1, mck2) {
		t.Fatalf("expected tampered MCK to verify as false")
	}

	// Length mismatch
	if VerifyMCK(mck1, mck1[:16]) {
		t.Fatalf("expected truncated MCK to verify as false")
	}
}

// End-to-end integration and KAT (Known Answer Test) test
func TestKDFEndToEndPipeline(t *testing.T) {
	password := []byte("correct-horse-battery-staple-2026")
	salt, _ := hex.DecodeString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")

	params := &Argon2Params{
		Time:    3,
		Memory:  64 * 1024,
		Threads: 4,
		KeyLen:  32,
		Salt:    salt,
	}

	// 1. Derive Master Key
	masterKey, err := DeriveMasterKey(password, params)
	if err != nil {
		t.Fatalf("DeriveMasterKey failed: %v", err)
	}
	if len(masterKey) != 32 {
		t.Fatalf("expected 32-byte master key, got %d", len(masterKey))
	}

	// 2. Derive KEK and MCK
	kek, err := DeriveKEK(masterKey)
	if err != nil {
		t.Fatalf("DeriveKEK failed: %v", err)
	}
	mck, err := DeriveMCK(masterKey)
	if err != nil {
		t.Fatalf("DeriveMCK failed: %v", err)
	}

	// 3. Verify confirmation check succeeds with correct password
	candidateMasterKey, _ := DeriveMasterKey(password, params)
	candidateMCK, _ := DeriveMCK(candidateMasterKey)

	if !VerifyMCK(candidateMCK, mck) {
		t.Fatalf("verification failed for matching password")
	}

	// 4. Verify rejection with wrong password
	wrongMasterKey, _ := DeriveMasterKey([]byte("wrong-password"), params)
	wrongMCK, _ := DeriveMCK(wrongMasterKey)

	if VerifyMCK(wrongMCK, mck) {
		t.Fatalf("verification succeeded for wrong password")
	}

	// 5. Cleanup / Zeroization
	Zeroize(masterKey)
	Zeroize(kek)
	Zeroize(candidateMasterKey)
	Zeroize(wrongMasterKey)
}
