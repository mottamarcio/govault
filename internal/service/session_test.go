package service_test

import (
	"bytes"
	"sync"
	"testing"

	"github.com/mottamarcio/govault/internal/service"
)

func TestSession(t *testing.T) {
	session := service.NewSession()

	if session.IsUnlocked() {
		t.Fatal("new session should be locked")
	}

	if _, err := session.VaultID(); err != service.ErrVaultLocked {
		t.Fatalf("expected ErrVaultLocked, got %v", err)
	}

	if _, err := session.VaultKey(); err != service.ErrVaultLocked {
		t.Fatalf("expected ErrVaultLocked, got %v", err)
	}

	vaultID := "vault-uuid-123"
	vaultKey := bytes.Repeat([]byte{0x42}, 32)

	session.Unlock(vaultID, vaultKey)

	if !session.IsUnlocked() {
		t.Fatal("session should be unlocked")
	}

	gotID, err := session.VaultID()
	if err != nil || gotID != vaultID {
		t.Fatalf("VaultID() got %s, err %v; want %s", gotID, err, vaultID)
	}

	gotKey, err := session.VaultKey()
	if err != nil || !bytes.Equal(gotKey, vaultKey) {
		t.Fatalf("VaultKey() mismatch")
	}

	// Test lock
	session.Lock()

	if session.IsUnlocked() {
		t.Fatal("session should be locked after Lock()")
	}

	if _, err := session.VaultID(); err != service.ErrVaultLocked {
		t.Fatalf("expected ErrVaultLocked after Lock(), got %v", err)
	}

	if _, err := session.VaultKey(); err != service.ErrVaultLocked {
		t.Fatalf("expected ErrVaultLocked after Lock(), got %v", err)
	}
}

func TestSessionConcurrency(t *testing.T) {
	session := service.NewSession()
	vaultID := "vault-concurrency"
	vaultKey := bytes.Repeat([]byte{0x77}, 32)

	session.Unlock(vaultID, vaultKey)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = session.IsUnlocked()
			_, _ = session.VaultID()
			_, _ = session.VaultKey()
		}()
		go func() {
			defer wg.Done()
			session.Unlock(vaultID, vaultKey)
		}()
	}

	wg.Wait()
	session.Lock()
}
