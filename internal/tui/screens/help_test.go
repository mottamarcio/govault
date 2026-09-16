package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/govault/internal/tui/theme"
)

func TestModalHelp(t *testing.T) {
	th := theme.DefaultTheme()
	h := NewModalHelp(th)
	h.SetDimensions(80, 24)

	if h.Active {
		t.Fatalf("expected help modal to start inactive")
	}
	if v := h.View(); v != "" {
		t.Fatalf("expected empty view when inactive, got %q", v)
	}

	h.Open()
	if !h.Active {
		t.Fatalf("expected active after Open")
	}

	view := h.View()
	if view == "" {
		t.Fatalf("expected rendered help view")
	}

	// Test close with Esc
	_, cmd := h.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatalf("expected close cmd")
	}
	if _, ok := cmd().(HelpCloseMsg); !ok {
		t.Fatalf("expected HelpCloseMsg")
	}
	if h.Active {
		t.Fatalf("expected inactive after close")
	}

	// Test toggle
	h.Toggle()
	if !h.Active {
		t.Fatalf("expected active after toggle")
	}
	// Test close with '?'
	_, cmd = h.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if cmd == nil {
		t.Fatalf("expected close cmd on '?'")
	}
	if h.Active {
		t.Fatalf("expected inactive after '?'")
	}
}
