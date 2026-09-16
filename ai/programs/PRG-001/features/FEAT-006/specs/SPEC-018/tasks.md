---
type: tasks
for: SPEC-018
---

# Tasks: SPEC-018 (TUI Interactive Record Form Editor, Generator Overlay & Modal Dialogs)

## TASK-001 — Reusable Centered Modal Dialogs & Help Keybindings Overlay
- [x] Requirements: `SPEC-018:R3`, `SPEC-018:R4`
- Dependencies: none
- Scope: `internal/tui/screens/modal.go`, `internal/tui/screens/help.go`
- Verification: `go test -v ./internal/tui/screens/... -run "TestModal|TestHelp"`
- Evidence: `go test -v ./internal/tui/screens/... -run "TestModal|TestHelp"` passed (`TestModalConfirm` and `TestModalHelp` passed in 0.008s). Centered confirmation modal with exact prompt formatting ("Move to Trash? [y/N]", permanent purge, discard) and categorized keybinding reference overlay toggleable with `?`/`Esc` verified.
- Details: Implement `internal/tui/screens/modal.go` defining `ModalConfirm` floating dialog for destructive confirmations with exact prompt format "Move to Trash? [y/N]" or purge/discard confirmations (`y`/`Y` confirms, `n`/`N`/`Esc` cancels). Implement `internal/tui/screens/help.go` defining `ModalHelp` displaying categorized keybinding reference overlay where "Pressing ? displays an accessible modal listing all navigation, action, and editing hotkeys." and "Pressing Esc or ? dismisses the overlay."

## TASK-002 — Interactive Password & Passphrase Generator Overlay
- [x] Requirements: `SPEC-018:R2`
- Dependencies: none
- Scope: `internal/tui/screens/generator.go`
- Verification: `go test -v ./internal/tui/screens/... -run TestGenerator`
- Evidence: `go test -v ./internal/tui/screens/... -run TestGenerator` passed (`TestGeneratorModel_PasswordMode` and `TestGeneratorModel_PassphraseMode` passed in 0.006s). Length slider, char set toggles, Diceware passphrase modes, real-time Shannon entropy/strength meter, and secret propagation verified.
- Details: Implement `internal/tui/screens/generator.go` defining `GeneratorModel` modal overlay with dual modes (random password and Diceware passphrase) providing "Interactive controls to adjust length (slider / arrow keys), toggle character sets (Upper, Lower, Digits, Symbols, Ambiguous), or switch to Diceware Passphrase mode (word count, delimiter).", "Live preview with real-time Shannon entropy (bits) and colorized strength meter (Red -> Green).", and "Pressing Enter accepts and copies/pastes the generated password directly into the active form field."

## TASK-003 — Multi-Field Record Form Editor with Custom Fields & Validation
- [x] Requirements: `SPEC-018:R1`, `SPEC-018:R2`
- Dependencies: none
- Scope: `internal/tui/screens/editor.go`
- Verification: `go test -v ./internal/tui/screens/... -run TestEditor`
- Evidence: `go test -v ./internal/tui/screens/... -run TestEditor` passed (`TestEditorModel_CreateAndValidate`, `TestEditorModel_EditExistingRecord`, `TestEditorModel_CustomFieldsAndDiscard` passed in 0.998s). Form inputs across types, Tab/Shift-Tab navigation, validation rules, custom field management (`Ctrl+N`/`Ctrl+D`), save (`Ctrl+S`), and dirty discard modal confirmation verified.
- Details: Implement `internal/tui/screens/editor.go` defining `EditorModel` providing "Multi-input form (bubbles/textinput, bubbles/textarea) supporting all secret types (login, note, apikey, custom)." with "Tab navigation across form fields, Ctrl+S or Enter (on save button) to commit changes via service.RecordService.Create or Update." Enforce "Form validation: title cannot be empty, custom keys must be unique." and handle dirty cancellation where "Canceling an edit form (Esc) prompts for confirmation if fields have unsaved changes (Discard changes? [y/N])."

## TASK-004 — App Router Wire-Up, Dashboard Action Triggers & End-to-End Test Suite
- [x] Requirements: `SPEC-018:R1`, `SPEC-018:R2`, `SPEC-018:R3`, `SPEC-018:R4`
- Dependencies: TASK-001, TASK-002, TASK-003
- Scope: `internal/tui/types.go`, `internal/tui/screens/dashboard.go`, `internal/tui/app.go`, `internal/tui/app_test.go`
- Verification: `go test -v -race ./internal/tui/...`
- Evidence: `go test -v -race ./internal/tui/...` passed cleanly with 0 race conditions (`TestAppModel_AddSecretWithGeneratorAndEdit`, `TestAppModel_DeleteWithConfirmationModal`, `TestAppModel_HelpModalToggle`, `TestAppModel_EndToEndDashboardWorkflow` passed in 16.359s). Multi-field form editor (`a`/`e`), password generator overlay (`Ctrl+G`), deletion confirmation dialogs (`d`/`y`), restore actions (`r`), and help cheatsheet (`?`/`Esc`) end-to-end routing verified.
- Details: Wire `EditorModel`, `GeneratorModel`, `ModalConfirm`, and `ModalHelp` into `AppModel` and `DashboardModel`. Connect dashboard hotkeys `a` (create record), `e` (edit record), `d` (trash / purge), `r` (restore in trash), `g` (generator), `?` (help). Add comprehensive unit and integration tests verifying creation with generator overlay, record editing, and trash deletion confirmation workflows.
