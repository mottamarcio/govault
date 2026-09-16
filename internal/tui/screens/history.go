package screens

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

// HistoryCloseMsg is emitted when exiting history revision view.
type HistoryCloseMsg struct{}

// HistoryModel manages browsing and inspecting historical revisions of a record.
type HistoryModel struct {
	Theme         *theme.Theme
	RecordService *service.RecordService
	RecordID      string
	RecordTitle   string
	Entries       []*domain.HistoryEntry
	Cursor        int
	Active        bool
	Width         int
	Height        int
}

// NewHistoryModel creates a new HistoryModel.
func NewHistoryModel(th *theme.Theme, rs *service.RecordService) *HistoryModel {
	return &HistoryModel{
		Theme:         th,
		RecordService: rs,
		RecordID:      "",
		RecordTitle:   "",
		Entries:       nil,
		Cursor:        0,
		Active:        false,
		Width:         70,
		Height:        20,
	}
}

// LoadRecordHistory fetches past revisions for the specified record.
func (m *HistoryModel) LoadRecordHistory(recordID, recordTitle string) error {
	m.RecordID = recordID
	m.RecordTitle = recordTitle
	m.Cursor = 0
	m.Active = true

	entries, err := m.RecordService.ListHistory(context.Background(), recordID)
	if err != nil {
		return err
	}
	m.Entries = entries
	return nil
}

// SetDimensions updates history modal dimensions.
func (m *HistoryModel) SetDimensions(w, h int) {
	m.Width = w
	m.Height = h
}

// Init initializes the history model.
func (m *HistoryModel) Init() tea.Cmd {
	return nil
}

// Update handles navigation within history timeline.
func (m *HistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q", "H":
			m.Active = false
			return m, func() tea.Msg { return HistoryCloseMsg{} }
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Entries)-1 {
				m.Cursor++
			}
		}
	}
	return m, nil
}

// View renders the history revision browser.
func (m *HistoryModel) View() string {
	if !m.Active {
		return ""
	}

	var lines []string
	lines = append(lines, m.Theme.TitleStyle.Render(fmt.Sprintf("📜 Revision History: %s", m.RecordTitle)), "")

	if len(m.Entries) == 0 {
		lines = append(lines, m.Theme.StatusMutedStyle.Render("No historical revisions found (Record is at revision v1)."), "")
		lines = append(lines, m.Theme.StatusMutedStyle.Render("Press Esc or 'H' to return to dashboard."))
		return m.Theme.NoticeBoxStyle.
			Width(m.Width).
			Height(m.Height).
			Render(strings.Join(lines, "\n"))
	}

	for i, entry := range m.Entries {
		prefix := "  "
		if i == m.Cursor {
			prefix = "▸ "
		}

		entryLine := fmt.Sprintf("%sv%d  •  Archived: %s",
			prefix,
			entry.Version,
			entry.ArchivedAt.Format("2006-01-02 15:04:05"),
		)

		if i == m.Cursor {
			lines = append(lines, m.Theme.KeyHintKeyStyle.Render(entryLine))
		} else {
			lines = append(lines, m.Theme.StatusMutedStyle.Render(entryLine))
		}
	}

	// Show selected revision snapshot preview
	if m.Cursor >= 0 && m.Cursor < len(m.Entries) {
		selected := m.Entries[m.Cursor]
		lines = append(lines, "", m.Theme.SubtitleStyle.Render(fmt.Sprintf("--- Revision v%d Snapshot ---", selected.Version)))
		switch p := selected.Payload.(type) {
		case *domain.LoginPayload:
			lines = append(lines, fmt.Sprintf("Username: %s | Password: %s", p.Username, p.Password))
		case domain.LoginPayload:
			lines = append(lines, fmt.Sprintf("Username: %s | Password: %s", p.Username, p.Password))
		case *domain.NotePayload:
			lines = append(lines, fmt.Sprintf("Note: %s", p.Content))
		case domain.NotePayload:
			lines = append(lines, fmt.Sprintf("Note: %s", p.Content))
		case *domain.APIKeyPayload:
			lines = append(lines, fmt.Sprintf("Service: %s | API Key: %s", p.Service, p.Secret))
		case domain.APIKeyPayload:
			lines = append(lines, fmt.Sprintf("Service: %s | API Key: %s", p.Service, p.Secret))
		}
	}

	lines = append(lines, "", m.Theme.StatusMutedStyle.Render("Press 'j'/'k' to navigate  •  'Esc' or 'H' to return"))

	content := strings.Join(lines, "\n")
	return m.Theme.PanelFocusStyle.
		Width(m.Width).
		Height(m.Height).
		Render(content)
}
