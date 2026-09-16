# Implementation Plan: SPEC-014 (CLI Secret Record Operations & In-Memory Search)

## Summary

This plan defines the architecture, design, and implementation sequence for the GoVault CLI secret record operations and search commands (`add`, `get`/`show`, `list`, `search`, `edit`, `delete`, `trash`). These commands bridge user terminal interactions and shell scripting pipelines to the underlying `RecordService`, `SearchService`, and SQLite repository engines implemented in FEAT-002 and FEAT-003, with strict adherence to the project constitution regarding offline operation, password masking, zero-leak memory management, and deterministic error exit codes.

## Repository Context

- `internal/service/record_service.go`: Provides high-level CRUD, history retrieval, trash management, and filtering.
- `internal/service/search.go`: Implements fuzzy and prefix search with scoring and tag filtering.
- `internal/service/password_gen.go`: Generates secure passwords when requested during interactive/flag creation.
- `internal/service/vault_service.go` & `session.go`: Manages vault authentication and unlocked memory sessions.
- `internal/domain`: Defines `Record`, `RecordType`, `LoginPayload`, `NotePayload`, `APIKeyPayload`, `CustomPayload`, and custom `Field`.
- `internal/cli/root.go`, `output.go`, `prompt.go`: Root CLI framework, context, formatting (text/tabular/JSON), and masked/unmasked terminal inputs.

## Requirement Coverage

- **R1 -> Secret Record Creation (`govault add`):**
  - Implemented in `internal/cli/record_add.go` as `newAddCmd(appCtx)`.
  - Supports `--type` (`login`, `note`, `apikey`, `custom`), `--title`, `--username`, `--uri`, `--notes`, `--tag` (slice), and `--field key=value` (slice).
  - For `login` and `apikey`, if password/secret is omitted, securely prompt for password with confirmation.
  - Supports `--generate` / `-g` flag to auto-generate a cryptographically secure 20-character password.
  - Passes structured typed payload to `recordService.Create(ctx, input)`. Emits record summary or JSON.
- **R2 -> Secret Retrieval & Masking (`govault get` / `govault show`):**
  - Implemented in `internal/cli/record_get.go` as `newGetCmd(appCtx)` (with alias `show`).
  - Accepts `<id>` or exact `<title>` query to locate record.
  - Masks password and secret fields by default (`••••••••••••`) unless `--show` / `-s` is explicitly supplied.
  - Supports single-field extraction with `--field <name>` (e.g. `--field password` or `--field username` or custom field key). When paired with `--raw` / `-r`, prints only the exact plaintext value to `stdout` with 0 extra decorations or trailing newlines (except standard newline if desired, formatted strictly for Unix pipes).
  - Supports `--json` flag to print complete decrypted record representation.
- **R3 -> Listing & In-Memory Search (`govault list` / `govault search`):**
  - Implemented in `internal/cli/record_list.go` (`newListCmd(appCtx)`) and `internal/cli/record_search.go` (`newSearchCmd(appCtx)`).
  - `govault list`: Formats tabular list of active records (`ID`, `Type`, `Title`, `Tags`, `Updated`) or JSON array with `--type` and `--tag` filtering.
  - `govault search <query>`: Uses `service.SearchFilter` and in-memory relevance ranking, returning matches with scores or clean tabular/JSON representations.
- **R4 -> Modification, Deletion & Trash Management (`govault edit` / `govault delete` / `govault trash`):**
  - Implemented in `internal/cli/record_edit.go`, `internal/cli/record_delete.go`, and `internal/cli/record_trash.go`.
  - `govault edit <id>`: Fetches existing record, applies updated flag values (`--title`, `--username`, `--tag`, etc.) or prompts, and calls `recordService.Update`.
  - `govault delete <id>`: Soft-deletes record (moves to trash) by default via `recordService.Delete`, or hard-deletes when `--permanent` is specified via `recordRepo.PurgeDeleted` / delete logic.
  - `govault trash`: Subcommands `govault trash list`, `govault trash restore <id>`, and `govault trash purge` delegating to `recordService.ListTrash`, `Restore`, and `PurgeTrash`.

## Architecture

```
                       [CLI User / Script Pipeline]
                                    │
                                    ▼
                          [Cobra Command Tree]
      ┌───────────┬───────────┬───────────┬───────────┬───────────┐
      │  addCmd   │  getCmd   │  listCmd  │ searchCmd │  editCmd  │ trashCmd
      └─────┬─────┴─────┬─────┴─────┬─────┴─────┬─────┴─────┬─────┴────┬──────┘
            │           │           │           │           │          │
            └───────────┴─────┬─────┴───────────┴───────────┴──────────┘
                              ▼
                     [AppContext / Helpers]
            - Prompt / Masked Password Authentication
            - Vault Opening & Session Unlock Helper (cliHelper)
            - OutputFormatter (Text / Tabular / JSON / Raw)
                              │
                              ▼
                    [service.RecordService]
              ├── Decrypts / Encrypts records with HKDF + XChaCha20-Poly1305
              ├── Evaluates in-memory search scoring (service.Search)
              └── Interacts with sqlite.RecordRepository & TagRepository
```

## Components Affected

- `internal/cli/root.go`: Register `add`, `get` (alias `show`), `list`, `search`, `edit`, `delete`, `trash` subcommands.
- `internal/cli/helpers.go`: Shared helper to prompt master password (or read `GOVAULT_PASSWORD`), unlock `VaultService`, instantiate `RecordService`, and ensure cleanup/locking.
- `internal/cli/record_add.go`: Implementation of `govault add`.
- `internal/cli/record_get.go`: Implementation of `govault get` and `govault show`.
- `internal/cli/record_list.go`: Implementation of `govault list`.
- `internal/cli/record_search.go`: Implementation of `govault search`.
- `internal/cli/record_edit.go`: Implementation of `govault edit`.
- `internal/cli/record_delete.go`: Implementation of `govault delete`.
- `internal/cli/record_trash.go`: Implementation of `govault trash` subcommands (`list`, `restore`, `purge`).
- `internal/cli/output.go`: Tabular formatting helper for record listings.

## Data Changes

None at the schema level. All underlying SQLite tables (`records`, `record_tags`, `tags`, `record_history`) and domain payloads are already defined and tested in FEAT-002 and FEAT-003.

## API Changes

- Exported helper or internal command registration functions in package `cli`:
  - `newAddCmd(appCtx *AppContext) *cobra.Command`
  - `newGetCmd(appCtx *AppContext) *cobra.Command`
  - `newListCmd(appCtx *AppContext) *cobra.Command`
  - `newSearchCmd(appCtx *AppContext) *cobra.Command`
  - `newEditCmd(appCtx *AppContext) *cobra.Command`
  - `newDeleteCmd(appCtx *AppContext) *cobra.Command`
  - `newTrashCmd(appCtx *AppContext) *cobra.Command`

## Integration Changes

- Seamless integration with Unix pipelines: `--raw` output on `govault get <id> --field <name>` writes raw bytes with no decoration or metadata.
- Error outputs routed to `stderr` with exit code 1.

## Implementation Sequence

1. **CLI Session & Unlock Helpers:** Implement unified helper in `internal/cli/session_helper.go` to authenticate the vault and provide active `RecordService` with defer teardown.
2. **Record Creation (`govault add`):** Implement flag parsing, password generation option, typed payload assembly, and record creation.
3. **Record Retrieval (`govault get` / `govault show`):** Implement record lookup by ID/title, masking logic, field extraction (`--field`, `--raw`), and JSON output.
4. **Listing & Searching (`govault list` & `govault search`):** Implement tabular printing, type/tag filters, and in-memory fuzzy/prefix search.
5. **Modification & Deletion (`govault edit`, `govault delete`):** Implement payload mutation, version incrementing, and soft/hard deletion.
6. **Trash Management (`govault trash`):** Implement `trash list`, `trash restore`, and `trash purge`.
7. **Root Command Wiring & Unit Tests:** Connect all commands in `root.go`, write comprehensive unit and end-to-end CLI tests in `internal/cli/record_test.go` checking race conditions and output formatting.

## Test Strategy

- **Interactive & Headless Tests:** Use `AppContext` with mock `bytes.Buffer` for `In`, `Out`, `Err` and `MockPrompter` to simulate terminal inputs.
- **Masking Verifications:** Assert that unauthenticated / default `govault get` outputs contain masked asterisks or bullets, and plaintext only appears when `--show` or `--raw` is used.
- **Pipeline `--raw` Tests:** Test stdout output matching exact field contents without trailing labels.
- **Search & Filtering Tests:** Create fixtures with various tags/types and verify that `list` and `search` return expected subsets in both table and JSON modes.
- **Trash Lifecycle Tests:** Add -> Delete -> Trash List -> Restore -> Get -> Trash Purge.
- **Race Condition Verification:** Execute `go test -v -race ./internal/cli/...`.

## Risks

- *Risk:* Accidental plaintext output in shell scripts or error messages.
  *Mitigation:* Masking is applied by default in `Formatter` and record views; errors use structured format strings without embedding decrypted payloads.
- *Risk:* Concurrency or open database locks in multi-command tests.
  *Mitigation:* All command executions cleanly close SQLite DB handles and zeroize keys in `defer` statements.

## Assumptions

- Master password can be supplied interactively or via `GOVAULT_PASSWORD` environment variable for non-interactive automation.
- `RecordService` handles optimistic concurrency and history versioning.
