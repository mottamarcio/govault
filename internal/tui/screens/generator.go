package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

// GeneratorMode specifies standard password vs diceware passphrase generation.
type GeneratorMode int

const (
	GeneratorModePassword GeneratorMode = iota
	GeneratorModePassphrase
)

// GeneratorCloseMsg is emitted when generator modal is dismissed without saving.
type GeneratorCloseMsg struct{}

// GeneratorResultMsg is emitted when a generated password is confirmed.
type GeneratorResultMsg struct {
	Secret string
}

// GeneratorControl identifies the focused setting/control index in the modal.
type GeneratorControl int

const (
	GenCtrlMode GeneratorControl = iota
	GenCtrlLengthWords
	GenCtrlOpt1 // Password: Upper, Passphrase: Delimiter
	GenCtrlOpt2 // Password: Lower, Passphrase: Capitalize
	GenCtrlOpt3 // Password: Digits
	GenCtrlOpt4 // Password: Symbols
	GenCtrlOpt5 // Password: Exclude Ambiguous
	GenCtrlRegenerate
	GenCtrlAccept
	GenCtrlCancel
	GenCtrlMaxCount
)

// GeneratorModel manages the interactive password & passphrase generation modal.
type GeneratorModel struct {
	Theme   *theme.Theme
	Active  bool
	Mode    GeneratorMode
	Control GeneratorControl

	// Password Options
	PassLength       int
	IncludeUpper     bool
	IncludeLower     bool
	IncludeDigits    bool
	IncludeSymbols   bool
	ExcludeAmbiguous bool

	// Passphrase Options
	WordCount    int
	DelimiterIdx int
	Capitalize   bool

	CurrentGenerated string
	CurrentEntropy   float64
	CurrentStrength  service.StrengthLevel
	ErrorMsg         string

	Width  int
	Height int
}

var delimiters = []string{"-", " ", "_", "."}

// NewGeneratorModel initializes a generator model with secure defaults.
func NewGeneratorModel(th *theme.Theme) *GeneratorModel {
	m := &GeneratorModel{
		Theme:            th,
		Active:           false,
		Mode:             GeneratorModePassword,
		Control:          GenCtrlMode,
		PassLength:       20,
		IncludeUpper:     true,
		IncludeLower:     true,
		IncludeDigits:    true,
		IncludeSymbols:   true,
		ExcludeAmbiguous: false,
		WordCount:        5,
		DelimiterIdx:     0,
		Capitalize:       false,
		CurrentGenerated: "",
		CurrentEntropy:   0,
		CurrentStrength:  service.StrengthVeryWeak,
		Width:            70,
		Height:           22,
	}
	m.Regenerate()
	return m
}

// Open activates the modal and generates a fresh secret.
func (m *GeneratorModel) Open() {
	m.Active = true
	m.Control = GenCtrlMode
	m.Regenerate()
}

// Close deactivates the modal.
func (m *GeneratorModel) Close() {
	m.Active = false
	// Zeroize buffer
	m.CurrentGenerated = ""
}

// SetDimensions sets viewport dimensions for centering.
func (m *GeneratorModel) SetDimensions(w, h int) {
	m.Width = w
	m.Height = h
}

// Init initializes the generator.
func (m *GeneratorModel) Init() tea.Cmd {
	return nil
}

// Regenerate creates a new random secret based on current options.
func (m *GeneratorModel) Regenerate() {
	m.ErrorMsg = ""
	if m.Mode == GeneratorModePassword {
		opts := service.PasswordOptions{
			Length:           m.PassLength,
			IncludeUpper:     m.IncludeUpper,
			IncludeLower:     m.IncludeLower,
			IncludeDigits:    m.IncludeDigits,
			IncludeSymbols:   m.IncludeSymbols,
			ExcludeAmbiguous: m.ExcludeAmbiguous,
		}
		secret, err := service.GeneratePassword(opts)
		if err != nil {
			m.ErrorMsg = err.Error()
			return
		}
		m.CurrentGenerated = secret
	} else {
		delim := delimiters[m.DelimiterIdx%len(delimiters)]
		opts := service.PassphraseOptions{
			WordCount:  m.WordCount,
			Delimiter:  delim,
			Capitalize: m.Capitalize,
			Wordlist:   service.DefaultWordlist,
		}
		secret, err := service.GeneratePassphrase(opts)
		if err != nil {
			m.ErrorMsg = err.Error()
			return
		}
		m.CurrentGenerated = secret
	}

	m.CurrentEntropy = service.CalculateEntropy(m.CurrentGenerated)
	m.CurrentStrength = service.EvaluateStrength(m.CurrentEntropy)
}

// Update processes keyboard navigation and toggles.
func (m *GeneratorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.Active {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.Close()
			return m, func() tea.Msg { return GeneratorCloseMsg{} }

		case tea.KeyEnter:
			if m.Control == GenCtrlCancel {
				m.Close()
				return m, func() tea.Msg { return GeneratorCloseMsg{} }
			}
			if m.Control == GenCtrlRegenerate {
				m.Regenerate()
				return m, nil
			}
			// Accept secret
			sec := m.CurrentGenerated
			m.Close()
			return m, func() tea.Msg { return GeneratorResultMsg{Secret: sec} }

		case tea.KeyTab, tea.KeyDown:
			m.nextControl(1)
			return m, nil

		case tea.KeyShiftTab, tea.KeyUp:
			m.nextControl(-1)
			return m, nil

		case tea.KeyLeft:
			m.adjustControl(-1)
			return m, nil

		case tea.KeyRight:
			m.adjustControl(1)
			return m, nil

		case tea.KeySpace:
			m.toggleOrAction()
			return m, nil

		case tea.KeyRunes:
			switch msg.String() {
			case "r", "R":
				m.Regenerate()
				return m, nil
			case "q":
				m.Close()
				return m, func() tea.Msg { return GeneratorCloseMsg{} }
			}
		}
	}

	return m, nil
}

func (m *GeneratorModel) maxControls() int {
	if m.Mode == GeneratorModePassword {
		return int(GenCtrlMaxCount)
	}
	// In Passphrase mode, we only have Mode, Words, Delimiter, Capitalize, Regenerate, Accept, Cancel (7 items)
	return 7
}

func (m *GeneratorModel) nextControl(delta int) {
	max := m.maxControls()
	cur := int(m.Control)
	cur = (cur + delta + max) % max
	m.Control = GeneratorControl(cur)
}

func (m *GeneratorModel) adjustControl(delta int) {
	switch m.Control {
	case GenCtrlMode:
		if m.Mode == GeneratorModePassword {
			m.Mode = GeneratorModePassphrase
		} else {
			m.Mode = GeneratorModePassword
		}
		m.Regenerate()

	case GenCtrlLengthWords:
		if m.Mode == GeneratorModePassword {
			m.PassLength += delta
			if m.PassLength < 8 {
				m.PassLength = 8
			}
			if m.PassLength > 64 {
				m.PassLength = 64
			}
		} else {
			m.WordCount += delta
			if m.WordCount < 3 {
				m.WordCount = 3
			}
			if m.WordCount > 10 {
				m.WordCount = 10
			}
		}
		m.Regenerate()

	case GenCtrlOpt1:
		if m.Mode == GeneratorModePassphrase {
			m.DelimiterIdx = (m.DelimiterIdx + delta + len(delimiters)) % len(delimiters)
			m.Regenerate()
		} else {
			m.IncludeUpper = !m.IncludeUpper
			m.Regenerate()
		}

	case GenCtrlOpt2:
		if m.Mode == GeneratorModePassphrase {
			m.Capitalize = !m.Capitalize
			m.Regenerate()
		} else {
			m.IncludeLower = !m.IncludeLower
			m.Regenerate()
		}
	}
}

func (m *GeneratorModel) toggleOrAction() {
	switch m.Control {
	case GenCtrlMode:
		if m.Mode == GeneratorModePassword {
			m.Mode = GeneratorModePassphrase
		} else {
			m.Mode = GeneratorModePassword
		}
		m.Regenerate()

	case GenCtrlOpt1:
		if m.Mode == GeneratorModePassword {
			m.IncludeUpper = !m.IncludeUpper
			m.Regenerate()
		} else {
			m.DelimiterIdx = (m.DelimiterIdx + 1) % len(delimiters)
			m.Regenerate()
		}

	case GenCtrlOpt2:
		if m.Mode == GeneratorModePassword {
			m.IncludeLower = !m.IncludeLower
			m.Regenerate()
		} else {
			m.Capitalize = !m.Capitalize
			m.Regenerate()
		}

	case GenCtrlOpt3:
		if m.Mode == GeneratorModePassword {
			m.IncludeDigits = !m.IncludeDigits
			m.Regenerate()
		}

	case GenCtrlOpt4:
		if m.Mode == GeneratorModePassword {
			m.IncludeSymbols = !m.IncludeSymbols
			m.Regenerate()
		}

	case GenCtrlOpt5:
		if m.Mode == GeneratorModePassword {
			m.ExcludeAmbiguous = !m.ExcludeAmbiguous
			m.Regenerate()
		}

	case GenCtrlRegenerate:
		m.Regenerate()
	}
}

// View renders the generator dialog.
func (m *GeneratorModel) View() string {
	if !m.Active {
		return ""
	}

	boxWidth := 64
	if m.Width > 0 && m.Width < boxWidth+4 {
		boxWidth = m.Width - 4
	}
	if boxWidth < 40 {
		boxWidth = 40
	}

	var lines []string
	lines = append(lines, m.Theme.HeaderBannerStyle.Render(" PASSWORD & PASSPHRASE GENERATOR "), "")

	// Mode selector
	passMode := "[○ Password]"
	phraseMode := "[○ Passphrase]"
	if m.Mode == GeneratorModePassword {
		passMode = "[● Password]"
	} else {
		phraseMode = "[● Passphrase]"
	}
	modeStr := fmt.Sprintf("Mode:  %s   %s", passMode, phraseMode)
	if m.Control == GenCtrlMode {
		lines = append(lines, m.Theme.KeyHintKeyStyle.Render("▸ "+modeStr))
	} else {
		lines = append(lines, "  "+modeStr)
	}
	lines = append(lines, "")

	renderControl := func(ctrl GeneratorControl, text string) string {
		if m.Control == ctrl {
			return m.Theme.KeyHintKeyStyle.Render("▸ " + text)
		}
		return "  " + text
	}

	if m.Mode == GeneratorModePassword {
		// Length
		lines = append(lines, renderControl(GenCtrlLengthWords, fmt.Sprintf("Length:  ◀ %2d ▶  (8 - 64)", m.PassLength)))
		// Checkboxes
		lines = append(lines, renderControl(GenCtrlOpt1, fmt.Sprintf("[%s] Uppercase (A-Z)", check(m.IncludeUpper))))
		lines = append(lines, renderControl(GenCtrlOpt2, fmt.Sprintf("[%s] Lowercase (a-z)", check(m.IncludeLower))))
		lines = append(lines, renderControl(GenCtrlOpt3, fmt.Sprintf("[%s] Digits (0-9)", check(m.IncludeDigits))))
		lines = append(lines, renderControl(GenCtrlOpt4, fmt.Sprintf("[%s] Symbols (!@#$)", check(m.IncludeSymbols))))
		lines = append(lines, renderControl(GenCtrlOpt5, fmt.Sprintf("[%s] Exclude Ambiguous (1lI0Oo)", check(m.ExcludeAmbiguous))))
	} else {
		// Passphrase controls
		lines = append(lines, renderControl(GenCtrlLengthWords, fmt.Sprintf("Word Count:  ◀ %2d ▶  (3 - 10 words)", m.WordCount)))
		delim := delimiters[m.DelimiterIdx%len(delimiters)]
		if delim == " " {
			delim = "space"
		}
		lines = append(lines, renderControl(GenCtrlOpt1, fmt.Sprintf("Delimiter:   ◀ %s ▶", delim)))
		lines = append(lines, renderControl(GenCtrlOpt2, fmt.Sprintf("[%s] Capitalize Words", check(m.Capitalize))))
	}

	lines = append(lines, "")

	// Preview Box
	if m.ErrorMsg != "" {
		lines = append(lines, m.Theme.ErrorBadgeStyle.Render(" "+m.ErrorMsg+" "))
	} else {
		previewBox := lipgloss.NewStyle().
			Bold(true).
			Foreground(m.Theme.Text).
			Background(m.Theme.Surface).
			Padding(0, 1).
			Render(m.CurrentGenerated)
		lines = append(lines, "Preview: "+previewBox)

		// Strength & Entropy Gauge
		gauge := m.renderStrengthGauge()
		lines = append(lines, fmt.Sprintf("Entropy: %.1f bits   %s", m.CurrentEntropy, gauge))
	}

	lines = append(lines, "")

	// Buttons
	btnRegen := "[r] Regenerate"
	btnAccept := "[Enter] Use Secret"
	btnCancel := "[Esc] Cancel"

	if m.Control == GenCtrlRegenerate || (m.Mode == GeneratorModePassphrase && m.Control == 4) {
		btnRegen = m.Theme.KeyHintKeyStyle.Render(btnRegen)
	}
	if m.Control == GenCtrlAccept || (m.Mode == GeneratorModePassphrase && m.Control == 5) {
		btnAccept = m.Theme.KeyHintKeyStyle.Render(btnAccept)
	}
	if m.Control == GenCtrlCancel || (m.Mode == GeneratorModePassphrase && m.Control == 6) {
		btnCancel = m.Theme.KeyHintKeyStyle.Render(btnCancel)
	}

	lines = append(lines, fmt.Sprintf("%s    %s    %s", btnRegen, btnAccept, btnCancel))

	content := strings.Join(lines, "\n")
	dialogBox := m.Theme.PanelFocusStyle.
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

func check(b bool) string {
	if b {
		return "x"
	}
	return " "
}

func (m *GeneratorModel) renderStrengthGauge() string {
	switch m.CurrentStrength {
	case service.StrengthVeryWeak, service.StrengthWeak:
		return m.Theme.ErrorBadgeStyle.Render(" " + string(m.CurrentStrength) + " ")
	case service.StrengthFair:
		return m.Theme.WarningBadgeStyle.Render(" " + string(m.CurrentStrength) + " ")
	case service.StrengthStrong, service.StrengthVeryStrong:
		return m.Theme.SuccessBadgeStyle.Render(" " + string(m.CurrentStrength) + " ")
	default:
		return string(m.CurrentStrength)
	}
}
