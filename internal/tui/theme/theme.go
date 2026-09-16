package theme

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme encapsulates the terminal visual styling system for GoVault TUI.
type Theme struct {
	// Palette colors
	Primary     lipgloss.Color
	Secondary   lipgloss.Color
	Accent      lipgloss.Color
	Background  lipgloss.Color
	Surface     lipgloss.Color
	Border      lipgloss.Color
	BorderFocus lipgloss.Color
	Text        lipgloss.Color
	TextMuted   lipgloss.Color
	Success     lipgloss.Color
	Warning     lipgloss.Color
	Danger      lipgloss.Color

	// Typography & Component Styles
	TitleStyle        lipgloss.Style
	SubtitleStyle     lipgloss.Style
	HeaderBannerStyle lipgloss.Style
	PanelStyle        lipgloss.Style
	PanelFocusStyle   lipgloss.Style
	StatusBarStyle    lipgloss.Style
	StatusBadgeStyle  lipgloss.Style
	StatusMutedStyle  lipgloss.Style
	StatusTimerStyle  lipgloss.Style
	ErrorBadgeStyle   lipgloss.Style
	SuccessBadgeStyle lipgloss.Style
	WarningBadgeStyle lipgloss.Style
	KeyHintKeyStyle   lipgloss.Style
	KeyHintDescStyle  lipgloss.Style
	InputPromptStyle  lipgloss.Style
	InputBoxStyle     lipgloss.Style
	NoticeBoxStyle    lipgloss.Style
}

// DefaultTheme returns the default Nord/Catppuccin inspired dark theme.
func DefaultTheme() *Theme {
	t := &Theme{
		Primary:     lipgloss.Color("#88C0D0"), // Frost Blue
		Secondary:   lipgloss.Color("#81A1C1"), // Soft Blue
		Accent:      lipgloss.Color("#8FBCBB"), // Frost Teal
		Background:  lipgloss.Color("#2E3440"), // Polar Night Dark
		Surface:     lipgloss.Color("#3B4252"), // Polar Night Lighter
		Border:      lipgloss.Color("#4C566A"), // Muted Border
		BorderFocus: lipgloss.Color("#88C0D0"), // Active Border Frost Blue
		Text:        lipgloss.Color("#ECEFF4"), // Snow Storm White
		TextMuted:   lipgloss.Color("#D8DEE9"), // Snow Storm Light Gray
		Success:     lipgloss.Color("#A3BE8C"), // Aurora Green
		Warning:     lipgloss.Color("#EBCB8B"), // Aurora Yellow
		Danger:      lipgloss.Color("#BF616A"), // Aurora Red
	}

	t.TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary)

	t.SubtitleStyle = lipgloss.NewStyle().
		Foreground(t.TextMuted)

	t.HeaderBannerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Text).
		Background(t.Surface).
		Padding(0, 1)

	t.PanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(0, 1)

	t.PanelFocusStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocus).
		Padding(0, 1)

	t.StatusBarStyle = lipgloss.NewStyle().
		Background(t.Surface).
		Foreground(t.Text).
		Padding(0, 1)

	t.StatusBadgeStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Background).
		Background(t.Primary).
		Padding(0, 1)

	t.StatusMutedStyle = lipgloss.NewStyle().
		Foreground(t.TextMuted)

	t.StatusTimerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Warning)

	t.ErrorBadgeStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(t.Danger).
		Padding(0, 1)

	t.SuccessBadgeStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Background).
		Background(t.Success).
		Padding(0, 1)

	t.WarningBadgeStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Background).
		Background(t.Warning).
		Padding(0, 1)

	t.KeyHintKeyStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary)

	t.KeyHintDescStyle = lipgloss.NewStyle().
		Foreground(t.TextMuted)

	t.InputPromptStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary)

	t.InputBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocus).
		Padding(1, 2)

	t.NoticeBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Warning).
		Padding(1, 2).
		Align(lipgloss.Center)

	return t
}
