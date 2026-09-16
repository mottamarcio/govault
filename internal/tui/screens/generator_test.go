package screens

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

func TestGeneratorModel_PasswordMode(t *testing.T) {
	th := theme.DefaultTheme()
	m := NewGeneratorModel(th)
	m.SetDimensions(80, 24)

	m.Open()
	if !m.Active {
		t.Fatalf("expected generator to be active")
	}
	if len(m.CurrentGenerated) != 20 {
		t.Fatalf("expected 20-char default generated password, got %d (%s)", len(m.CurrentGenerated), m.CurrentGenerated)
	}
	if m.CurrentEntropy <= 0 {
		t.Fatalf("expected positive entropy, got %.2f", m.CurrentEntropy)
	}

	// Change length
	m.Control = GenCtrlLengthWords
	m.Update(tea.KeyMsg{Type: tea.KeyRight}) // 21
	if len(m.CurrentGenerated) != 21 {
		t.Fatalf("expected 21-char password after length increment, got %d", len(m.CurrentGenerated))
	}

	// Trigger manual regenerate with 'r'
	oldSec := m.CurrentGenerated
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if m.CurrentGenerated == "" {
		t.Fatalf("expected non-empty secret after regen")
	}
	// Note: randomly could theoretically be identical, but with 21 chars it's virtually impossible
	_ = oldSec

	// Accept generated secret
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatalf("expected cmd on Enter")
	}
	resMsg, ok := cmd().(GeneratorResultMsg)
	if !ok || resMsg.Secret == "" {
		t.Fatalf("expected non-empty GeneratorResultMsg, got %+v", resMsg)
	}
	if m.Active {
		t.Fatalf("expected modal inactive after accept")
	}
}

func TestGeneratorModel_PassphraseMode(t *testing.T) {
	th := theme.DefaultTheme()
	m := NewGeneratorModel(th)
	m.SetDimensions(80, 24)

	m.Open()
	// Switch to Passphrase mode
	m.Control = GenCtrlMode
	m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if m.Mode != GeneratorModePassphrase {
		t.Fatalf("expected Passphrase mode")
	}

	if m.CurrentStrength != service.StrengthStrong && m.CurrentStrength != service.StrengthVeryStrong && m.CurrentStrength != service.StrengthFair {
		t.Fatalf("expected reasonable passphrase strength, got %s", m.CurrentStrength)
	}

	// Toggle delimiter
	m.Control = GenCtrlOpt1
	m.Update(tea.KeyMsg{Type: tea.KeyRight}) // from '-' to ' '
	if m.CurrentGenerated == "" {
		t.Fatalf("expected generated passphrase")
	}

	// View rendering
	v := m.View()
	if v == "" {
		t.Fatalf("expected non-empty view")
	}

	// Cancel with Esc
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatalf("expected cancel cmd")
	}
	if _, ok := cmd().(GeneratorCloseMsg); !ok {
		t.Fatalf("expected GeneratorCloseMsg")
	}
}
