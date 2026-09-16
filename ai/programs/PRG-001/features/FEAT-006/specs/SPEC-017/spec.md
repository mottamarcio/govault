---
id: SPEC-017
type: spec
status: ready
parent: FEAT-006
depends_on:
  - SPEC-007
  - SPEC-009
  - SPEC-010
  - SPEC-016
supersedes: []
---

# SPEC-017: TUI Dashboard, Sidebar Navigation, Live Search & Record Detail View

## Intent

Specify the main TUI workspace: sidebar for categories (Logins, Notes, API Keys, Custom) and tags, main filterable records table/list with fuzzy incremental search (`/`), record detail inspector with toggleable password unmasking (`v`), clipboard hotkeys (`c`, `u`), and revision history browser.

## Requirements

- **R1: Split-Pane Main Dashboard:**
  - Sidebar panel (left): Tree/list of record categories (`All Items`, `Logins`, `Notes`, `API Keys`, `Custom`, `Trash`) and custom tags with unread/item counts.
  - Main panel (right): Responsive `bubbles/table` or `bubbles/list` displaying matching records with columns for Title, Type, Tags, and Updated date.
  - Keyboard navigation between panels using `Tab` / `Shift+Tab` or `h`/`l` (Vim keys).
- **R2: Incremental Live Search & Filtering:**
  - Pressing `/` activates the search bar at the top of the records table.
  - Live in-memory fuzzy filtering as the user types, ranking matches via `service.FilterAndRankRecords`.
  - Pressing `Esc` clears/exits the search input and restores category view.
- **R3: Record Detail & History Inspector:**
  - Viewing selected record displays formatted card with Title, Type, Username, masked Password (`••••••••••••`), URLs, Notes, and custom key-value pairs.
  - Pressing `v` toggles plaintext visibility for sensitive fields (password, API secret, custom masked fields).
  - Pressing `H` opens the historical revision timeline, allowing navigation across previous versions and inspection of previous secret values.
- **R4: Clipboard Action Hotkeys:**
  - `c`: Copy primary secret (password, API key, note body) to clipboard with automatic 45-second background clear.
  - `u`: Copy username or service identifier to clipboard.
  - Status bar or transient toast alert displays confirmation (e.g. `✓ Password copied (clears in 45s)`).

## Acceptance Scenarios

- **Scenario 1: Category and Tag Switching**
  - *Given* an unlocked dashboard with 10 records across Logins and Notes
  - *When* the user selects `Notes` in the sidebar
  - *Then* the main table updates immediately to show only Note records.
- **Scenario 2: Toggle Visibility and Copy Password**
  - *Given* a Login record selected in the dashboard
  - *When* the user presses `v`
  - *Then* the masked password is revealed in plaintext; pressing `c` copies it to the system clipboard with an auto-clear timer.
- **Scenario 3: Live Fuzzy Search**
  - *Given* 50 records in the vault
  - *When* the user presses `/` and types `git`
  - *Then* only records with titles or tags containing `git` (e.g. `GitHub`, `GitLab`) are listed in ranked order.

## Edge Cases

- Empty vault displays an onboarding card ("Press 'a' to add your first secret").
- Record with 0 historical revisions gracefully disables the history hotkey with a notice.

## Constraints

- Sensitive values must remain masked by default in the detail view unless explicitly toggled with `v`.
- Live search queries must operate in-memory on decrypted cache without re-querying SQLite on every keystroke.

## Non-Goals

- Remote synchronization triggers from TUI.
