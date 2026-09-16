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

func setupFreshVaultService(t *testing.T) (*service.VaultService, string) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "fresh_vault.db")

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

	return vs, dbPath
}

func TestInitModel_Validation(t *testing.T) {
	vs, dbPath := setupFreshVaultService(t)
	th := theme.DefaultTheme()
	m := screens.NewInitModel(th, vs, dbPath)

	// Short password
	for _, ch := range "short" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, ch := range "short" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("expected nil cmd on invalid short password")
	}
	if !strings.Contains(m.ErrMsg, "8 characters") {
		t.Fatalf("expected 8 characters error, got: %q", m.ErrMsg)
	}

	// Mismatched passwords
	m.ResetInput()
	for _, ch := range "CorrectPassword123!" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, ch := range "DifferentPassword123!" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("expected nil cmd on mismatched passwords")
	}
	if !strings.Contains(m.ErrMsg, "match") {
		t.Fatalf("expected mismatch error, got: %q", m.ErrMsg)
	}
}

func TestInitModel_Success(t *testing.T) {
	vs, dbPath := setupFreshVaultService(t)
	th := theme.DefaultTheme()
	m := screens.NewInitModel(th, vs, dbPath)

	// Enter password
	for _, ch := range "ValidMasterPass123!" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	// Tab to confirm input
	m.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Enter confirmation
	for _, ch := range "ValidMasterPass123!" {
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	// Press Enter
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on Enter with valid input")
	}

	// Execute async initialization cmd
	msg := cmd()
	initSuccessMsg, ok := msg.(screens.InitSuccessMsg)
	if !ok {
		t.Fatalf("expected InitSuccessMsg, got %T: %+v", msg, msg)
	}
	if initSuccessMsg.Session == nil || !initSuccessMsg.Session.IsUnlocked() {
		t.Fatal("expected unlocked session in init success message")
	}

	// Update model with success
	m.Update(initSuccessMsg)
	if m.PassInput.Value() != "" || m.ConfirmInput.Value() != "" {
		t.Errorf("expected inputs to be wiped after success, got pass=%q confirm=%q", m.PassInput.Value(), m.ConfirmInput.Value())
	}
}
