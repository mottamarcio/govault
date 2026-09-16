package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
	"github.com/mottamarcio/govault/internal/tui/screens"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

const (
	minTerminalWidth  = 60
	minTerminalHeight = 15
)

// AppModel is the root Bubble Tea model and state machine router for GoVault TUI.
type AppModel struct {
	Theme         *theme.Theme
	VaultService  *service.VaultService
	RecordService *service.RecordService
	Clipboard     *service.ClipboardService
	DB            *sqlite.DB
	VaultPath     string
	State         ScreenState
	Width         int
	Height        int
	AutoLock      *AutoLock
	Session       *service.Session

	// Screens
	InitScreen      *screens.InitModel
	UnlockScreen    *screens.UnlockModel
	DashboardScreen *screens.DashboardModel
	EditorScreen    *screens.EditorModel
	StatusBarScreen *screens.StatusBarModel

	// Exit flag
	Quitting bool
}

// NewApp creates a new root AppModel.
func NewApp(
	db *sqlite.DB,
	metaRepo *sqlite.MetadataRepository,
	recordRepo *sqlite.RecordRepository,
	tagRepo *sqlite.TagRepository,
	historyRepo *sqlite.HistoryRepository,
	vaultPath string,
	inactivityTimeout time.Duration,
) *AppModel {
	th := theme.DefaultTheme()
	vs := service.NewVaultService(db, metaRepo, vaultPath)
	rs := service.NewRecordService(vs.Session(), recordRepo, tagRepo, historyRepo)
	cs := service.NewClipboardService(nil)

	initScreen := screens.NewInitModel(th, vs, vaultPath)
	unlockScreen := screens.NewUnlockModel(th, vs, vaultPath)
	dashboardScreen := screens.NewDashboardModel(th, rs, cs, 80, 24)
	editorScreen := screens.NewEditorModel(th, rs)
	statusBar := screens.NewStatusBar(th, vaultPath)
	autoLock := NewAutoLock(inactivityTimeout)

	initialState := ScreenUnlock
	initialized, err := vs.IsInitialized(context.Background())
	if err != nil || !initialized {
		initialState = ScreenInit
	}

	return &AppModel{
		Theme:           th,
		VaultService:    vs,
		RecordService:   rs,
		Clipboard:       cs,
		DB:              db,
		VaultPath:       vaultPath,
		State:           initialState,
		Width:           80,
		Height:          24,
		AutoLock:        autoLock,
		Session:         nil,
		InitScreen:      initScreen,
		UnlockScreen:    unlockScreen,
		DashboardScreen: dashboardScreen,
		EditorScreen:    editorScreen,
		StatusBarScreen: statusBar,
		Quitting:        false,
	}
}

// Init initializes the root Bubble Tea application.
func (a *AppModel) Init() tea.Cmd {
	return tea.Batch(
		a.InitScreen.Init(),
		a.UnlockScreen.Init(),
		a.DashboardScreen.Init(),
		a.EditorScreen.Init(),
	)
}

// Update processes Bubble Tea messages and manages screen routing / auto-lock.
func (a *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.Width = msg.Width
		a.Height = msg.Height
		a.InitScreen.SetDimensions(a.Width, a.Height)
		a.UnlockScreen.SetDimensions(a.Width, a.Height)
		a.DashboardScreen.SetDimensions(a.Width, a.Height-1)
		a.EditorScreen.SetDimensions(a.Width, a.Height-1)
		a.StatusBarScreen.SetWidth(a.Width)
		return a, nil

	case tea.KeyMsg:
		// Reset auto-lock countdown on any user key interaction
		a.AutoLock.Reset()

		switch msg.Type {
		case tea.KeyCtrlC:
			a.Quitting = true
			a.LockVault()
			return a, tea.Quit

		case tea.KeyCtrlL:
			// Manual lock keybinding
			a.LockVault()
			return a, nil

		case tea.KeyRunes:
			if msg.String() == "q" && (a.State == ScreenUnlock || a.State == ScreenInit) {
				a.Quitting = true
				return a, tea.Quit
			}
			if msg.String() == "q" && a.State == ScreenDashboard && a.DashboardScreen.Focus != screens.FocusSearch && a.DashboardScreen.Focus != screens.FocusHistory {
				a.Quitting = true
				a.LockVault()
				return a, tea.Quit
			}
		}

	case screens.InitSuccessMsg:
		a.Session = msg.Session
		a.State = ScreenDashboard
		// Reload dashboard records
		_ = a.DashboardScreen.ReloadRecords()

		// Start auto-lock timer
		cmds = append(cmds, a.AutoLock.Start())

		// Update statusbar
		a.StatusBarScreen.SetLockRemaining(a.AutoLock.Remaining(), true)
		a.StatusBarScreen.SetRecordCount(len(a.DashboardScreen.AllRecords))
		a.StatusBarScreen.SetKeyHints([]string{"Tab: Switch", "/: Search", "a: Add", "e: Edit", "d: Delete", "v: Reveal", "c: Copy", "u: User", "Ctrl+L: Lock", "q: Quit", "?: Help"})
		return a, tea.Batch(cmds...)

	case screens.UnlockSuccessMsg:
		a.Session = msg.Session
		a.State = ScreenDashboard
		// Reload dashboard records
		_ = a.DashboardScreen.ReloadRecords()

		// Start auto-lock timer
		cmds = append(cmds, a.AutoLock.Start())

		// Update statusbar
		a.StatusBarScreen.SetLockRemaining(a.AutoLock.Remaining(), true)
		a.StatusBarScreen.SetRecordCount(len(a.DashboardScreen.AllRecords))
		a.StatusBarScreen.SetKeyHints([]string{"Tab: Switch", "/: Search", "a: Add", "e: Edit", "d: Delete", "v: Reveal", "c: Copy", "u: User", "Ctrl+L: Lock", "q: Quit", "?: Help"})
		return a, tea.Batch(cmds...)

	case screens.OpenEditorMsg:
		if msg.Mode == screens.EditorModeCreate {
			a.EditorScreen.OpenCreate(msg.RecordType)
		} else {
			a.EditorScreen.OpenEdit(msg.Record)
		}
		a.State = ScreenEditor
		a.StatusBarScreen.SetKeyHints([]string{"Tab: Switch Field", "Ctrl+G: Generator", "Ctrl+V: Reveal", "Ctrl+N: Add Field", "Ctrl+S: Save", "Esc: Cancel"})
		return a, nil

	case screens.EditorSaveMsg:
		_ = a.DashboardScreen.ReloadRecords()
		a.State = ScreenDashboard
		a.StatusBarScreen.SetRecordCount(len(a.DashboardScreen.AllRecords))
		a.StatusBarScreen.SetKeyHints([]string{"Tab: Switch", "/: Search", "a: Add", "e: Edit", "d: Delete", "v: Reveal", "c: Copy", "u: User", "Ctrl+L: Lock", "q: Quit", "?: Help"})
		return a, func() tea.Msg {
			return screens.ToastMsg{Message: fmt.Sprintf("✓ Record %q saved", msg.Record.Title)}
		}

	case screens.EditorCancelMsg:
		a.State = ScreenDashboard
		a.StatusBarScreen.SetKeyHints([]string{"Tab: Switch", "/: Search", "a: Add", "e: Edit", "d: Delete", "v: Reveal", "c: Copy", "u: User", "Ctrl+L: Lock", "q: Quit", "?: Help"})
		return a, nil

	case AutoLockTickMsg:
		if a.Session != nil && a.Session.IsUnlocked() {
			if a.AutoLock.IsExpired() {
				a.LockVault()
				return a, nil
			}
			a.StatusBarScreen.SetLockRemaining(a.AutoLock.Remaining(), true)
			cmds = append(cmds, a.AutoLock.Tick())
		}
		return a, tea.Batch(cmds...)
	}

	// Route updates based on active screen
	switch a.State {
	case ScreenInit:
		newModel, cmd := a.InitScreen.Update(msg)
		if im, ok := newModel.(*screens.InitModel); ok {
			a.InitScreen = im
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case ScreenUnlock:
		newModel, cmd := a.UnlockScreen.Update(msg)
		if um, ok := newModel.(*screens.UnlockModel); ok {
			a.UnlockScreen = um
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case ScreenDashboard:
		newModel, cmd := a.DashboardScreen.Update(msg)
		if dm, ok := newModel.(*screens.DashboardModel); ok {
			a.DashboardScreen = dm
		}
		if cmd != nil {
			// If cmd returned immediate synchronous messages (e.g. SidebarSelectMsg, OpenEditorMsg), dispatch them
			if innerMsg := cmd(); innerMsg != nil {
				return a.Update(innerMsg)
			}
			cmds = append(cmds, cmd)
		}
		// Sync record count
		a.StatusBarScreen.SetRecordCount(len(a.DashboardScreen.AllRecords))

	case ScreenEditor:
		newModel, cmd := a.EditorScreen.Update(msg)
		if em, ok := newModel.(*screens.EditorModel); ok {
			a.EditorScreen = em
		}
		if cmd != nil {
			if innerMsg := cmd(); innerMsg != nil {
				return a.Update(innerMsg)
			}
			cmds = append(cmds, cmd)
		}
	}

	return a, tea.Batch(cmds...)
}

// LockVault locks the active session, clears in-memory decrypted keys, and returns to UnlockScreen.
func (a *AppModel) LockVault() {
	if a.VaultService != nil {
		_ = a.VaultService.Lock()
	}
	a.Session = nil
	a.AutoLock.Stop()
	a.State = ScreenUnlock
	a.UnlockScreen.ResetInput()
	a.EditorScreen.Reset()
	a.StatusBarScreen.SetLockRemaining(0, false)
	a.StatusBarScreen.SetKeyHints([]string{"q: Quit", "Enter: Unlock"})
}

// View renders the TUI layout.
func (a *AppModel) View() string {
	if a.Quitting {
		return ""
	}

	// Terminal minimum dimension check
	if a.Width < minTerminalWidth || a.Height < minTerminalHeight {
		return a.renderTooSmall()
	}

	var mainView string
	switch a.State {
	case ScreenInit:
		mainView = a.InitScreen.View()
	case ScreenUnlock:
		mainView = a.UnlockScreen.View()
	case ScreenDashboard:
		mainView = a.DashboardScreen.View()
	case ScreenEditor:
		mainView = a.EditorScreen.View()
	default:
		mainView = a.Theme.StatusMutedStyle.Render("Screen under construction")
	}

	statusBarView := a.StatusBarScreen.View()

	return lipgloss.JoinVertical(
		lipgloss.Top,
		mainView,
		statusBarView,
	)
}

func (a *AppModel) renderTooSmall() string {
	msg := fmt.Sprintf(
		"Terminal window too small!\n\nCurrent: %dx%d\nRequired minimum: %dx%d\n\nPlease enlarge your terminal.",
		a.Width, a.Height, minTerminalWidth, minTerminalHeight,
	)
	return lipgloss.Place(
		a.Width,
		a.Height,
		lipgloss.Center,
		lipgloss.Center,
		a.Theme.NoticeBoxStyle.Render(msg),
	)
}
