---
type: tasks
for: SPEC-019
---

# Tasks: SPEC-019 (TUI First-Run Setup & Interactive Vault Initialization Wizard)

## TASK-001 — VaultService State Query & Initialization Detection Helper
- [x] Requirements: `SPEC-019:R1`
- Dependencies: none
- Scope: `internal/service/vault_service.go`, `internal/service/vault_service_test.go`
- Verification: `go test -v ./internal/service/... -run TestVaultServiceIsInitialized` (PASS)
- Details: Implement `IsInitialized(ctx context.Context) (bool, error)` on `*service.VaultService` inspecting `MetadataRepository.Get` to detect if the vault database is initialized or fresh, returning `false` on `sqlite.ErrVaultNotInitialized`. Add unit tests in `internal/service/vault_service_test.go`.

## TASK-002 — Interactive First-Run Wizard Screen Model
- [x] Requirements: `SPEC-019:R2`, `SPEC-019:R3`
- Dependencies: none
- Scope: `internal/tui/screens/init.go`, `internal/tui/screens/init_test.go`
- Verification: `go test -v ./internal/tui/screens/... -run TestInitModel` (PASS)
- Details: Implement `internal/tui/screens/init.go` defining `InitModel` with master password input, confirmation input, live entropy & strength gauge, enforcing "Master password must be at least 8 characters long." and "Confirm password must match master password exactly." Submitting valid credentials invokes `VaultService.Init` + `Unlock`, zeroizes password buffers, and emits `InitSuccessMsg`.

## TASK-003 — Root App Routing, Auto-Detection & End-to-End Test Suite
- [x] Requirements: `SPEC-019:R1`, `SPEC-019:R2`, `SPEC-019:R3`, `SPEC-019:R4`
- Dependencies: TASK-001, TASK-002
- Scope: `internal/tui/types.go`, `internal/tui/app.go`, `internal/tui/app_test.go`
- Verification: `go test -v -race ./internal/tui/...` (PASS)
- Details: Add `ScreenInit` to `ScreenState` enum in `internal/tui/types.go`. Wire `InitModel` into `AppModel`, dynamically routing initial state to `ScreenInit` when uninitialized and `ScreenUnlock` when initialized. Handle `InitSuccessMsg` to transition directly to `ScreenDashboard`. Ensure "Pressing q or Ctrl+C on ScreenInit aborts the wizard and quits the application". Add comprehensive unit and integration tests in `app_test.go`.
