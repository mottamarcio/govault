package screens

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mottamarcio/govault/internal/crypto/kdf"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

// InitSuccessMsg is sent when vault initialization and auto-unlock succeed.
type InitSuccessMsg struct {
	Session *service.Session
}

// InitErrMsg is sent when vault initialization fails.
type InitErrMsg struct {
	Err error
}

// InitFocus tracks focus between password and confirm password inputs.
type InitFocus int

const (
	InitFocusPassword InitFocus = iota
	InitFocusConfirm
)

// InitModel manages the interactive first-run vault initialization screen.
type InitModel struct {
	Theme        *theme.Theme
	VaultService *service.VaultService
	PassInput    textinput.Model
	ConfirmInput textinput.Model
	Focus        InitFocus
	VaultPath    string
	ErrMsg       string
	Loading      bool
	Width        int
	Height       int
}

// NewInitModel creates a new InitModel.
func NewInitModel(th *theme.Theme, vs *service.VaultService, vaultPath string) *InitModel {
	pInput := textinput.New()
	pInput.Placeholder = "Create master password (min 8 chars)..."
	pInput.EchoMode = textinput.EchoPassword
	pInput.EchoCharacter = '•'
	pInput.Focus()
	pInput.Prompt = "🔑 Master Password:  "
	pInput.PromptStyle = th.InputPromptStyle
	pInput.TextStyle = lipgloss.NewStyle().Foreground(th.Text)
	pInput.CharLimit = 128
	pInput.Width = 40

	cInput := textinput.New()
	cInput.Placeholder = "Confirm master password..."
	cInput.EchoMode = textinput.EchoPassword
	cInput.EchoCharacter = '•'
	cInput.Blur()
	cInput.Prompt = "🔒 Confirm Password: "
	cInput.PromptStyle = th.InputPromptStyle
	cInput.TextStyle = lipgloss.NewStyle().Foreground(th.Text)
	cInput.CharLimit = 128
	cInput.Width = 40

	return &InitModel{
		Theme:        th,
		VaultService: vs,
		PassInput:    pInput,
		ConfirmInput: cInput,
		Focus:        InitFocusPassword,
		VaultPath:    vaultPath,
		ErrMsg:       "",
		Loading:      false,
		Width:        80,
		Height:       24,
	}
}

// Init initializes the model.
func (m *InitModel) Init() tea.Cmd {
	return textinput.Blink
}

// SetDimensions updates window dimensions.
func (m *InitModel) SetDimensions(w, h int) {
	m.Width = w
	m.Height = h
}

// ResetInput clears input fields and error state.
func (m *InitModel) ResetInput() {
	m.PassInput.Reset()
	m.ConfirmInput.Reset()
	m.Focus = InitFocusPassword
	m.PassInput.Focus()
	m.ConfirmInput.Blur()
	m.ErrMsg = ""
	m.Loading = false
}

// Update handles Bubble Tea messages for InitModel.
func (m *InitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyTab, tea.KeyDown:
			if m.Focus == InitFocusPassword {
				m.Focus = InitFocusConfirm
				m.PassInput.Blur()
				m.ConfirmInput.Focus()
			} else {
				m.Focus = InitFocusPassword
				m.ConfirmInput.Blur()
				m.PassInput.Focus()
			}
			return m, nil

		case tea.KeyShiftTab, tea.KeyUp:
			if m.Focus == InitFocusConfirm {
				m.Focus = InitFocusPassword
				m.ConfirmInput.Blur()
				m.PassInput.Focus()
			} else {
				m.Focus = InitFocusConfirm
				m.PassInput.Blur()
				m.ConfirmInput.Focus()
			}
			return m, nil

		case tea.KeyEnter:
			pass := m.PassInput.Value()
			confirm := m.ConfirmInput.Value()

			if len(pass) < 8 {
				m.ErrMsg = "Password must be at least 8 characters long"
				return m, nil
			}
			if pass != confirm {
				m.ErrMsg = "Passwords do not match"
				return m, nil
			}

			m.Loading = true
			m.ErrMsg = ""

			return m, func() tea.Msg {
				pwBytes := []byte(pass)
				defer kdf.Zeroize(pwBytes)

				ctx := context.Background()
				if err := m.VaultService.Init(ctx, pass); err != nil {
					return InitErrMsg{Err: err}
				}

				session, err := m.VaultService.Unlock(ctx, pass)
				if err != nil {
					return InitErrMsg{Err: err}
				}

				return InitSuccessMsg{Session: session}
			}

		case tea.KeyEsc:
			m.ResetInput()
			return m, nil
		}

	case InitErrMsg:
		m.Loading = false
		m.ErrMsg = "Initialization failed: " + msg.Err.Error()
		return m, nil

	case InitSuccessMsg:
		m.Loading = false
		m.ResetInput()
		return m, nil
	}

	var cmd tea.Cmd
	if m.Focus == InitFocusPassword {
		m.PassInput, cmd = m.PassInput.Update(msg)
	} else {
		m.ConfirmInput, cmd = m.ConfirmInput.Update(msg)
	}
	return m, cmd
}

func (m *InitModel) renderStrengthGauge(entropy float64, strength service.StrengthLevel) string {
	var badge string
	switch strength {
	case service.StrengthVeryWeak, service.StrengthWeak:
		badge = m.Theme.ErrorBadgeStyle.Render(" " + string(strength) + " ")
	case service.StrengthFair:
		badge = m.Theme.WarningBadgeStyle.Render(" " + string(strength) + " ")
	case service.StrengthStrong, service.StrengthVeryStrong:
		badge = m.Theme.SuccessBadgeStyle.Render(" " + string(strength) + " ")
	default:
		badge = string(strength)
	}
	return fmt.Sprintf("Entropy: %.1f bits  %s", entropy, badge)
}

// View renders the first-run initialization screen UI.
func (m *InitModel) View() string {
	title := m.Theme.TitleStyle.Render("GOVAULT — FIRST-TIME SETUP")
	subtitle := m.Theme.SubtitleStyle.Render(fmt.Sprintf("Vault File: %s", m.VaultPath))

	securityNotice := lipgloss.NewStyle().
		Foreground(m.Theme.TextMuted).
		Width(60).
		Align(lipgloss.Center).
		Render("This master password derives the master key that encrypts your entire vault. It is never stored and cannot be recovered if lost.")

	passVal := m.PassInput.Value()
	var strengthLine string
	if len(passVal) > 0 {
		entropy := service.CalculateEntropy(passVal)
		strength := service.EvaluateStrength(entropy)
		strengthLine = m.renderStrengthGauge(entropy, strength)
	} else {
		strengthLine = m.Theme.StatusMutedStyle.Render("Password minimum length: 8 characters")
	}

	var statusLine string
	if m.Loading {
		statusLine = m.Theme.WarningBadgeStyle.Render("Deriving Argon2id encryption envelope & initializing database...")
	} else if m.ErrMsg != "" {
		statusLine = m.Theme.ErrorBadgeStyle.Render("✕ " + m.ErrMsg)
	} else {
		statusLine = m.Theme.StatusMutedStyle.Render("Tab/Shift+Tab switch field • Enter submit • q / Ctrl+C quit")
	}

	formBox := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		subtitle,
		"",
		securityNotice,
		"",
		m.PassInput.View(),
		m.ConfirmInput.View(),
		"",
		strengthLine,
		"",
		statusLine,
	)

	styledBox := m.Theme.InputBoxStyle.Render(formBox)

	return lipgloss.Place(
		m.Width,
		m.Height-1,
		lipgloss.Center,
		lipgloss.Center,
		styledBox,
	)
}
