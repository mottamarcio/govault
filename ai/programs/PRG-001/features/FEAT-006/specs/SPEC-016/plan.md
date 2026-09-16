# Implementation Plan: SPEC-016 (TUI Application Framework, Themes, Unlock Screen & Session Inactivity Lock)

## Summary

This plan defines the technical architecture, component design, and implementation sequence for the GoVault Terminal User Interface (TUI) core framework using Charm's Bubble Tea (`tea.Model`), Lip Gloss styling system, and Bubbles component library. It establishes the central state machine router (`ScreenUnlock`, `ScreenDashboard`, `ScreenEditor`, `ScreenGenerator`), a dark terminal theme, an interactive masked unlock screen with rate-limit and error feedback, and a background inactivity auto-lock countdown timer with zero-leak session cleanup and graceful terminal restoration.

## Repository Context

- `internal/service/vault_service.go` & `session.go`: Manages vault authentication, session unlocking, and key zeroization upon lock.
- `internal/service/record_service.go`: Provides high-level secret retrieval and management.
- `internal/storage/sqlite`: Underlying database storage engine and repositories.
- `internal/cli/root.go`: Root CLI entrypoint where the TUI launcher command will be integrated.

## Requirement Coverage

- **R1 -> Bubble Tea Main Architecture & Navigation Router:**
  - Implemented in `internal/tui/app.go` (`AppModel`) implementing `tea.Model`.
  - Maintains active screen state enum (`ScreenUnlock`, `ScreenDashboard`, `ScreenEditor`, `ScreenGenerator`).
  - Handles `tea.WindowSizeMsg` to dynamically adjust child model dimensions.
  - Controls alternate screen buffer transitions (`tea.EnterAltScreen`, `tea.ExitAltScreen`) and cursor restoration.
  - Dispatches global navigation commands and handles graceful exit on `Ctrl+C` or `q` (when on unlock screen).
- **R2 -> Lip Gloss Theme & Layout Primitives:**
  - Implemented in `internal/tui/theme/theme.go`.
  - Defines Nord/Catppuccin inspired dark palette: Primary (Accent Blue/Cyan), Secondary (Muted Gray), Border (Subtle Blue/Gray), Success (Emerald Green), Warning (Amber), Danger (Crimson).
  - Provides pre-styled Lip Gloss components: header title banner, active/inactive panel borders, status bar badges, input boxes, and toast notifications.
- **R3 -> Interactive Unlock Screen Model:**
  - Implemented in `internal/tui/screens/unlock.go` (`UnlockModel`).
  - Embeds `bubbles/textinput` configured with `EchoMode = EchoPassword` (masked asterisks/bullets).
  - On Enter submission, asynchronously triggers `service.VaultService.Unlock`.
  - Displays animated/shaking visual error badge on authentication failure.
  - On success, transmits unlocked `*service.Session` to `AppModel` and immediately clears/zeroizes the input buffer.
- **R4 -> Session Inactivity Auto-Lock & Status Bar:**
  - Implemented in `internal/tui/screens/statusbar.go` and `internal/tui/autolock.go`.
  - Tracks last activity timestamp; every tick message (`tea.Tick`) updates the countdown display (e.g. `Auto-lock: 4m 52s`).
  - Any `tea.KeyMsg` resets the inactivity timer to the configured duration (default 5 minutes).
  - Manual lock hotkey (`Ctrl+L`) or timer expiration immediately calls `Session.Lock()`, wipes memory, and returns the app router to `ScreenUnlock`.

## Architecture

```
                                [Bubble Tea Event Loop]
                                          │
                                          ▼
                                   [AppModel (Root)]
                   ├── Router: ScreenUnlock | ScreenDashboard | ScreenEditor
                   ├── Theme: Lip Gloss Styles & Layout Primitives
                   ├── AutoLock: Inactivity Countdown Manager
                   └── Window: Dynamic (Width, Height) Propagator
                                          │
                  ┌───────────────────────┴───────────────────────┐
                  ▼                                               ▼
         [Screen: UnlockModel]                         [Screen: DashboardModel]
    - bubbles/textinput (EchoPassword)             - Left: Category/Tag Sidebar
    - Async auth command                           - Right: Records Table & Detail
    - Error feedback badge                         - Bottom: StatusBar & Hints
                  │                                               │
                  └───────────────────────┬───────────────────────┘
                                          ▼
                              [Application Services]
                       - service.VaultService (Unlock/Lock)
                       - service.RecordService
```

## Components Affected

- `go.mod` / `go.sum`: Add dependencies for `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/charmbracelet/bubbles`.
- `internal/tui/theme/theme.go`: Colors, styles, badges, and layout helpers.
- `internal/tui/screens/unlock.go`: Masked master password prompt model.
- `internal/tui/screens/statusbar.go`: Footer status bar with lock countdown and help hints.
- `internal/tui/autolock.go`: Inactivity timer messages and ticks.
- `internal/tui/app.go`: Root model, screen router, window resize handler, and lifecycle manager.
- `internal/cli/tui.go`: `govault tui` (or default `govault` interactive mode) command wrapper.

## Data Changes

None. Interacts entirely in-memory with existing domain and service layers.

## API Changes

- New package `internal/tui` exposing:
  - `NewApp(vaultPath string, inactivityTimeout time.Duration) *AppModel`
  - `Run(vaultPath string) error`
- Cobra command: `govault tui` (and launching TUI when `govault` is invoked without subcommands in an interactive TTY).

## Integration Changes

- Terminal alternate screen buffer handling ensures terminal state is 100% restored upon exit or signal interruption.

## Implementation Sequence

1. **Add Bubble Tea Dependencies:** Run `go get github.com/charmbracelet/bubbletea@v0.25.0 github.com/charmbracelet/lipgloss@v0.9.1 github.com/charmbracelet/bubbles@v0.18.0`.
2. **Theme & Layout System:** Implement `internal/tui/theme/theme.go` with palette and reusable Lip Gloss box/text styles.
3. **Inactivity Lock Engine:** Implement `internal/tui/autolock.go` with tick messages and reset triggers.
4. **Unlock Screen Model:** Implement `internal/tui/screens/unlock.go` with password input masking and async unlock command.
5. **Status Bar Model:** Implement `internal/tui/screens/statusbar.go` rendering active vault ID, lock countdown, and hints.
6. **Root App Router:** Implement `internal/tui/app.go` coordinating screens, window sizing, and transition events.
7. **CLI Integration & Tests:** Create `internal/cli/tui.go`, register command, and write unit/mock model tests in `internal/tui/tui_test.go`.

## Test Strategy

- **Model Unit Tests:** Use `tea.Msg` injection (e.g. `app.Update(tea.KeyMsg{...})`) to test screen transitions, password submission, error badge triggers, and window resizing without requiring a physical TTY.
- **Inactivity Timer Tests:** Simulate tick messages and verify transition from unlocked dashboard to locked screen.
- **Race Condition Verification:** Execute `go test -v -race ./internal/tui/...`.

## Risks

- *Risk:* Terminal corruption if TUI crashes unexpectedly.
  *Mitigation:* Bubble Tea automatically catches panics and restores standard terminal mode; defer cleanup ensures `tea.ExitAltScreen` is sent.
- *Risk:* Memory leaks with long-lived decrypted session.
  *Mitigation:* Inactivity lock aggressively zeroes session keys and purges decrypted cache on lock or exit.

## Assumptions

- Terminal supports standard ANSI/VT100 escape codes and 256 colors.
