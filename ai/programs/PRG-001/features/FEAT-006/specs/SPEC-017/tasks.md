---
type: tasks
for: SPEC-017
---

# Tasks: SPEC-017 (TUI Dashboard, Sidebar Navigation, Live Search & Record Detail View)

## TASK-001 — Sidebar Navigation Component & Dynamic Tag Counting
- [x] Requirements: `SPEC-017:R1`
- Dependencies: none
- Scope: `internal/tui/screens/sidebar.go`
- Verification: `go test -v ./internal/tui/screens/... -run TestSidebar`
- Evidence: `go test -v ./internal/tui/screens/... -run TestSidebar` passed (`TestSidebarModel` 0.00s). Category list, custom tag discovery, record counts, and keyboard selection emission verified.
- Details: Implement `internal/tui/screens/sidebar.go` managing category tree/list (`All Items`, `Logins`, `Notes`, `API Keys`, `Custom`, `Trash`) with item counts and dynamic custom tags. Support arrow and Vim key (`j`/`k`) navigation, selection emission, and focus styling.

## TASK-002 — Detail Inspector Card & History Revision Timeline
- [x] Requirements: `SPEC-017:R3`
- Dependencies: none
- Scope: `internal/tui/screens/detail.go`, `internal/tui/screens/history.go`
- Verification: `go test -v ./internal/tui/screens/... -run TestDetail`
- Evidence: `go test -v ./internal/tui/screens/... -run "TestDetail|TestHistory"` passed (`TestDetailModel_ToggleSecret` 0.00s, `TestHistoryModel_LoadAndBrowse` 0.42s). Masked secret formatting, visibility toggle (`v`), and historical revision browsing (`H`) verified.
- Details: Implement `internal/tui/screens/detail.go` rendering structured secret card with masked-by-default secrets (`••••••••••••`) and plaintext visibility toggle on `v`. Implement `internal/tui/screens/history.go` revision timeline browser triggered by `H` fetching historical versions from `RecordService.GetHistory`.

## TASK-003 — Split-Pane Dashboard Table, Incremental Live Search & Clipboard Hotkeys
- [x] Requirements: `SPEC-017:R1`, `SPEC-017:R2`, `SPEC-017:R4`
- Dependencies: TASK-001, TASK-002
- Scope: `internal/tui/screens/dashboard.go`
- Verification: `go test -v ./internal/tui/screens/... -run TestDashboard`
- Evidence: `go test -v ./internal/tui/screens/... -run TestDashboard` passed (`TestDashboardModel_SearchFilterAndClipboard` 0.44s). Split-pane table layout, live fuzzy search ranking on `/`, pane switching, and clipboard copying (`c` with 45s auto-clear, `u`) verified.
- Details: Implement `internal/tui/screens/dashboard.go` embedding `SidebarModel`, `bubbles/table` for records, live search input triggered by `/` (delegating to `service.FilterAndRankRecords`), panel switching (`Tab`/`Shift+Tab`, `h`/`l`), detail inspector panel, and clipboard actions (`c` for primary secret with 45-second background auto-clear, `u` for username).

## TASK-004 — Root App Router Dashboard Integration & End-to-End Verification
- [x] Requirements: `SPEC-017:R1`, `SPEC-017:R2`, `SPEC-017:R3`, `SPEC-017:R4`
- Dependencies: TASK-003
- Scope: `internal/tui/app.go`, `internal/tui/dashboard_test.go`
- Verification: `go test -v -race ./internal/tui/...`
- Evidence: `go test -v -race ./internal/tui/...` passed cleanly with 0 race conditions (`TestAppModel_EndToEndDashboardWorkflow` passed). End-to-end category switching, live fuzzy search ranking on `/`, detail toggle with `v`, clipboard copying with `c` and `u`, revision history modal with `H`, and auto-lock zeroization verified.
- Details: Wire `DashboardModel` into `AppModel` under `ScreenDashboard`, propagating window size, unlock session record refresh, auto-lock activity reset, and toast/statusbar status. Add comprehensive test suite covering category switching, search typing, unmask toggle, clipboard copying, and revision inspection.
