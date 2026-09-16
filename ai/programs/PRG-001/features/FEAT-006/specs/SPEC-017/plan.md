# Implementation Plan: SPEC-017 (TUI Dashboard, Sidebar Navigation, Live Search & Record Detail View)

## Summary

This plan defines the technical architecture, component design, and implementation sequence for the GoVault TUI interactive workspace. It implements the two-pane dashboard model (`DashboardModel`): a left sidebar (`SidebarModel`) for category and tag filtering, a central records table (`bubbles/table`) with instant in-memory fuzzy search (`/`), a full-detail record inspector (`DetailModel`) with masked-by-default secret values and toggleable reveal (`v`), clipboard hotkeys (`c`, `u`) with automatic 45-second background clearing, and an interactive revision history browser (`HistoryModel`) triggered by `H`.

## Repository Context

- `internal/service/record_service.go`: Provides secret retrieval, listing, history retrieval, and category counting.
- `internal/service/search.go`: `FilterAndRankRecords` provides in-memory fuzzy multi-criteria scoring and filtering.
- `internal/service/clipboard.go`: `ClipboardService` manages clipboard driver integration and background auto-clear timers.
- `internal/tui/app.go`: Root `AppModel` router where `ScreenDashboard` will be wired to delegate to `DashboardModel`.
- `internal/tui/theme/theme.go`: Terminal styles, panels, badges, and layout helpers.

## Requirement Coverage

- **R1 -> Split-Pane Main Dashboard:**
  - Implemented in `internal/tui/screens/dashboard.go` (`DashboardModel`) and `internal/tui/screens/sidebar.go` (`SidebarModel`).
  - Left panel: Tree/list of categories (`All Items`, `Logins`, `Notes`, `API Keys`, `Custom`, `Trash`) with item count badges and dynamically loaded custom tags.
  - Main panel: Responsive `bubbles/table` showing columns: Title, Type, Tags, and Updated At.
  - Keyboard navigation: `Tab` / `Shift+Tab` or `h` / `l` switches active focus between Sidebar and Table.
- **R2 -> Incremental Live Search & Filtering:**
  - Search bar embedded at the top of the records table using `bubbles/textinput`.
  - Pressing `/` focuses search bar; typing updates in-memory filter in real-time via `service.FilterAndRankRecords(cachedRecords, filter)` without hitting SQLite on every keystroke.
  - Pressing `Esc` clears/unfocuses search input and restores standard category view.
- **R3 -> Record Detail & History Inspector:**
  - Implemented in `internal/tui/screens/detail.go` (`DetailModel`) and `internal/tui/screens/history.go` (`HistoryModel`).
  - Formatted detail view card showing Title, Type, Username, masked Password (`••••••••••••`), URLs, Notes, and custom key-value pairs.
  - Pressing `v` toggles plaintext visibility for sensitive fields.
  - Pressing `H` opens historical revision timeline, loading historical revisions via `RecordService.GetHistory(ctx, recordID)` and allowing side-by-side or sequential inspection.
- **R4 -> Clipboard Action Hotkeys:**
  - Pressing `c`: Copies primary secret (password, API secret, note text) using `service.ClipboardService.CopyWithAutoClear(45s)`.
  - Pressing `u`: Copies username or identifier to clipboard.
  - Renders transient confirmation toast or status bar alert (e.g. `✓ Password copied (clears in 45s)`).

## Architecture

```
                                      [AppModel (Root Router)]
                                                 │
                                                 ▼
                                     [DashboardModel (Screen)]
                   ┌─────────────────────────────┴─────────────────────────────┐
                   ▼                                                           ▼
          [SidebarModel (Left)]                                    [RecordsPane (Right)]
    - Categories: All, Logins, Notes,                        ├── [SearchInput] (activated on `/`)
      API Keys, Custom, Trash                                ├── [bubbles/table] (records list)
    - Custom Tags with counts                                ├── [DetailModel] (view/reveal with `v`)
    - Focus: h / Tab                                         └── [HistoryModel] (revision browser `H`)
                   │                                                           │
                   └─────────────────────────────┬─────────────────────────────┘
                                                 ▼
                                     [Application Services]
                              - service.RecordService (List/History)
                              - service.FilterAndRankRecords
                              - service.ClipboardService
```

## Components Affected

- `internal/tui/screens/sidebar.go`: Sidebar navigation component for category and tag selection.
- `internal/tui/screens/detail.go`: Detail inspector card with masked/unmasked toggle and field extraction.
- `internal/tui/screens/history.go`: Historical revision modal/viewer model.
- `internal/tui/screens/dashboard.go`: Split-pane dashboard model embedding sidebar, search, table, detail, and clipboard actions.
- `internal/tui/app.go`: Wire `DashboardModel` into `AppModel`'s `ScreenDashboard` branch.

## Data Changes

None. Purely in-memory presentation and caching over existing SQLite schema and domain models.

## API Changes

- Package `internal/tui/screens` additions:
  - `NewDashboardModel(th *theme.Theme, rs *service.RecordService, cs *service.ClipboardService, width, height int) *DashboardModel`
  - `NewSidebarModel(th *theme.Theme) *SidebarModel`
  - `NewDetailModel(th *theme.Theme) *DetailModel`
  - `NewHistoryModel(th *theme.Theme, rs *service.RecordService) *HistoryModel`

## Integration Changes

- Integrate system and in-memory clipboard drivers with auto-clear timers in the TUI loop.
- `AppModel` updates `DashboardModel` whenever the session unlocks and on window resize.

## Implementation Sequence

1. **Sidebar Navigation Model:** Implement `internal/tui/screens/sidebar.go` with category items, custom tags, item counting, and arrow/vim key navigation.
2. **Detail Inspector & History Model:** Implement `internal/tui/screens/detail.go` (formatted card, mask toggle `v`) and `internal/tui/screens/history.go` (revision browsing).
3. **Split-Pane Dashboard & Live Search:** Implement `internal/tui/screens/dashboard.go` combining sidebar, `bubbles/table`, live search input (`/`), detail view, and clipboard integration (`c`, `u`).
4. **App Integration & Routing:** Wire `DashboardModel` into `internal/tui/app.go` with key delegation, auto-lock refresh, and resize handling.
5. **Dashboard Test Suite:** Write comprehensive unit and event tests in `internal/tui/screens/dashboard_test.go` and `internal/tui/dashboard_test.go`.

## Test Strategy

- **Sidebar Navigation Tests:** Verify category and tag selection, count updates, and key transitions.
- **Live Search Tests:** Verify in-memory typing updates table rows with ranked results.
- **Detail & Mask Toggle Tests:** Verify sensitive fields start masked and toggle plaintext on `v`.
- **Clipboard Tests:** Verify `c` and `u` hotkeys write to `ClipboardService` and trigger auto-clear timers.
- **Race Condition Verification:** Execute `go test -v -race ./internal/tui/...`.

## Risks

- *Risk:* Table rendering overflow on narrow terminals.
  *Mitigation:* Dynamically calculate column widths based on available right-pane width with minimum column truncations.
- *Risk:* Clipboard driver incompatibility on headless Linux.
  *Mitigation:* `ClipboardService` gracefully falls back without panicking; toast displays copy outcome.

## Assumptions

- Decrypted records list comfortably fits in memory during active vault session.
