package theme_test

import (
	"strings"
	"testing"

	"github.com/mottamarcio/govault/internal/tui/theme"
)

func TestDefaultTheme(t *testing.T) {
	th := theme.DefaultTheme()
	if th == nil {
		t.Fatal("expected non-nil theme")
	}

	if string(th.Primary) == "" {
		t.Error("expected non-empty Primary color")
	}
	if string(th.Background) == "" {
		t.Error("expected non-empty Background color")
	}
	if string(th.Success) == "" {
		t.Error("expected non-empty Success color")
	}
	if string(th.Danger) == "" {
		t.Error("expected non-empty Danger color")
	}

	renderedTitle := th.TitleStyle.Render("GoVault")
	if !strings.Contains(renderedTitle, "GoVault") {
		t.Errorf("expected rendered title to contain 'GoVault', got %q", renderedTitle)
	}

	renderedBadge := th.StatusBadgeStyle.Render("UNLOCKED")
	if !strings.Contains(renderedBadge, "UNLOCKED") {
		t.Errorf("expected rendered badge to contain 'UNLOCKED', got %q", renderedBadge)
	}
}
