package screens_test

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/tui/screens"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

func TestDashboardModel_SearchFilterAndClipboard(t *testing.T) {
	rs, _, cleanup := setupTestRecordService(t)
	defer cleanup()

	ctx := context.Background()

	// Seed records
	_, err := rs.Create(ctx, service.RecordInput{
		Title: "GitHub Personal",
		Type:  domain.RecordTypeLogin,
		Tags:  []string{"dev", "personal"},
		Payload: &domain.LoginPayload{
			Username: "marcio",
			Password: "GitPassword123!",
			URI:      "https://github.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to create login: %v", err)
	}

	_, err = rs.Create(ctx, service.RecordInput{
		Title: "Server SSH",
		Type:  domain.RecordTypeNote,
		Tags:  []string{"infra"},
		Payload: &domain.NotePayload{
			Content: "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5",
		},
	})
	if err != nil {
		t.Fatalf("failed to create note: %v", err)
	}

	clipDriver := service.NewInMemoryClipboardDriver()
	clipService := service.NewClipboardService(clipDriver)

	th := theme.DefaultTheme()
	dash := screens.NewDashboardModel(th, rs, clipService, 100, 30)

	if err := dash.ReloadRecords(); err != nil {
		t.Fatalf("failed to reload records: %v", err)
	}

	if len(dash.FilteredRecords) != 2 {
		t.Fatalf("expected 2 records loaded, got %d", len(dash.FilteredRecords))
	}

	// Activate search with '/'
	m, _ := dash.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	dash = m.(*screens.DashboardModel)
	if dash.Focus != screens.FocusSearch {
		t.Fatalf("expected FocusSearch, got %v", dash.Focus)
	}

	// Type 'git'
	for _, ch := range "git" {
		m, _ = dash.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
		dash = m.(*screens.DashboardModel)
	}

	if len(dash.FilteredRecords) != 1 || dash.FilteredRecords[0].Title != "https://github.com" {
		t.Fatalf("expected 1 ranked record 'https://github.com', got %d (title: %q, search val: %q)", len(dash.FilteredRecords), dash.FilteredRecords[0].Title, dash.SearchInput.Value())
	}

	// Exit search with Enter
	dash.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if dash.Focus != screens.FocusTable {
		t.Fatalf("expected FocusTable after Enter, got %v", dash.Focus)
	}

	// Press 'c' to copy password to clipboard
	_, cmd := dash.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on copy hotkey")
	}

	copied, err := clipDriver.ReadText()
	if err != nil || copied != "GitPassword123!" {
		t.Fatalf("expected copied password 'GitPassword123!', got %q (err: %v)", copied, err)
	}

	// Press 'u' to copy username to clipboard
	dash.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	copiedUser, err := clipDriver.ReadText()
	if err != nil || copiedUser != "marcio" {
		t.Fatalf("expected copied username 'marcio', got %q (err: %v)", copiedUser, err)
	}

	view := dash.View()
	if !strings.Contains(view, "https://github.com") {
		t.Errorf("expected view to contain record title, got: %s", view)
	}
}
