package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mottamarcio/govault/internal/tui/theme"
)

// ModalAction defines the type of action requiring confirmation.
type ModalAction string

const (
	ModalActionTrash   ModalAction = "trash"
	ModalActionPurge   ModalAction = "purge"
	ModalActionRestore ModalAction = "restore"
	ModalActionDiscard ModalAction = "discard"
)

// ModalConfirmMsg is emitted when a confirmation modal is accepted or rejected.
type ModalConfirmMsg struct {
	Action    ModalAction
	Confirmed bool
	TargetID  string
}

// ModalConfirm manages a floating confirmation dialog.
type ModalConfirm struct {
	Theme    *theme.Theme
	Active   bool
	Action   ModalAction
	Title    string
	Prompt   string
	TargetID string
	Width    int
	Height   int
}

// NewModalConfirm initializes a confirmation modal.
func NewModalConfirm(th *theme.Theme) *ModalConfirm {
	return &ModalConfirm{
		Theme:    th,
		Active:   false,
		Action:   "",
		Title:    "",
		Prompt:   "",
		TargetID: "",
		Width:    60,
		Height:   10,
	}
}

// Open activates the confirmation dialog with the given action and text.
func (m *ModalConfirm) Open(action ModalAction, title, prompt, targetID string) {
	m.Action = action
	m.Title = title
	m.Prompt = prompt
	m.TargetID = targetID
	m.Active = true
}

// Close deactivates the modal.
func (m *ModalConfirm) Close() {
	m.Active = false
	m.Action = ""
	m.Title = ""
	m.Prompt = ""
	m.TargetID = ""
}

// SetDimensions sets the dimensions of the parent window for layout centering.
func (m *ModalConfirm) SetDimensions(w, h int) {
	m.Width = w
	m.Height = h
}

// Init initializes the modal.
func (m *ModalConfirm) Init() tea.Cmd {
	return nil
}

// Update handles modal keyboard responses (y / n / Esc).
func (m *ModalConfirm) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			act := m.Action
			id := m.TargetID
			m.Close()
			return m, func() tea.Msg {
				return ModalConfirmMsg{
					Action:    act,
					Confirmed: true,
					TargetID:  id,
				}
			}

		case "n", "N", "esc", "q":
			act := m.Action
			id := m.TargetID
			m.Close()
			return m, func() tea.Msg {
				return ModalConfirmMsg{
					Action:    act,
					Confirmed: false,
					TargetID:  id,
				}
			}
		}
	}

	return m, nil
}

// View renders the confirmation dialog overlay.
func (m *ModalConfirm) View() string {
	if !m.Active {
		return ""
	}

	boxWidth := 54
	if m.Width > 0 && m.Width < boxWidth+4 {
		boxWidth = m.Width - 4
	}
	if boxWidth < 30 {
		boxWidth = 30
	}

	header := m.Theme.WarningBadgeStyle.Render(" CONFIRMATION REQUIRED ")
	if m.Action == ModalActionPurge {
		header = m.Theme.ErrorBadgeStyle.Render(" PERMANENT DELETION ")
	}

	var lines []string
	lines = append(lines, header, "")
	if m.Title != "" {
		lines = append(lines, m.Theme.TitleStyle.Render(m.Title))
	}
	lines = append(lines, lipgloss.NewStyle().Foreground(m.Theme.Text).Render(m.Prompt))
	lines = append(lines, "")

	btnConfirm := m.Theme.KeyHintKeyStyle.Render("[y] Yes / Confirm")
	btnCancel := m.Theme.StatusMutedStyle.Render("[n / Esc] Cancel")
	buttons := fmt.Sprintf("%s    %s", btnConfirm, btnCancel)
	lines = append(lines, buttons)

	content := strings.Join(lines, "\n")
	dialogBox := m.Theme.NoticeBoxStyle.
		Width(boxWidth).
		Render(content)

	if m.Width <= 0 || m.Height <= 0 {
		return dialogBox
	}

	return lipgloss.Place(
		m.Width,
		m.Height,
		lipgloss.Center,
		lipgloss.Center,
		dialogBox,
	)
}
