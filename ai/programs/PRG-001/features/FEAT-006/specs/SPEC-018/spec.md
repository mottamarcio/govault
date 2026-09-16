---
id: SPEC-018
type: spec
status: ready
parent: FEAT-006
depends_on:
  - SPEC-007
  - SPEC-010
  - SPEC-016
  - SPEC-017
supersedes: []
---

# SPEC-018: TUI Interactive Record Form Editor, Generator Overlay & Modal Dialogs

## Intent

Specify the interactive creation, editing, deletion, and utility overlays of the TUI: multi-field form editor (`ScreenEditor`) for typed secrets, embedded password/passphrase generator modal (`ScreenGenerator`) with live entropy meter, confirmation dialogs (delete/trash/purge), and help overlay (`?`).

## Requirements

- **R1: Interactive Record Form Editor (`ScreenEditor`):**
  - Multi-input form (`bubbles/textinput`, `bubbles/textarea`) supporting all secret types (`login`, `note`, `apikey`, `custom`).
  - Inputs for Title, Username, Password/Secret, URI, Notes, Tags (comma-separated or tag selector), and arbitrary custom key-value rows.
  - Tab navigation across form fields, `Ctrl+S` or `Enter` (on save button) to commit changes via `service.RecordService.Create` or `Update`.
  - Form validation: title cannot be empty, custom keys must be unique.
- **R2: Embedded Password & Passphrase Generator Overlay (`ScreenGenerator`):**
  - Modal overlay accessible directly or from password fields via hotkey (e.g. `Ctrl+G` / `g`).
  - Interactive controls to adjust length (slider / arrow keys), toggle character sets (`Upper`, `Lower`, `Digits`, `Symbols`, `Ambiguous`), or switch to Diceware Passphrase mode (word count, delimiter).
  - Live preview with real-time Shannon entropy (bits) and colorized strength meter (Red -> Green).
  - Pressing `Enter` accepts and copies/pastes the generated password directly into the active form field.
- **R3: Confirmation Modals & Trash Actions:**
  - Deletion prompt modal: `Move to Trash? [y/N]` (or permanent purge confirmation).
  - Restoration and purge actions accessible from the Trash sidebar view.
- **R4: Keybinding Help Overlay:**
  - Pressing `?` displays an accessible modal listing all navigation, action, and editing hotkeys.
  - Pressing `Esc` or `?` dismisses the overlay.

## Acceptance Scenarios

- **Scenario 1: Add New Secret with Generator Overlay**
  - *Given* the TUI dashboard
  - *When* the user presses `a`, fills in Title and Username, presses `Ctrl+G` to generate a 24-character password, and presses `Ctrl+S`
  - *Then* the record is encrypted and saved to the database, and the dashboard updates with the new item selected.
- **Scenario 2: Edit Existing Record**
  - *Given* an existing record selected in the dashboard
  - *When* the user presses `e`, updates the URL, and saves
  - *Then* the record version increments to 2, previous version is archived in history, and the detail view reflects the new URL.
- **Scenario 3: Delete Record with Confirmation**
  - *Given* a record selected in the dashboard
  - *When* the user presses `d`, confirms `y` in the confirmation modal
  - *Then* the record is soft-deleted to the trash and removed from the active items list.

## Edge Cases

- Canceling an edit form (`Esc`) prompts for confirmation if fields have unsaved changes (`Discard changes? [y/N]`).
- Custom field rows can be dynamically added (`Ctrl+N`) and deleted (`Ctrl+D`) in the editor.

## Constraints

- Modal overlays must cleanly restore the background dashboard state upon dismissal.
- Memory zeroization: sensitive input buffers in form models must be cleared upon exit or screen transition.

## Non-Goals

- External editor integration (e.g. launching `$EDITOR` for Markdown notes is deferred to future extensions).
