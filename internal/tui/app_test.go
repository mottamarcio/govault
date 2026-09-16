package tui_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
	"github.com/mottamarcio/govault/internal/tui"
	"github.com/mottamarcio/govault/internal/tui/screens"
)

func setupTestApp(t *testing.T) (*tui.AppModel, *sqlite.DB, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vault.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		db.Close()
		t.Fatalf("failed to migrate db: %v", err)
	}

	metaRepo := sqlite.NewMetadataRepository(db)
	tagRepo := sqlite.NewTagRepository(db)
	recordRepo := sqlite.NewRecordRepository(db, tagRepo)
	historyRepo := sqlite.NewHistoryRepository(db)

	vs := service.NewVaultService(db, metaRepo, dbPath)
	if err := vs.Init(ctx, "SecretPassword123!"); err != nil {
		db.Close()
		t.Fatalf("failed to init vault: %v", err)
	}

	app := tui.NewApp(db, metaRepo, recordRepo, tagRepo, historyRepo, dbPath, 1*time.Second)

	cleanup := func() {
		_ = db.Close()
	}

	return app, db, cleanup
}

func TestAppRouter_WindowResizeAndTooSmall(t *testing.T) {
	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	// Normal size
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	view := app.View()
	if strings.Contains(view, "Terminal window too small") {
		t.Errorf("expected normal view, got small window warning: %s", view)
	}

	// Too small size
	app.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	viewSmall := app.View()
	if !strings.Contains(viewSmall, "Terminal window too small") {
		t.Errorf("expected small window warning, got: %s", viewSmall)
	}
}

func TestAppRouter_UnlockFlowAndManualLock(t *testing.T) {
	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	if app.State != tui.ScreenUnlock {
		t.Fatalf("expected initial state ScreenUnlock, got %v", app.State)
	}

	// Simulate unlock message
	ctx := context.Background()
	sess, err := app.VaultService.Unlock(ctx, "SecretPassword123!")
	if err != nil {
		t.Fatalf("failed to unlock: %v", err)
	}

	app.Update(screens.UnlockSuccessMsg{Session: sess})

	if app.State != tui.ScreenDashboard {
		t.Fatalf("expected state ScreenDashboard after unlock, got %v", app.State)
	}
	if !app.Session.IsUnlocked() {
		t.Fatal("expected session to be unlocked")
	}

	dashView := app.View()
	if !strings.Contains(dashView, "CATEGORIES") {
		t.Errorf("expected dashboard view to contain CATEGORIES, got: %s", dashView)
	}

	// Manual lock with Ctrl+L
	app.Update(tea.KeyMsg{Type: tea.KeyCtrlL})

	if app.State != tui.ScreenUnlock {
		t.Fatalf("expected state ScreenUnlock after Ctrl+L, got %v", app.State)
	}
	if app.Session != nil {
		t.Fatal("expected session to be nil after lock")
	}
}

func TestAppRouter_AutoLockExpiration(t *testing.T) {
	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	ctx := context.Background()
	sess, err := app.VaultService.Unlock(ctx, "SecretPassword123!")
	if err != nil {
		t.Fatalf("failed to unlock: %v", err)
	}

	app.Update(screens.UnlockSuccessMsg{Session: sess})
	if app.State != tui.ScreenDashboard {
		t.Fatalf("expected ScreenDashboard, got %v", app.State)
	}

	// Wait for timeout (1s) to pass
	time.Sleep(1100 * time.Millisecond)

	// Send tick msg
	app.Update(tui.AutoLockTickMsg(time.Now()))

	if app.State != tui.ScreenUnlock {
		t.Fatalf("expected state ScreenUnlock after auto-lock timeout, got %v", app.State)
	}
	if app.Session != nil {
		t.Fatal("expected nil session after auto-lock")
	}
}

func TestAppModel_AddSecretWithGeneratorAndEdit(t *testing.T) {
	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	ctx := context.Background()
	sess, err := app.VaultService.Unlock(ctx, "SecretPassword123!")
	if err != nil {
		t.Fatalf("failed to unlock: %v", err)
	}
	app.Update(screens.UnlockSuccessMsg{Session: sess})

	// 1. Press 'a' to open form editor
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if app.State != tui.ScreenEditor {
		t.Fatalf("expected ScreenEditor after pressing 'a', got %v", app.State)
	}

	// 2. Set Title, Username, and URI
	app.EditorScreen.TitleInput.SetValue("Production DB")
	app.EditorScreen.UsernameInput.SetValue("postgres")
	app.EditorScreen.URIInput.SetValue("postgres://prod.internal:5432")

	// 3. Press Ctrl+G to open Generator overlay
	app.Update(tea.KeyMsg{Type: tea.KeyCtrlG})
	if !app.EditorScreen.Generator.Active {
		t.Fatalf("expected generator modal to be active")
	}

	// 4. Press Enter in Generator modal to accept generated password
	app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.EditorScreen.Generator.Active {
		t.Fatalf("expected generator modal to be closed after accept")
	}
	generatedPass := app.EditorScreen.SecretInput.Value()
	if len(generatedPass) < 8 {
		t.Fatalf("expected generated password in secret input, got %q", generatedPass)
	}

	// 5. Save record with Ctrl+S
	app.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if app.State != tui.ScreenDashboard {
		t.Fatalf("expected return to ScreenDashboard after saving, got %v", app.State)
	}
	if len(app.DashboardScreen.AllRecords) != 1 {
		t.Fatalf("expected 1 record in dashboard, got %d", len(app.DashboardScreen.AllRecords))
	}
	savedRec := app.DashboardScreen.AllRecords[0]
	if savedRec.Version != 1 {
		t.Fatalf("unexpected saved record version: %+v", savedRec)
	}

	// 6. Press 'e' to edit the selected record
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if app.State != tui.ScreenEditor {
		t.Fatalf("expected ScreenEditor after pressing 'e', got %v", app.State)
	}

	// 7. Update URI and save
	app.EditorScreen.URIInput.SetValue("postgres://prod-v2.internal:5432")
	app.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if app.State != tui.ScreenDashboard {
		t.Fatalf("expected return to ScreenDashboard after edit save")
	}
	if app.DashboardScreen.AllRecords[0].Version != 2 {
		t.Fatalf("expected version 2 after update, got %d", app.DashboardScreen.AllRecords[0].Version)
	}
}

func TestAppModel_DeleteWithConfirmationModal(t *testing.T) {
	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	ctx := context.Background()
	sess, err := app.VaultService.Unlock(ctx, "SecretPassword123!")
	if err != nil {
		t.Fatalf("failed to unlock: %v", err)
	}
	app.Update(screens.UnlockSuccessMsg{Session: sess})

	// Add record first
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	app.EditorScreen.TitleInput.SetValue("Staging API")
	app.EditorScreen.UsernameInput.SetValue("staging-admin")
	app.EditorScreen.URIInput.SetValue("https://staging.api.com")
	app.Update(tea.KeyMsg{Type: tea.KeyCtrlS})

	if len(app.DashboardScreen.AllRecords) != 1 {
		t.Fatalf("expected 1 record before delete test")
	}

	// Press 'd' -> triggers confirmation modal
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !app.DashboardScreen.ConfirmModal.Active {
		t.Fatalf("expected ConfirmModal to be active")
	}

	// Cancel with 'n'
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if app.DashboardScreen.ConfirmModal.Active {
		t.Fatalf("expected ConfirmModal to be inactive after 'n'")
	}
	if len(app.DashboardScreen.FilteredRecords) != 1 {
		t.Fatalf("expected record still present after cancel")
	}

	// Press 'd' again and confirm with 'y'
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})

	// Verify record is removed from active filtered records
	if len(app.DashboardScreen.FilteredRecords) != 0 {
		t.Fatalf("expected 0 active records after trash delete, got %d", len(app.DashboardScreen.FilteredRecords))
	}
}

func TestAppModel_HelpModalToggle(t *testing.T) {
	app, _, cleanup := setupTestApp(t)
	defer cleanup()

	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	ctx := context.Background()
	sess, err := app.VaultService.Unlock(ctx, "SecretPassword123!")
	if err != nil {
		t.Fatalf("failed to unlock: %v", err)
	}
	app.Update(screens.UnlockSuccessMsg{Session: sess})

	// Press '?' to toggle help overlay
	app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !app.DashboardScreen.HelpModal.Active {
		t.Fatalf("expected help modal active after '?'")
	}

	view := app.View()
	if !strings.Contains(view, "KEYBOARD SHORTCUTS") {
		t.Fatalf("expected view to render shortcuts overlay, got: %s", view)
	}

	// Press Esc to dismiss
	app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if app.DashboardScreen.HelpModal.Active {
		t.Fatalf("expected help modal dismissed after Esc")
	}
}

func TestAppModel_FirstRunInitializationAndRouting(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "uninitialized_vault.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	metaRepo := sqlite.NewMetadataRepository(db)
	tagRepo := sqlite.NewTagRepository(db)
	recordRepo := sqlite.NewRecordRepository(db, tagRepo)
	historyRepo := sqlite.NewHistoryRepository(db)

	app := tui.NewApp(db, metaRepo, recordRepo, tagRepo, historyRepo, dbPath, 1*time.Minute)
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// 1. Verify initial state is ScreenInit
	if app.State != tui.ScreenInit {
		t.Fatalf("expected initial state ScreenInit on fresh db, got %v", app.State)
	}

	view := app.View()
	if !strings.Contains(view, "FIRST-TIME SETUP") {
		t.Fatalf("expected view to render FIRST-TIME SETUP, got: %s", view)
	}

	// 2. Perform initialization via wizard
	for _, ch := range "MasterKey12345!" {
		app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}
	app.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, ch := range "MasterKey12345!" {
		app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{ch}})
	}

	// Press Enter to trigger initialization
	_, cmd := app.InitScreen.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected async cmd on Enter in ScreenInit")
	}

	msg := cmd()
	initSuccessMsg, ok := msg.(screens.InitSuccessMsg)
	if !ok {
		t.Fatalf("expected InitSuccessMsg, got %T: %+v", msg, msg)
	}

	// Dispatch success to app
	app.Update(initSuccessMsg)

	// 3. Verify state transitioned directly to ScreenDashboard
	if app.State != tui.ScreenDashboard {
		t.Fatalf("expected State to be ScreenDashboard after InitSuccessMsg, got %v", app.State)
	}
	if app.Session == nil || !app.Session.IsUnlocked() {
		t.Fatal("expected authenticated active session after first-run init")
	}
}

func TestAppModel_ScreenInit_Quit(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "quit_vault.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate db: %v", err)
	}

	metaRepo := sqlite.NewMetadataRepository(db)
	tagRepo := sqlite.NewTagRepository(db)
	recordRepo := sqlite.NewRecordRepository(db, tagRepo)
	historyRepo := sqlite.NewHistoryRepository(db)

	app := tui.NewApp(db, metaRepo, recordRepo, tagRepo, historyRepo, dbPath, 1*time.Minute)
	app.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	if app.State != tui.ScreenInit {
		t.Fatalf("expected ScreenInit, got %v", app.State)
	}

	// Press 'q' to quit
	_, cmd := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if !app.Quitting {
		t.Fatal("expected app to be quitting after 'q'")
	}
	if cmd == nil {
		t.Fatal("expected tea.Quit cmd after 'q'")
	}
}

