package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mottamarcio/govault/internal/tui/theme"
)

// HelpCloseMsg is emitted when the help overlay is dismissed.
type HelpCloseMsg struct{}

// ModalHelp manages displaying a comprehensive hotkey reference overlay.
type ModalHelp struct {
	Theme  *theme.Theme
	Active bool
	Width  int
	Height int
}

// NewModalHelp creates a new ModalHelp instance.
func NewModalHelp(th *theme.Theme) *ModalHelp {
	return &ModalHelp{
		Theme:  th,
		Active: false,
		Width:  70,
		Height: 20,
	}
}

// Open activates the help overlay.
func (h *ModalHelp) Open() {
	h.Active = true
}

// Close dismisses the help overlay.
func (h *ModalHelp) Close() {
	h.Active = false
}

// Toggle flips the active state.
func (h *ModalHelp) Toggle() {
	h.Active = !h.Active
}

// SetDimensions sets the dimensions of the parent window for layout centering.
func (h *ModalHelp) SetDimensions(w, hDim int) {
	h.Width = w
	h.Height = hDim
}

// Init initializes the help model.
func (h *ModalHelp) Init() tea.Cmd {
	return nil
}

// Update handles closing the help overlay on Esc, ?, or q.
func (h *ModalHelp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !h.Active {
		return h, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "?", "q":
			h.Close()
			return h, func() tea.Msg { return HelpCloseMsg{} }
		}
	}
	return h, nil
}

// View renders the help overlay.
func (h *ModalHelp) View() string {
	if !h.Active {
		return ""
	}

	boxWidth := 66
	if h.Width > 0 && h.Width < boxWidth+4 {
		boxWidth = h.Width - 4
	}
	if boxWidth < 40 {
		boxWidth = 40
	}

	var sections []string
	sections = append(sections, h.Theme.HeaderBannerStyle.Render(" GOVAULT TUI KEYBOARD SHORTCUTS "))
	sections = append(sections, "")

	renderEntry := func(key, desc string) string {
		k := h.Theme.KeyHintKeyStyle.Width(14).Render(key)
		d := h.Theme.KeyHintDescStyle.Render(desc)
		return fmt.Sprintf("  %s %s", k, d)
	}

	sections = append(sections, h.Theme.TitleStyle.Render("Global & Navigation"))
	sections = append(sections, renderEntry("Tab / Shift+Tab", "Switch focus between sidebar & item list"))
	sections = append(sections, renderEntry("j / k, ↑ / ↓", "Navigate up and down lists"))
	sections = append(sections, renderEntry("h / l", "Jump between sidebar and records list"))
	sections = append(sections, renderEntry("Ctrl+L", "Instantly lock vault and zeroize session"))
	sections = append(sections, renderEntry("q / Ctrl+C", "Quit application / Lock vault"))
	sections = append(sections, renderEntry("?", "Toggle this keyboard shortcuts guide"))
	sections = append(sections, "")

	sections = append(sections, h.Theme.TitleStyle.Render("Record Operations"))
	sections = append(sections, renderEntry("a", "Add new secret (opens form editor)"))
	sections = append(sections, renderEntry("e", "Edit selected record"))
	sections = append(sections, renderEntry("d", "Move to Trash (or permanently purge if in Trash)"))
	sections = append(sections, renderEntry("r", "Restore record from Trash"))
	sections = append(sections, renderEntry("H", "View historical revision snapshots"))
	sections = append(sections, "")

	sections = append(sections, h.Theme.TitleStyle.Render("Clipboard & Secret Inspection"))
	sections = append(sections, renderEntry("v", "Toggle secret masking (reveal/hide plaintext)"))
	sections = append(sections, renderEntry("c", "Copy primary secret/password (auto-clears in 45s)"))
	sections = append(sections, renderEntry("u", "Copy username to clipboard"))
	sections = append(sections, "")

	sections = append(sections, h.Theme.TitleStyle.Render("Search & Form Editing"))
	sections = append(sections, renderEntry("/", "Start live fuzzy search query"))
	sections = append(sections, renderEntry("Ctrl+S", "Save record in form editor"))
	sections = append(sections, renderEntry("Ctrl+G", "Open embedded password/passphrase generator"))
	sections = append(sections, renderEntry("Ctrl+N", "Add custom field row in editor"))
	sections = append(sections, renderEntry("Ctrl+D", "Delete focused custom field in editor"))
	sections = append(sections, renderEntry("Esc", "Dismiss overlay / Exit search / Cancel editor"))
	sections = append(sections, "")

	sections = append(sections, h.Theme.StatusMutedStyle.Render("Press Esc or '?' to close this guide"))

	content := strings.Join(sections, "\n")
	dialogBox := h.Theme.PanelFocusStyle.
		Width(boxWidth).
		Render(content)

	if h.Width <= 0 || h.Height <= 0 {
		return dialogBox
	}

	return lipgloss.Place(
		h.Width,
		h.Height,
		lipgloss.Center,
		lipgloss.Center,
		dialogBox,
	)
}
