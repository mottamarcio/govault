package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/govault/internal/tui/theme"
)

func TestModalConfirm(t *testing.T) {
	th := theme.DefaultTheme()
	m := NewModalConfirm(th)
	m.SetDimensions(80, 24)

	if m.Active {
		t.Fatalf("expected modal to start inactive")
	}
	if v := m.View(); v != "" {
		t.Fatalf("expected empty view when inactive, got %q", v)
	}

	// Open trash confirmation
	m.Open(ModalActionTrash, "Confirm Delete", "Move to Trash? [y/N]", "rec-123")
	if !m.Active {
		t.Fatalf("expected modal to be active after Open")
	}

	view := m.View()
	if view == "" {
		t.Fatalf("expected rendered modal view, got empty string")
	}

	// Test confirm with 'y'
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if cmd == nil {
		t.Fatalf("expected non-nil cmd on confirmation")
	}
	msg := cmd()
	confirmMsg, ok := msg.(ModalConfirmMsg)
	if !ok {
		t.Fatalf("expected ModalConfirmMsg, got %T", msg)
	}
	if !confirmMsg.Confirmed || confirmMsg.Action != ModalActionTrash || confirmMsg.TargetID != "rec-123" {
		t.Fatalf("unexpected confirmMsg: %+v", confirmMsg)
	}
	if m.Active {
		t.Fatalf("expected modal to be inactive after confirmation")
	}

	// Test cancel with 'n'
	m.Open(ModalActionPurge, "Permanent Purge", "Permanently delete record? [y/N]", "rec-456")
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if cmd == nil {
		t.Fatalf("expected non-nil cmd on cancel")
	}
	cancelMsg, ok := cmd().(ModalConfirmMsg)
	if !ok || cancelMsg.Confirmed || cancelMsg.Action != ModalActionPurge {
		t.Fatalf("unexpected cancelMsg: %+v", cancelMsg)
	}
}
