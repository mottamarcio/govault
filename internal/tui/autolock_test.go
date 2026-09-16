package tui_test

import (
	"strings"
	"testing"
	"time"

	"github.com/mottamarcio/govault/internal/tui"
	"github.com/mottamarcio/govault/internal/tui/screens"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

func TestAutoLock(t *testing.T) {
	al := tui.NewAutoLock(100 * time.Millisecond)
	if al.IsExpired() {
		t.Fatal("expected new inactive autolock to not be expired")
	}

	al.Start()
	if al.IsExpired() {
		t.Fatal("expected freshly started autolock to not be expired")
	}

	time.Sleep(120 * time.Millisecond)
	if !al.IsExpired() {
		t.Fatal("expected autolock to be expired after timeout")
	}

	al.Reset()
	if al.IsExpired() {
		t.Fatal("expected autolock to not be expired after reset")
	}
}

func TestStatusBar(t *testing.T) {
	th := theme.DefaultTheme()
	sb := screens.NewStatusBar(th, "test-vault.db")
	sb.SetWidth(100)
	sb.SetLockRemaining(4*time.Minute+30*time.Second, true)
	sb.SetRecordCount(15)

	view := sb.View()
	if !strings.Contains(view, "GoVault") {
		t.Errorf("expected status bar to contain GoVault, got: %s", view)
	}
	if !strings.Contains(view, "test-vault.db") {
		t.Errorf("expected status bar to contain test-vault.db, got: %s", view)
	}
	if !strings.Contains(view, "Auto-lock: 4m 30s") {
		t.Errorf("expected status bar to contain countdown timer, got: %s", view)
	}
	if !strings.Contains(view, "15 records") {
		t.Errorf("expected status bar to contain record count, got: %s", view)
	}
}
