package service

import (
	"sync"

	"github.com/mottamarcio/govault/internal/crypto/kdf"
)

// Session manages the in-memory state and decrypted VaultKey for an unlocked vault.
type Session struct {
	mu       sync.RWMutex
	vaultID  string
	vaultKey []byte
	unlocked bool
}

// NewSession creates a new locked session.
func NewSession() *Session {
	return &Session{
		unlocked: false,
	}
}

// Unlock sets the active vault ID and decrypted VaultKey.
func (s *Session) Unlock(vaultID string, vaultKey []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If previously holding a key, zeroize it first
	if len(s.vaultKey) > 0 {
		kdf.Zeroize(s.vaultKey)
	}

	s.vaultID = vaultID
	s.vaultKey = make([]byte, len(vaultKey))
	copy(s.vaultKey, vaultKey)
	s.unlocked = true
}

// IsUnlocked returns whether the session is currently active and unlocked.
func (s *Session) IsUnlocked() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.unlocked
}

// VaultID returns the current unlocked vault ID, or ErrVaultLocked if locked.
func (s *Session) VaultID() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.unlocked {
		return "", ErrVaultLocked
	}
	return s.vaultID, nil
}

// VaultKey returns a copy of the decrypted VaultKey, or ErrVaultLocked if locked.
func (s *Session) VaultKey() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.unlocked {
		return nil, ErrVaultLocked
	}

	keyCopy := make([]byte, len(s.vaultKey))
	copy(keyCopy, s.vaultKey)
	return keyCopy, nil
}

// Lock clears the active session and zeroizes the decrypted VaultKey from memory.
func (s *Session) Lock() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.vaultKey) > 0 {
		kdf.Zeroize(s.vaultKey)
		s.vaultKey = nil
	}
	s.vaultID = ""
	s.unlocked = false
}
