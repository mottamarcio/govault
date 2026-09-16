---
id: SPEC-014
type: spec
status: ready
parent: FEAT-005
depends_on:
  - SPEC-007
  - SPEC-009
  - SPEC-013
supersedes: []
---

# SPEC-014: CLI Secret Record Operations & In-Memory Search

## Intent

Specify the CLI commands for managing secret records (`add`, `get`, `list`, `edit`, `delete`, `trash`, `search`), supporting typed secret payloads, masked output by default, field filtering, and interactive/JSON output modes.

## Requirements

- **R1: Secret Record Creation (`govault add`):** Implement interactive and non-interactive command to create typed secret records:
  - Supported types: `login`, `note`, `apikey`, `custom` (specified via `--type` or prompt).
  - Flags for title, username, URI, tags (`--tag`), and custom key-value fields (`--field key=value`).
  - Secure masked password prompt with option to auto-generate a strong password during record creation.
- **R2: Secret Retrieval & Masking (`govault get` / `govault show`):** Retrieve record by ID, title, or alias:
  - By default, secret fields (passwords, API keys) are masked (e.g. `••••••••••••`) unless explicit `--show` flag is provided.
  - Support extracting specific single fields (e.g. `govault get <id> --field password --raw` for shell pipeline integration).
  - Support `--json` mode emitting complete decrypted record payload.
- **R3: Listing & In-Memory Search (`govault list` / `govault search`):**
  - `govault list`: Display tabular overview of records (ID, Type, Title, Tags, Updated) with optional filtering by `--type` and `--tag`.
  - `govault search <query>`: Execute fuzzy/prefix relevance ranking across decrypted titles and tags in memory.
- **R4: Modification, Deletion & Trash Management (`govault edit` / `govault delete` / `govault trash`):**
  - `govault edit <id>`: Modify existing record title, payload fields, or tags.
  - `govault delete <id>`: Soft-delete record to trash (with `--permanent` flag for immediate purging).
  - `govault trash list`, `govault trash restore <id>`, `govault trash purge`: Manage soft-deleted records.

## Acceptance Scenarios

- **Scenario 1: Add and Get Login Record**
  - *Given* an unlocked session
  - *When* `govault add --type login --title "Github" --tag dev` is executed with username "alice" and password "secret123"
  - *Then* the record is encrypted and saved, and `govault get "Github" --show` prints username and decrypted password.
- **Scenario 2: Single Field Extraction for Scripting**
  - *Given* a record with password "TopSecret!"
  - *When* `govault get <id> --field password --raw` is executed in a script
  - *Then* exactly `TopSecret!` is written to stdout with no trailing decoration and exit code 0.
- **Scenario 3: Search and Tag Filtering**
  - *Given* 5 records with various tags
  - *When* `govault search --tag dev --json` is executed
  - *Then* only records containing tag "dev" are returned as a JSON array.

## Edge Cases

- Getting or editing a non-existent record returns exit code 1 with error message to stderr.
- Extracting a non-existent field via `--field` returns an informative error without displaying other sensitive fields.
- Soft-deleted records are hidden from default `list` and `search` outputs unless `--trash` flag is provided.

## Constraints

- Secret values must never be logged or echoed in unmasked plaintext unless explicitly requested via `--show` or `--raw`.
- Tabular outputs must use standard POSIX formatting compatible with `awk`, `cut`, or interactive terminals.

## Non-Goals

- Full-screen TUI text editor widgets (delegated to FEAT-006).
