package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mottamarcio/govault/internal/crypto/kdf"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

// UnlockSuccessMsg is sent when vault unlock succeeds.
type UnlockSuccessMsg struct {
	Session *service.Session
}

// UnlockErrMsg is sent when unlock authentication fails.
type UnlockErrMsg struct {
	Err error
}

// UnlockModel manages the interactive master password input screen.
type UnlockModel struct {
	Theme        *theme.Theme
	VaultService *service.VaultService
	Input        textinput.Model
	VaultPath    string
	ErrMsg       string
	Loading      bool
	Width        int
	Height       int
}

// NewUnlockModel creates a new UnlockModel.
func NewUnlockModel(th *theme.Theme, vs *service.VaultService, vaultPath string) *UnlockModel {
	ti := textinput.New()
	ti.Placeholder = "Enter master password..."
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()
	ti.Prompt = "🔑 Master Password: "
	ti.PromptStyle = th.InputPromptStyle
	ti.TextStyle = lipgloss.NewStyle().Foreground(th.Text)
	ti.CharLimit = 128
	ti.Width = 40

	return &UnlockModel{
		Theme:        th,
		VaultService: vs,
		Input:        ti,
		VaultPath:    vaultPath,
		ErrMsg:       "",
		Loading:      false,
		Width:        80,
		Height:       24,
	}
}

// Init initializes the unlock model.
func (m *UnlockModel) Init() tea.Cmd {
	return textinput.Blink
}

// SetDimensions updates window dimensions.
func (m *UnlockModel) SetDimensions(w, h int) {
	m.Width = w
	m.Height = h
}

// ResetInput clears the input field and error state.
func (m *UnlockModel) ResetInput() {
	m.Input.Reset()
	m.ErrMsg = ""
	m.Loading = false
	m.Input.Focus()
}

// Update handles Bubble Tea messages for UnlockModel.
func (m *UnlockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			password := m.Input.Value()
			if strings.TrimSpace(password) == "" {
				m.ErrMsg = "Password cannot be empty"
				return m, nil
			}
			m.Loading = true
			m.ErrMsg = ""

			// Execute unlock asynchronously
			return m, func() tea.Msg {
				pwBytes := []byte(password)
				defer kdf.Zeroize(pwBytes)

				session, err := m.VaultService.Unlock(context.Background(), password)
				if err != nil {
					return UnlockErrMsg{Err: err}
				}
				return UnlockSuccessMsg{Session: session}
			}

		case tea.KeyEsc:
			m.ResetInput()
			return m, nil
		}

	case UnlockErrMsg:
		m.Loading = false
		m.ErrMsg = "Authentication failed: " + msg.Err.Error()
		m.Input.Reset()
		m.Input.Focus()
		return m, nil

	case UnlockSuccessMsg:
		m.Loading = false
		m.Input.Reset()
		return m, nil
	}

	var cmd tea.Cmd
	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}

// View renders the unlock screen UI.
func (m *UnlockModel) View() string {
	title := m.Theme.TitleStyle.Render("GOVAULT — SECURE OFFLINE VAULT")
	subtitle := m.Theme.SubtitleStyle.Render(fmt.Sprintf("Vault File: %s", m.VaultPath))

	var statusLine string
	if m.Loading {
		statusLine = m.Theme.WarningBadgeStyle.Render("Authenticating Argon2id KDF...")
	} else if m.ErrMsg != "" {
		statusLine = m.Theme.ErrorBadgeStyle.Render("✕ " + m.ErrMsg)
	} else {
		statusLine = m.Theme.StatusMutedStyle.Render("Press Enter to unlock • Ctrl+C to quit")
	}

	formBox := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		m.Input.View(),
		"",
		statusLine,
	)

	styledBox := m.Theme.InputBoxStyle.Render(formBox)

	return lipgloss.Place(
		m.Width,
		m.Height-1, // Reserve line for statusbar if present
		lipgloss.Center,
		lipgloss.Center,
		styledBox,
	)
}
