---
type: tasks
for: SPEC-016
---

# Tasks: SPEC-016 (TUI Application Framework, Themes, Unlock Screen & Session Inactivity Lock)

## TASK-001 — Add Bubble Tea and Lip Gloss Dependencies & Theme Styling Primitives
- [x] Requirements: `SPEC-016:R2`
- Dependencies: none
- Scope: `go.mod`, `go.sum`, `internal/tui/theme/theme.go`
- Verification: `go test -v ./internal/tui/theme/...`
- Evidence: `go test -v ./internal/tui/theme/...` passed (`TestDefaultTheme` 0.003s). Charm dependencies added to `go.mod`.
- Details: Add Charm dependencies (`bubbletea`, `lipgloss`, `bubbles`). Implement `internal/tui/theme/theme.go` with Nord/Catppuccin inspired dark terminal styling palette (Primary, Secondary, Accent, Border, Success, Warning, Danger) and reusable Lip Gloss box, border, text, header, and badge styles.

## TASK-002 — Inactivity Auto-Lock Timer & Status Bar Model
- [x] Requirements: `SPEC-016:R4`
- Dependencies: TASK-001
- Scope: `internal/tui/autolock.go`, `internal/tui/screens/statusbar.go`
- Verification: `go test -v ./internal/tui/... -run TestAutoLock`
- Evidence: `go test -v ./internal/tui/...` passed (`TestAutoLock` 0.12s, `TestStatusBar` 0.00s). Inactivity countdown tracking, tick messages, reset semantics, and status bar rendering verified.
- Details: Implement `internal/tui/autolock.go` managing inactivity countdown ("default 5 minutes of inactivity") with periodic `tea.Tick` messages and reset trigger on any keyboard input (`tea.KeyMsg`). Implement `internal/tui/screens/statusbar.go` rendering active vault identifier, record counts, auto-lock countdown timer indicator, and contextual keybinding hints (`?` for help).

## TASK-003 — Interactive Masked Unlock Screen Model
- [x] Requirements: `SPEC-016:R3`
- Dependencies: TASK-001
- Scope: `internal/tui/screens/unlock.go`
- Verification: `go test -v ./internal/tui/screens/... -run TestUnlock`
- Evidence: `go test -v ./internal/tui/screens/... -run TestUnlock` passed (`TestUnlockModel_Success` 0.35s, `TestUnlockModel_Failure` 0.26s). Masked password input, async authentication, error handling, and zeroization verified.
- Details: Implement `internal/tui/screens/unlock.go` with masked master password input field using `bubbles/textinput` (`EchoMode = EchoPassword`), master password verification delegating to `service.VaultService.Unlock`, visual error feedback on failed authentication (shaking prompt / alert badge), and clear transition upon successful unlock without retaining plaintext master password in model state.

## TASK-004 — Bubble Tea Root Application Router & Terminal Lifecycle
- [x] Requirements: `SPEC-016:R1`, `SPEC-016:R4`
- Dependencies: TASK-002, TASK-003
- Scope: `internal/tui/app.go`, `internal/tui/types.go`
- Verification: `go test -v ./internal/tui/... -run TestAppRouter`
- Evidence: `go test -v ./internal/tui/... -run TestAppRouter` passed (`TestAppRouter_WindowResizeAndTooSmall` 0.18s, `TestAppRouter_UnlockFlowAndManualLock` 0.29s, `TestAppRouter_AutoLockExpiration` 1.31s). Screen router, resize handling, minimum size notice (< 60x15), manual Ctrl+L lock, and auto-lock zeroization verified.
- Details: Implement root `AppModel` managing screen states (`ScreenUnlock`, `ScreenDashboard`, `ScreenEditor`, `ScreenGenerator`), window resizing events (`tea.WindowSizeMsg`), minimum dimension fallback ("< 60x15" renders clean "Terminal too small" warning), alternate screen buffer lifecycle (`tea.EnterAltScreen`, `tea.ExitAltScreen`), manual lock hotkey (`Ctrl+L`), auto-lock expiry zeroization transition back to `ScreenUnlock`, and graceful cancellation on `Ctrl+C` or `q`.

## TASK-005 — CLI Integration & End-to-End TUI Test Suite
- [x] Requirements: `SPEC-016:R1`, `SPEC-016:R2`, `SPEC-016:R3`, `SPEC-016:R4`
- Dependencies: TASK-004
- Scope: `internal/cli/tui.go`, `internal/cli/root.go`, `internal/tui/tui_test.go`
- Verification: `go test -v -race ./internal/tui/... ./internal/cli/...`
- Evidence: `go test -v -race ./internal/tui/... ./internal/cli/...` passed cleanly across all test suites without race conditions. `govault tui` CLI launcher command successfully attached to root CLI with configurable inactivity timeout flag.
- Details: Create Cobra command `govault tui` in `internal/cli/tui.go`, register it in `root.go`, and add comprehensive unit and mock event tests covering startup, unlock transition, auto-lock countdown expiration, zeroization, and clean exit.
