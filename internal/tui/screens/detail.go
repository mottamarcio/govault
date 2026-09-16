package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

// DetailModel formats and displays record attributes with masked/unmasked toggle.
type DetailModel struct {
	Theme        *theme.Theme
	Record       *domain.Record
	ShowSecret   bool
	Width        int
	Height       int
}

// NewDetailModel creates a new DetailModel.
func NewDetailModel(th *theme.Theme) *DetailModel {
	return &DetailModel{
		Theme:      th,
		Record:     nil,
		ShowSecret: false,
		Width:      50,
		Height:     20,
	}
}

// SetRecord updates the displayed record and resets secret visibility.
func (m *DetailModel) SetRecord(r *domain.Record) {
	m.Record = r
	m.ShowSecret = false
}

// SetDimensions updates panel dimensions.
func (m *DetailModel) SetDimensions(w, h int) {
	m.Width = w
	m.Height = h
}

// ToggleSecret toggles masked/plaintext display of sensitive fields.
func (m *DetailModel) ToggleSecret() {
	m.ShowSecret = !m.ShowSecret
}

// Init initializes the detail model.
func (m *DetailModel) Init() tea.Cmd {
	return nil
}

// Update handles detail view events.
func (m *DetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "v" {
			m.ToggleSecret()
			return m, nil
		}
	}
	return m, nil
}

// View renders the formatted secret detail card.
func (m *DetailModel) View() string {
	if m.Record == nil {
		return m.Theme.PanelStyle.
			Width(m.Width).
			Height(m.Height).
			Align(lipgloss.Center, lipgloss.Center).
			Render(m.Theme.StatusMutedStyle.Render("No record selected"))
	}

	var lines []string

	// Header
	header := fmt.Sprintf("🏷️  %s  [%s]", m.Record.Title, strings.ToUpper(string(m.Record.Type)))
	lines = append(lines, m.Theme.TitleStyle.Render(header))
	lines = append(lines, m.Theme.StatusMutedStyle.Render(fmt.Sprintf("ID: %s  •  v%d  •  Updated: %s",
		m.Record.ID, m.Record.Version, m.Record.UpdatedAt.Format("2006-01-02 15:04:05"))))

	if len(m.Record.Tags) > 0 {
		var tagBadges []string
		for _, t := range m.Record.Tags {
			tagBadges = append(tagBadges, m.Theme.SubtitleStyle.Render("#"+t))
		}
		lines = append(lines, strings.Join(tagBadges, " "))
	}
	lines = append(lines, "")

	// Render typed payload fields
	switch p := m.Record.Payload.(type) {
	case *domain.LoginPayload:
		lines = append(lines, m.renderLoginFields(p.Username, p.Password, p.URI, p.Notes))
	case domain.LoginPayload:
		lines = append(lines, m.renderLoginFields(p.Username, p.Password, p.URI, p.Notes))

	case *domain.NotePayload:
		lines = append(lines, m.renderNoteFields(p.Content))
	case domain.NotePayload:
		lines = append(lines, m.renderNoteFields(p.Content))

	case *domain.APIKeyPayload:
		lines = append(lines, m.renderAPIKeyFields(p.Service, p.Secret, p.Key, p.Endpoint, p.Notes))
	case domain.APIKeyPayload:
		lines = append(lines, m.renderAPIKeyFields(p.Service, p.Secret, p.Key, p.Endpoint, p.Notes))

	case *domain.CustomPayload:
		lines = append(lines, m.renderCustomFields(p.Fields, p.Notes)...)
	case domain.CustomPayload:
		lines = append(lines, m.renderCustomFields(p.Fields, p.Notes)...)
	}

	lines = append(lines, "", m.Theme.StatusMutedStyle.Render("Press 'v' to toggle secret visibility  •  'H' for revision history"))

	content := strings.Join(lines, "\n")
	return m.Theme.PanelStyle.
		Width(m.Width).
		Height(m.Height).
		Render(content)
}

func (m *DetailModel) renderLoginFields(user, pass, uri, notes string) string {
	var lines []string
	lines = append(lines, m.renderField("Username / Email", user))
	pwDisplay := "••••••••••••••••"
	if m.ShowSecret {
		pwDisplay = pass
	}
	lines = append(lines, m.renderField("Password", pwDisplay))
	if uri != "" {
		lines = append(lines, m.renderField("URI", uri))
	}
	if notes != "" {
		lines = append(lines, "", m.Theme.SubtitleStyle.Render("Notes:"), notes)
	}
	return strings.Join(lines, "\n")
}

func (m *DetailModel) renderNoteFields(content string) string {
	bodyDisplay := "••••••••••••••••"
	if m.ShowSecret {
		bodyDisplay = content
	}
	return fmt.Sprintf("%s\n%s", m.Theme.SubtitleStyle.Render("Note Content:"), bodyDisplay)
}

func (m *DetailModel) renderAPIKeyFields(service, secret, key, endpoint, notes string) string {
	var lines []string
	lines = append(lines, m.renderField("Service", service))
	secDisplay := "••••••••••••••••"
	if m.ShowSecret {
		secDisplay = secret
	}
	lines = append(lines, m.renderField("Secret", secDisplay))
	if key != "" {
		lines = append(lines, m.renderField("Key / ID", key))
	}
	if endpoint != "" {
		lines = append(lines, m.renderField("Endpoint", endpoint))
	}
	if notes != "" {
		lines = append(lines, "", m.Theme.SubtitleStyle.Render("Notes:"), notes)
	}
	return strings.Join(lines, "\n")
}

func (m *DetailModel) renderCustomFields(fields []domain.Field, notes string) []string {
	var lines []string
	lines = append(lines, m.Theme.SubtitleStyle.Render("Custom Fields:"))
	for _, f := range fields {
		val := f.Value
		if f.Masked && !m.ShowSecret {
			val = "••••••••••••"
		}
		lines = append(lines, m.renderField(f.Key, val))
	}
	if notes != "" {
		lines = append(lines, "", m.Theme.SubtitleStyle.Render("Notes:"), notes)
	}
	return lines
}

func (m *DetailModel) renderField(label, value string) string {
	lbl := m.Theme.KeyHintKeyStyle.Render(label + ":")
	val := lipgloss.NewStyle().Foreground(m.Theme.Text).Render(value)
	return fmt.Sprintf("%-20s %s", lbl, val)
}
