package screens_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
	"github.com/mottamarcio/govault/internal/tui/screens"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

func setupTestVaultService(t *testing.T) (*service.VaultService, string) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	metaRepo := sqlite.NewMetadataRepository(db)
	vs := service.NewVaultService(db, metaRepo, dbPath)

	err = vs.Init(ctx, "CorrectPassword123!")
	if err != nil {
		t.Fatalf("failed to init vault: %v", err)
	}

	return vs, dbPath
}

func TestUnlockModel_Success(t *testing.T) {
	vs, dbPath := setupTestVaultService(t)
	th := theme.DefaultTheme()
	m := screens.NewUnlockModel(th, vs, dbPath)

	// Simulate typing password
	for _, ch := range "CorrectPassword123!" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	// Press Enter
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on Enter")
	}

	// Execute async unlock cmd
	msg := cmd()
	successMsg, ok := msg.(screens.UnlockSuccessMsg)
	if !ok {
		t.Fatalf("expected UnlockSuccessMsg, got %T: %+v", msg, msg)
	}
	if successMsg.Session == nil || !successMsg.Session.IsUnlocked() {
		t.Fatal("expected unlocked session in success message")
	}

	// Update with success message
	m.Update(successMsg)
	if m.Input.Value() != "" {
		t.Errorf("expected input to be wiped, got %q", m.Input.Value())
	}
}

func TestUnlockModel_Failure(t *testing.T) {
	vs, dbPath := setupTestVaultService(t)
	th := theme.DefaultTheme()
	m := screens.NewUnlockModel(th, vs, dbPath)

	// Type wrong password
	for _, ch := range "WrongPassword" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	// Press Enter
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on Enter")
	}

	msg := cmd()
	errMsg, ok := msg.(screens.UnlockErrMsg)
	if !ok {
		t.Fatalf("expected UnlockErrMsg, got %T: %+v", msg, msg)
	}

	m.Update(errMsg)
	if m.ErrMsg == "" {
		t.Fatal("expected ErrMsg to be set on failure")
	}

	view := m.View()
	if !strings.Contains(view, "Authentication failed") {
		t.Errorf("expected view to contain error message, got: %s", view)
	}
}
