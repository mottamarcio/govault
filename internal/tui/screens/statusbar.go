package screens

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/mottamarcio/govault/internal/tui/theme"
)

// StatusBarModel renders the bottom status bar across TUI screens.
type StatusBarModel struct {
	Theme          *theme.Theme
	VaultID        string
	RecordCount    int
	RemainingLock  time.Duration
	LockTimerShown bool
	KeyHints       []string
	Width          int
}

// NewStatusBar creates a new StatusBarModel.
func NewStatusBar(th *theme.Theme, vaultID string) *StatusBarModel {
	return &StatusBarModel{
		Theme:          th,
		VaultID:        vaultID,
		RecordCount:    0,
		RemainingLock:  5 * time.Minute,
		LockTimerShown: false,
		KeyHints:       []string{"?: Help", "q: Quit"},
		Width:          80,
	}
}

// SetWidth updates the status bar rendered width.
func (m *StatusBarModel) SetWidth(w int) {
	m.Width = w
}

// SetLockRemaining updates the auto-lock countdown timer.
func (m *StatusBarModel) SetLockRemaining(rem time.Duration, shown bool) {
	m.RemainingLock = rem
	m.LockTimerShown = shown
}

// SetRecordCount updates total record count displayed.
func (m *StatusBarModel) SetRecordCount(count int) {
	m.RecordCount = count
}

// SetKeyHints updates the keyboard shortcut hints displayed in the status bar.
func (m *StatusBarModel) SetKeyHints(hints []string) {
	m.KeyHints = hints
}

// View renders the status bar.
func (m *StatusBarModel) View() string {
	if m.Width <= 0 {
		return ""
	}

	left := m.Theme.StatusBadgeStyle.Render("GoVault") + " " +
		m.Theme.StatusMutedStyle.Render(fmt.Sprintf("[%s]", m.VaultID))

	if m.LockTimerShown {
		left += " " + m.Theme.StatusMutedStyle.Render(fmt.Sprintf("(%d records)", m.RecordCount))
	}

	var right string
	if m.LockTimerShown {
		mins := int(m.RemainingLock.Minutes())
		secs := int(m.RemainingLock.Seconds()) % 60
		timerText := fmt.Sprintf("Auto-lock: %dm %02ds", mins, secs)
		right += m.Theme.StatusTimerStyle.Render(timerText) + "  "
	}

	for i, hint := range m.KeyHints {
		if i > 0 {
			right += " "
		}
		right += m.Theme.KeyHintDescStyle.Render(hint)
	}

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spaceWidth := m.Width - leftWidth - rightWidth - 2 // 2 for padding
	if spaceWidth < 1 {
		spaceWidth = 1
	}

	spaces := lipgloss.NewStyle().Width(spaceWidth).Render(" ")
	content := lipgloss.JoinHorizontal(lipgloss.Top, left, spaces, right)

	return m.Theme.StatusBarStyle.Width(m.Width).Render(content)
}
