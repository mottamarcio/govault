---
type: plan
for: SPEC-019
status: ready
---

# Implementation Plan: TUI First-Run Setup & Interactive Vault Initialization Wizard

## Summary

Design and implement first-run vault state detection and an interactive First-Time Setup Wizard (`ScreenInit`) for the GoVault TUI:
1. **Uninitialized Vault Detection (`R1`):** In `VaultService` and `AppModel`, check whether metadata exists in the vault database. If uninitialized or the database file is missing, automatically launch into `ScreenInit`; otherwise route to `ScreenUnlock`.
2. **Interactive Setup Wizard Model (`ScreenInit` in `screens/init.go`, `R2`):** Build a dedicated Bubble Tea model featuring master password input (`textinput`), confirmation password input (`textinput`), live entropy & strength meter, explanatory security notices, and validation rules (minimum 8 characters, exact confirmation matching).
3. **Cryptographic Initialization & Auto-Unlock (`R3`):** On valid submission, invoke `VaultService.Init` (Argon2id KDF + SQLite schema migration), automatically unlock the session with `VaultService.Unlock`, zeroize memory buffers, and transition directly into `ScreenDashboard`.
4. **Clean Cancellation & Exit (`R4`):** Support safe abort on `q` or `Ctrl+C`.

## Repository Context

- **Existing Services & Storage:**
  - `internal/service/vault_service.go`: `Init(ctx, masterPassword)`, `Unlock(ctx, masterPassword)`, `Session()`.
  - `internal/storage/sqlite/metadata_repo.go`: `Get(ctx)` returning `ErrVaultNotInitialized` when empty.
  - `internal/service/password_gen.go`: `CalculateEntropy(secret)`, `EvaluateStrength(entropy)`.
- **Existing TUI Architecture:**
  - `internal/tui/types.go`: `ScreenState` enum (`ScreenUnlock`, `ScreenDashboard`, `ScreenEditor`, `ScreenGenerator`).
  - `internal/tui/app.go`: Root `AppModel` routing state machine and lifecycle manager.
  - `internal/tui/screens/unlock.go`: Master password unlock screen model.
  - `internal/tui/theme/theme.go`: Terminal styling palette, typography, badges, and boxes.

## Requirement Coverage

- **R1 → Uninitialized Vault Detection & Initial Routing:**
  - Add `IsInitialized(ctx context.Context) (bool, error)` to `VaultService`.
  - In `NewApp` / `AppModel.Init()`, query `VaultService.IsInitialized(ctx)`. If `false`, set `State = ScreenInit`; if `true`, set `State = ScreenUnlock`.
- **R2 → First-Run Setup Screen & Confirmation Inputs (`ScreenInit` in `internal/tui/screens/init.go`):**
  - Create `InitModel` with two `textinput.Model` instances (Password and Confirm Password) masked with `•`.
  - Live entropy and strength badge updating as the user types in the Password field.
  - Keyboard navigation: `Tab`, `Shift+Tab`, `↑`, `↓` to switch focus between fields.
  - Validation: reject passwords shorter than 8 characters and reject mismatched confirmations with visible inline error badges.
- **R3 → Cryptographic Vault Initialization & Auto-Unlock:**
  - Submitting a valid form invokes `VaultService.Init(ctx, pass)` to run SQLite migrations and generate the Argon2id encryption envelope.
  - Automatically invokes `VaultService.Unlock(ctx, pass)` to create an active authenticated `Session`.
  - Zeroizes input text buffers.
  - Emits `InitSuccessMsg{Session: session}`, which causes `AppModel` to switch `State = ScreenDashboard`, reload records, start the auto-lock timer, and configure the status bar.
- **R4 → Safe Cancellation & Clean Exit:**
  - Pressing `q` or `Ctrl+C` in `ScreenInit` sets `Quitting = true` and exits Bubble Tea cleanly.

## Architecture

```
                                  Startup Check
                        +--------------------------------+
                        |  VaultService.IsInitialized()  |
                        +--------------------------------+
                                   /           \
                           (false)/             \(true)
                                 v               v
                   +-------------------+   +--------------------+
                   |    ScreenInit     |   |    ScreenUnlock    |
                   | (First-Run Setup) |   | (Master Pass Auth) |
                   +-------------------+   +--------------------+
                             |                       |
                  (Init & Auto-Unlock)         (Unlock Auth)
                             \                       /
                              v                     v
                        +--------------------------------+
                        |        ScreenDashboard         |
                        |      (Active Vault Session)    |
                        +--------------------------------+
```

## Components Affected

1. `internal/service/vault_service.go` & `vault_service_test.go`:
   - Add `IsInitialized(ctx context.Context) (bool, error)` method.
2. `internal/tui/types.go`:
   - Add `ScreenInit` to `ScreenState` enum.
3. `internal/tui/screens/init.go` & `init_test.go`:
   - New `InitModel` implementing the first-run wizard UI, confirmation inputs, entropy meter, and initialization command.
4. `internal/tui/app.go` & `app_test.go`:
   - Add `InitScreen *screens.InitModel` to `AppModel`.
   - Update startup state selection logic (`ScreenInit` vs `ScreenUnlock`).
   - Handle `InitSuccessMsg` and route `ScreenInit` update/view events.

## Data Changes

None. Uses existing SQLite Schema V1 and metadata tables.

## API Changes

- `*service.VaultService.IsInitialized(ctx context.Context) (bool, error)` added as a non-breaking helper.

## Integration Changes

- In `internal/cli/tui.go` and `internal/tui/app.go`, ensure that missing database files trigger directory creation (`0700`) and SQLite database open in preparation for initialization or unlock.

## Implementation Sequence

1. **Step 1: VaultService Helper (`internal/service/vault_service.go`):**
   - Implement `IsInitialized` and add unit test.
2. **Step 2: Init Screen Model (`internal/tui/screens/init.go`):**
   - Implement `InitModel` with password, confirmation, entropy gauge, validation, and `InitSuccessMsg`.
   - Add unit tests in `internal/tui/screens/init_test.go`.
3. **Step 3: Root App Routing & Integration (`internal/tui/app.go`, `internal/tui/types.go`):**
   - Add `ScreenInit` to state router, wire `InitScreen`, handle auto-routing on uninitialized vaults, and handle `InitSuccessMsg`.
4. **Step 4: End-to-End Verification & Integration Tests (`internal/tui/app_test.go`):**
   - Add integration tests verifying uninitialized vault detection, password mismatch validation, short password error, and successful first-run setup flow directly into dashboard.

## Test Strategy

- **Unit Tests (`internal/service/vault_service_test.go`):**
  - Verify `IsInitialized` returns `false` on fresh uninitialized DB, and `true` after `Init`.
- **Unit Tests (`internal/tui/screens/init_test.go`):**
  - Verify `InitModel` input validation (empty, <8 chars, mismatch).
  - Verify Tab/Shift-Tab focus cycling between password and confirm fields.
  - Verify entropy/strength calculation updates on typing.
  - Verify valid submission invokes `Init` + `Unlock` and emits `InitSuccessMsg`.
- **Integration Tests (`internal/tui/app_test.go`):**
  - Verify opening uninitialized vault routes to `ScreenInit`.
  - Verify completing first-run setup transitions directly to `ScreenDashboard`.
  - Verify opening initialized vault routes to `ScreenUnlock`.

## Risks

- **Race on missing directory creation:**
  *Mitigation:* Ensure parent directory of `VaultPath` is created with `0700` before opening SQLite DB.
- **Sensitive data persistence in inputs:**
  *Mitigation:* Explicit zeroization of password buffers on completion or quit.

## Assumptions

- Terminal dimensions meet minimum required `minTerminalWidth` (60) and `minTerminalHeight` (15).
