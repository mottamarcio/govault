package screens_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/tui/screens"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

func TestSidebarModel(t *testing.T) {
	th := theme.DefaultTheme()
	sb := screens.NewSidebarModel(th)
	sb.SetDimensions(30, 20)

	// Mock records
	records := []*domain.Record{
		{ID: "1", Title: "GitHub", Type: domain.RecordTypeLogin, Tags: []string{"work", "dev"}},
		{ID: "2", Title: "Server Note", Type: domain.RecordTypeNote, Tags: []string{"work"}},
		{ID: "3", Title: "Stripe Key", Type: domain.RecordTypeAPIKey, Tags: []string{"prod"}},
	}

	sb.SetCountsAndTags(records)

	// Verify counts
	selected := sb.SelectedItem()
	if selected.Label != "All Items" || selected.Count != 3 {
		t.Fatalf("expected All Items with count 3, got: %+v", selected)
	}

	// Move down to Logins
	_, cmd := sb.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if cmd == nil {
		t.Fatal("expected non-nil cmd on navigation")
	}

	msg := cmd()
	selectMsg, ok := msg.(screens.SidebarSelectMsg)
	if !ok || selectMsg.Item.Label != "Logins" || selectMsg.Item.Count != 1 {
		t.Fatalf("expected Logins selected with count 1, got: %+v", msg)
	}

	view := sb.View()
	if !strings.Contains(view, "CATEGORIES") {
		t.Errorf("expected view to contain CATEGORIES, got: %s", view)
	}
	if !strings.Contains(view, "TAGS") {
		t.Errorf("expected view to contain TAGS header, got: %s", view)
	}
	if !strings.Contains(view, "# work") {
		t.Errorf("expected view to list '# work' tag, got: %s", view)
	}
}
