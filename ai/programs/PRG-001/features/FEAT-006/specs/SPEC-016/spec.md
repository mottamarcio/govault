---
id: SPEC-016
type: spec
status: ready
parent: FEAT-006
depends_on:
  - SPEC-008
  - SPEC-010
  - SPEC-013
supersedes: []
---

# SPEC-016: TUI Application Framework, Themes, Unlock Screen & Session Inactivity Lock

## Intent

Specify the core Bubble Tea architecture for GoVault TUI: main application router (Unlock, Dashboard, Form, Modal), Lip Gloss theme styling, interactive masked unlock screen with rate-limit / failure feedback, and inactivity auto-lock countdown timer with clean alternate screen buffer management.

## Requirements

- **R1: Bubble Tea Main Architecture & Navigation Router:** Implement root Bubble Tea `tea.Model` managing screen states (`ScreenUnlock`, `ScreenDashboard`, `ScreenEditor`, `ScreenGenerator`), window resizing events (`tea.WindowSizeMsg`), alternate screen buffer lifecycle (`tea.EnterAltScreen`, `tea.ExitAltScreen`), and graceful cancellation on `Ctrl+C` or exit.
- **R2: Lip Gloss Theme & Layout Primitives:** Define consistent terminal styling system (Nord/Catppuccin inspired dark palette with accessible high-contrast accents) for headers, active borders, status bar badges, keybinding hints, and error alerts.
- **R3: Interactive Unlock Screen Model:**
  - Masked password input field using `bubbles/textinput`.
  - Master password verification delegating to `service.VaultService.Unlock`.
  - Visual error feedback on failed authentication (shaking prompt / alert badge).
  - Clear transition to Dashboard upon successful unlock without retaining plaintext password in model state.
- **R4: Session Inactivity Auto-Lock & Status Bar:**
  - Configurable countdown timer (default 5 minutes of inactivity).
  - Reset timer on any keyboard navigation message (`tea.KeyMsg`).
  - Status bar displaying active vault identifier, record counts, auto-lock countdown indicator, and contextual keybinding hints (`?` for help).
  - On countdown expiry or manual lock keybinding (`Ctrl+L`), wipe memory session and return to `ScreenUnlock`.

## Acceptance Scenarios

- **Scenario 1: TUI Startup and Unlock Transition**
  - *Given* an initialized vault
  - *When* `govault` TUI is launched without session
  - *Then* `ScreenUnlock` is rendered; upon entering correct master password, session is unlocked and `ScreenDashboard` is displayed.
- **Scenario 2: Inactivity Auto-Lock Trigger**
  - *Given* an active unlocked TUI session
  - *When* no key events occur for the configured inactivity duration
  - *Then* the session is automatically locked, decrypted state is wiped, and the view transitions back to `ScreenUnlock`.
- **Scenario 3: Graceful Exit Screen Restoration**
  - *Given* a running TUI session
  - *When* `q` or `Ctrl+C` is pressed
  - *Then* the terminal restores normal screen buffer with cursor visible and exit code 0.

## Edge Cases

- Terminal resized to very small dimensions (< 60x15) displays a clean "Terminal too small" notice instead of breaking layout.
- Incorrect master password retains focus on the input field and displays clear error message.

## Constraints

- Pure presentation: TUI models must interact with the application exclusively through Application Services (`VaultService`, `RecordService`).
- Plaintext master password must never be stored as a persistent struct field in Bubble Tea models.

## Non-Goals

- Web UI or GUI desktop widgets.
