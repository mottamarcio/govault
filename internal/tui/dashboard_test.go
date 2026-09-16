package tui_test

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/tui"
	"github.com/mottamarcio/govault/internal/tui/screens"
)

func TestAppModel_EndToEndDashboardWorkflow(t *testing.T) {
	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	ctx := context.Background()

	// Unlock vault
	sess, err := app.VaultService.Unlock(ctx, "SecretPassword123!")
	if err != nil {
		t.Fatalf("failed to unlock vault: %v", err)
	}

	// Seed records into vault: seed note first, then login
	_, err = rsSeedNote(app.RecordService, ctx)
	if err != nil {
		t.Fatalf("failed to seed note: %v", err)
	}

	_, err = app.RecordService.Create(ctx, service.RecordInput{
		Title: "Google Work Account",
		Type:  domain.RecordTypeLogin,
		Tags:  []string{"work", "email"},
		Payload: &domain.LoginPayload{
			Username: "marcio@google.com",
			Password: "GoogleSuperSecretPassword!",
			URI:      "https://accounts.google.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to create google login: %v", err)
	}

	// Set window dimensions and unlock
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	app.Update(screens.UnlockSuccessMsg{Session: sess})

	if app.State != tui.ScreenDashboard {
		t.Fatalf("expected state ScreenDashboard, got %v", app.State)
	}

	// View should render categories and table
	view := app.View()
	if !strings.Contains(view, "CATEGORIES") {
		t.Errorf("expected view to contain CATEGORIES, got: %s", view)
	}
	if !strings.Contains(view, "accounts.google") {
		t.Errorf("expected view to contain login title URI, got: %s", view)
	}

	// Move sidebar to Logins (down arrow / 'j')
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	// Switch focus to Table (Tab)
	app.Update(tea.KeyMsg{Type: tea.KeyTab})

	// Toggle password reveal (v)
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	viewRevealed := app.View()
	if !strings.Contains(viewRevealed, "GoogleSuperSecretPassword!") {
		t.Errorf("expected revealed password in view, got: %s", viewRevealed)
	}

	// Copy password (c)
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	clipContent, err := app.Clipboard.Read()
	if err != nil || clipContent != "GoogleSuperSecretPassword!" {
		t.Errorf("expected clipboard to contain 'GoogleSuperSecretPassword!', got %q (err: %v)", clipContent, err)
	}

	// Copy username (u)
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	userContent, err := app.Clipboard.Read()
	if err != nil || userContent != "marcio@google.com" {
		t.Errorf("expected clipboard to contain 'marcio@google.com', got %q (err: %v)", userContent, err)
	}

	// Lock with Ctrl+L
	app.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	if app.State != tui.ScreenUnlock {
		t.Fatalf("expected state ScreenUnlock after Ctrl+L, got %v", app.State)
	}
}

func rsSeedNote(rs *service.RecordService, ctx context.Context) (*domain.Record, error) {
	return rs.Create(ctx, service.RecordInput{
		Title: "Server Notes",
		Type:  domain.RecordTypeNote,
		Tags:  []string{"infra"},
		Payload: &domain.NotePayload{
			Content: "Master cluster node IP: 10.0.0.1",
		},
	})
}
