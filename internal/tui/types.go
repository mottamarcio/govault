package tui

// ScreenState defines the active screen state in the TUI application.
type ScreenState int

const (
	// ScreenInit is the first-run onboarding screen when vault is uninitialized.
	ScreenInit ScreenState = iota
	// ScreenUnlock is the initial master password prompt screen when vault is initialized.
	ScreenUnlock
	// ScreenDashboard is the main vault browsing and management screen.
	ScreenDashboard
	// ScreenEditor is the record creation / edit screen.
	ScreenEditor
	// ScreenGenerator is the interactive password generator modal overlay.
	ScreenGenerator
)

// String returns a human-readable name of the ScreenState.
func (s ScreenState) String() string {
	switch s {
	case ScreenInit:
		return "ScreenInit"
	case ScreenUnlock:
		return "ScreenUnlock"
	case ScreenDashboard:
		return "ScreenDashboard"
	case ScreenEditor:
		return "ScreenEditor"
	case ScreenGenerator:
		return "ScreenGenerator"
	default:
		return "Unknown"
	}
}
