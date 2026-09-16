---
type: plan
for: SPEC-018
status: ready
---

# Implementation Plan: TUI Interactive Record Form Editor, Generator Overlay & Modal Dialogs

## Summary

Design and implement the full suite of interactive management and modal overlays for the GoVault Terminal User Interface (TUI):
1. **Interactive Multi-Field Record Form Editor (`ScreenEditor`):** Comprehensive form supporting all record types (`login`, `note`, `apikey`, `custom`) with `bubbles/textinput` and `bubbles/textarea`, dynamic field focus (`Tab`/`Shift+Tab`), dynamic custom field rows (`Ctrl+N`/`Ctrl+D`), form validation, and save/cancel shortcuts (`Ctrl+S`/`Esc`).
2. **Embedded Password & Passphrase Generator (`ScreenGenerator`):** Configurable interactive modal overlay supporting standard random password generation (length, char sets, ambiguous exclusion) and Diceware multi-word passphrases, real-time Shannon entropy calculation, colorized strength gauge, and seamless field insertion/clipboard export.
3. **Confirmation Dialogs & Trash Management (`ModalConfirm`):** Centered floating confirmation modals for destructive operations (move to trash, permanent purge, discard unsaved changes, restore from trash).
4. **Keybinding Help Overlay (`ModalHelp`):** Context-aware hotkey guide modal summoned via `?` and dismissed via `Esc` or `?`.

## Repository Context

- **Existing Services & Crypto:**
  - `internal/service/record_service.go`: `Create`, `Update`, `Trash`, `Restore`, `Purge`, `List`, `Get`.
  - `internal/service/password_gen.go`: `GeneratePassword`, `CalculateEntropy`, `EvaluateStrength`, `DefaultPasswordOptions`.
  - `internal/service/passphrase.go`: `GeneratePassphrase`, `DefaultPassphraseOptions`.
  - `internal/service/clipboard.go`: `Copy` with auto-clear timeout.
- **Existing TUI Components:**
  - `internal/tui/app.go`: Root `AppModel` state machine and router (`ScreenUnlock`, `ScreenDashboard`, `ScreenEditor`, `ScreenGenerator`).
  - `internal/tui/theme/theme.go`: Nord/Catppuccin color scheme, typography styles, panel borders, and badges.
  - `internal/tui/screens/dashboard.go`: Vault browser, sidebar category navigation, search bar, and detail panel.
  - `internal/tui/screens/statusbar.go`: Auto-lock countdown and dynamic key hints.

## Requirement Coverage

- **R1 → Multi-Field Record Form Editor (`ScreenEditor` in `internal/tui/screens/editor.go`):**
  - Create `EditorModel` supporting `ModeCreate` and `ModeEdit`.
  - Implements field models: Type selector, Title input, Username/Identity input, Password/Secret input (toggleable visibility `Ctrl+V`), URI input, Notes textarea (`bubbles/textarea`), Tags input (comma-separated), and dynamic Custom Key-Value slice.
  - Keyboard navigation: `Tab` advances focus, `Shift+Tab` reverses focus, `Ctrl+N` adds custom field, `Ctrl+D` removes focused custom field.
  - Form validation: title cannot be whitespace-only, custom field keys must be non-empty and unique. Displays error message if invalid.
  - Save via `Ctrl+S` or Enter on `[Save]` button; cancel via `Esc` (triggers dirty check).
  - Integrates with `service.RecordService.Create` and `Update`.

- **R2 → Embedded Password & Passphrase Generator Overlay (`ScreenGenerator` in `internal/tui/screens/generator.go`):**
  - Implement `GeneratorModel` modal overlay with dual modes: Random Password vs. Diceware Passphrase.
  - Interactive parameter adjustments: length (8–128), character toggles (`Upper`, `Lower`, `Digits`, `Symbols`, `Ambiguous`), passphrase word count (3–12), delimiter, capitalization.
  - Real-time entropy computation via `service.CalculateEntropy` and strength tier evaluation via `service.EvaluateStrength`.
  - Colorized visual strength bar (Red -> Yellow -> Green).
  - Re-generate on parameter change or `r` / Space.
  - `Enter` accepts and emits `GeneratorResultMsg`, inserting password directly into editor's secret field or copying to clipboard in standalone mode.

- **R3 → Confirmation Modals & Trash Actions (`ModalConfirm` in `internal/tui/screens/modal.go`):**
  - Reusable floating modal dialog centered over dimmed background.
  - Actions:
    - Trash confirmation: `Move "<title>" to Trash? [y/N]`
    - Purge confirmation: `Permanently delete "<title>"? This cannot be undone! [y/N]`
    - Restore confirmation: `Restore "<title>" to active vault? [y/N]`
    - Discard confirmation: `Discard unsaved changes? [y/N]`
  - Listens for `y`/`Y` (confirm) and `n`/`N`/`Esc` (cancel), emitting `ModalConfirmMsg`.
  - Connected to dashboard actions: `d` key triggers trash/purge, `r` key in Trash view triggers restore.

- **R4 → Keybinding Help Overlay (`ModalHelp` in `internal/tui/screens/help.go`):**
  - Overlay displaying organized table of hotkeys: Navigation, Vault Actions, Editing, Clipboard, Overlays.
  - Activated by `?` anywhere in dashboard or editor.
  - Dismissed by `?` or `Esc`.

## Architecture

```
+-------------------------------------------------------------------------------+
|                                  AppModel                                     |
|                                                                               |
|  State: ScreenUnlock | ScreenDashboard | ScreenEditor | ScreenGenerator       |
|                                                                               |
|  +---------------------------+   +-----------------------------------------+  |
|  |       DashboardModel      |   |               EditorModel               |  |
|  | - Sidebar                 |   | - Title, Username, Password, URI        |  |
|  | - Record Table            |   | - Notes (textarea), Tags                |  |
|  | - Detail Panel            |   | - Custom Fields list                    |  |
|  | - History Overlay         |   | - Validation & Dirty check              |  |
|  +---------------------------+   +-----------------------------------------+  |
|               |                                       |                       |
|               +-------------------+-------------------+                       |
|                                   |                                           |
|                  +---------------------------------+                          |
|                  |     Modal Overlays / Popups     |                          |
|                  | - ModalConfirm (Trash/Purge)    |                          |
|                  | - ModalHelp (Hotkeys ?)         |                          |
|                  | - GeneratorModel (Ctrl+G)       |                          |
|                  +---------------------------------+                          |
+-------------------------------------------------------------------------------+
```

### Overlay Layering Strategy
Modals render as framed floating boxes positioned with Lipgloss `Place` or horizontal/vertical joins centered against the terminal window dimensions (`Width`, `Height`), overlaying the background screen.

## Components Affected

1. `internal/tui/types.go`:
   - Extend `ScreenState` and define shared event messages (`OpenEditorMsg`, `EditorSaveMsg`, `EditorCancelMsg`, `OpenGeneratorMsg`, `GeneratorResultMsg`, `ModalConfirmMsg`).
2. `internal/tui/screens/editor.go` & `editor_test.go`:
   - New `EditorModel` implementing Bubble Tea `Init`, `Update`, `View` for record editing.
3. `internal/tui/screens/generator.go` & `generator_test.go`:
   - New `GeneratorModel` implementing Bubble Tea modal with entropy meter and parameter sliders.
4. `internal/tui/screens/modal.go` & `modal_test.go`:
   - Generic confirmation modal component (`ModalConfirm`).
5. `internal/tui/screens/help.go` & `help_test.go`:
   - Keybindings reference modal (`ModalHelp`).
6. `internal/tui/screens/dashboard.go` & `dashboard_test.go`:
   - Wire `a` (add), `e` (edit), `d` (trash/purge), `r` (restore), `?` (help), `g` (generator).
7. `internal/tui/app.go` & `app_test.go`:
   - Wire routing for `ScreenEditor` and modal overlay handling.

## Data Changes

None. Uses existing SQLite tables (`records`, `metadata`, `record_tags`, `record_history`) through `service.RecordService`.

## API Changes

None. TUI interacts exclusively with existing `service.RecordService`, `service.VaultService`, and `service.ClipboardService` APIs.

## Integration Changes

- In `DashboardModel`, pressing `a` dispatches `OpenEditorMsg{Mode: ModeCreate}`.
- In `DashboardModel`, pressing `e` dispatches `OpenEditorMsg{Mode: ModeEdit, Record: selectedRecord}`.
- In `DashboardModel`, pressing `d` opens `ModalConfirm` for trash (or purge if in Trash category).
- In `DashboardModel` (Trash view), pressing `r` opens `ModalConfirm` for restore.
- In `DashboardModel` or `EditorModel`, pressing `?` opens `ModalHelp`.
- In `EditorModel`, pressing `Ctrl+G` opens `GeneratorModel`; upon completion, the generated secret replaces the secret input value.
- In `AppModel`, `Update` coordinates transitions between `ScreenDashboard`, `ScreenEditor`, and modal overlays.

## Implementation Sequence

1. **Step 1: Modal Primitives (`screens/modal.go` and `screens/help.go`):**
   - Implement `ModalConfirm` for destructive prompts.
   - Implement `ModalHelp` for keybinding references.
2. **Step 2: Password & Passphrase Generator Overlay (`screens/generator.go`):**
   - Implement parameter controls, live entropy rendering, and result propagation.
3. **Step 3: Multi-Field Record Form Editor (`screens/editor.go`):**
   - Build form inputs, custom field row manager, validation rules, and dirty tracking.
4. **Step 4: Dashboard & App Model Wiring (`dashboard.go`, `app.go`):**
   - Connect hotkeys (`a`, `e`, `d`, `r`, `?`, `Ctrl+G`), message loops, screen transitions, and status bar hints.
5. **Step 5: Automated Verification & Unit Tests:**
   - Add unit tests for form input validation, generator options/entropy updates, confirmation dialog state transitions, and root app routing.

## Test Strategy

- **Unit Tests (`screens/editor_test.go`):**
  - Verify initialization in Create and Edit modes across all secret types (`login`, `note`, `apikey`, `custom`).
  - Verify `Tab`/`Shift+Tab` focus cycling through all fields and custom rows.
  - Verify field validation (empty title rejected, duplicate custom keys rejected).
  - Verify `Ctrl+S` compiles payload and dispatches `EditorSaveMsg`.
- **Unit Tests (`screens/generator_test.go`):**
  - Verify password length and character set toggles affect generated candidate and entropy calculation.
  - Verify passphrase mode generates valid multi-word passphrases.
  - Verify `Enter` dispatches `GeneratorResultMsg`.
- **Unit Tests (`screens/modal_test.go` & `screens/help_test.go`):**
  - Verify modal confirmation emits confirm/cancel messages on `y` and `n`/`Esc`.
  - Verify help modal opens and closes properly on `?`/`Esc`.
- **Integration Tests (`app_test.go` & `screens/dashboard_test.go`):**
  - Verify end-to-end flow: dashboard -> press `a` -> editor -> save -> record added to dashboard list.
  - Verify edit flow: select record -> press `e` -> edit -> save -> updated record reflected.
  - Verify delete flow: select record -> press `d` -> confirm `y` -> record moved to trash.

## Risks

- **Terminal Size Limitations:** If the terminal is small (<80 columns, <24 rows), multi-field forms or generator overlays could overflow.
  *Mitigation:* Use compact vertical layouts, clamp inputs to visible viewport, and utilize scrolling/clipping where needed.
- **Plaintext Secret Retention in Editor Memory:**
  *Mitigation:* Implement `Reset()` on `EditorModel` and `GeneratorModel` to zero out text buffers on cancel, save, or vault lock.

## Assumptions

- Terminal supports ANSI escape sequences for UTF-8 borders and color rendering.
- `RecordService` payload types (`domain.LoginPayload`, `domain.NotePayload`, `domain.APIKeyPayload`, `domain.CustomPayload`) can be safely constructed from form string maps.
