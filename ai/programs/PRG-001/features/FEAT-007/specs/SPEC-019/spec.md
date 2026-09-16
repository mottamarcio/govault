---
id: SPEC-019
type: spec
status: ready
parent: FEAT-007
depends_on:
  - SPEC-001
  - SPEC-002
  - SPEC-007
  - SPEC-016
supersedes: []
---

# SPEC-019: TUI First-Run Setup & Interactive Vault Initialization Wizard

## Intent

Specify the automatic detection of uninitialized or missing vault database files upon TUI startup, providing an interactive first-run onboarding screen (`ScreenInit`) that guides users to create and confirm their master password, enforces password strength & matching validation, initializes the database with Argon2id encryption, and transitions directly into the unlocked dashboard without requiring prior CLI `govault init`.

## Requirements

### R1 — Uninitialized Vault Detection & Initial Routing
- On TUI startup, `AppModel` inspects the target `VaultPath` via `VaultService` to determine if a valid initialized vault exists.
- If the vault is already initialized: `AppModel` routes initial state to `ScreenUnlock`.
- If the vault is not initialized (file does not exist or metadata table is missing/empty): `AppModel` routes initial state to `ScreenInit`.

### R2 — First-Run Setup Screen & Confirmation Inputs (`ScreenInit`)
- `ScreenInit` renders a dedicated onboarding card displaying:
  - Header: " Welcome to GoVault — First-Time Setup "
  - Security notice: "This master password derives the master key that encrypts your entire vault. It is never stored and cannot be recovered if lost."
  - Master Password input field (`bubbles/textinput`, masked with `•`).
  - Confirm Password input field (`bubbles/textinput`, masked with `•`).
  - Real-time password strength meter (entropy bits and qualitative level: Weak / Fair / Strong).
- Field navigation using `Tab`, `Shift+Tab`, `↑`, and `↓`.
- Input validation:
  - Master password must be at least 8 characters long.
  - Confirm password must match master password exactly.
  - If validation fails, display a visible error badge ("Password must be at least 8 characters" or "Passwords do not match").

### R3 — Cryptographic Vault Initialization & Auto-Unlock
- Pressing `Enter` on a valid form submits the initialization request:
  - Invokes `VaultService.Init(ctx, masterPassword)` to run SQLite migrations, derive the root encryption key via Argon2id, and persist the metadata envelope.
  - Automatically unlocks the session (`VaultService.Unlock`) with the newly established credentials.
  - Zeroizes all password buffers in memory immediately.
  - Emits `InitSuccessMsg{Session: session}` transitioning `AppModel.State` directly to `ScreenDashboard`, starting the auto-lock countdown, and updating the status bar.

### R4 — Safe Cancellation & Clean Exit
- Pressing `q` or `Ctrl+C` on `ScreenInit` aborts the wizard and quits the application without leaving incomplete database states.

## Acceptance Scenarios

- **Scenario 1: First-Time User Setup**
  - *Given* no vault database exists at `~/.govault/vault.db`
  - *When* the user runs `govault tui`
  - *Then* `ScreenInit` is displayed with fields for Password and Confirm Password
  - *When* the user inputs "MasterPassword123!", confirms "MasterPassword123!", and presses Enter
  - *Then* the vault is initialized, authenticated session is created, and the user enters `ScreenDashboard` with 0 records.

- **Scenario 2: Password Mismatch Validation**
  - *Given* `ScreenInit` is displayed
  - *When* the user inputs "SecretPassword123!" in password and "DifferentPassword123!" in confirm password
  - *Then* an inline validation error "Passwords do not match" is displayed, and initialization is prevented.

- **Scenario 3: Short Password Validation**
  - *Given* `ScreenInit` is displayed
  - *When* the user inputs "short" (<8 characters)
  - *Then* an inline validation error "Password must be at least 8 characters" is displayed.

- **Scenario 4: Subsequent Launch Routes to Unlock**
  - *Given* an already initialized vault
  - *When* the user runs `govault tui`
  - *Then* `ScreenUnlock` is displayed directly.

## Edge Cases

- Non-existent parent directory in `VaultPath` (e.g. `~/.govault`) is automatically created with strict permissions (`0700`).
- Window resized during initialization wizard dynamically centers the setup card.

## Constraints

- Zero network calls.
- Immediate zeroization of master password strings and byte slices in memory.
- Strictly adhere to Argon2id KDF and Schema V1 database layout.

## Non-Goals

- Cloud recovery phrase or email-based account linkage.
