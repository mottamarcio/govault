package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// AutoLockTickMsg represents a periodic tick to update the inactivity timer.
type AutoLockTickMsg time.Time

// AutoLockExpiredMsg signals that inactivity duration has elapsed and vault must lock.
type AutoLockExpiredMsg struct{}

// AutoLock tracks session activity and triggers lock events on inactivity timeout.
type AutoLock struct {
	timeout      time.Duration
	lastActivity time.Time
	active       bool
}

// NewAutoLock creates a new AutoLock manager with the given inactivity timeout.
func NewAutoLock(timeout time.Duration) *AutoLock {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	return &AutoLock{
		timeout:      timeout,
		lastActivity: time.Now(),
		active:       false,
	}
}

// Start activates inactivity tracking.
func (a *AutoLock) Start() tea.Cmd {
	a.active = true
	a.lastActivity = time.Now()
	return a.Tick()
}

// Stop deactivates inactivity tracking.
func (a *AutoLock) Stop() {
	a.active = false
}

// Reset updates the last activity timestamp to now.
func (a *AutoLock) Reset() {
	if a.active {
		a.lastActivity = time.Now()
	}
}

// Remaining returns the duration until auto-lock fires.
func (a *AutoLock) Remaining() time.Duration {
	if !a.active {
		return a.timeout
	}
	elapsed := time.Since(a.lastActivity)
	rem := a.timeout - elapsed
	if rem < 0 {
		return 0
	}
	return rem
}

// IsExpired checks if the inactivity timeout has elapsed.
func (a *AutoLock) IsExpired() bool {
	if !a.active {
		return false
	}
	return time.Since(a.lastActivity) >= a.timeout
}

// Tick returns a tea.Cmd for a 1-second interval tick.
func (a *AutoLock) Tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return AutoLockTickMsg(t)
	})
}
