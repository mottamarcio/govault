---
type: tasks
for: SPEC-010
---

# Tasks

## TASK-034 — Embedded Wordlist and EFF Diceware Passphrase Generator

- [x] Completed
- **Serves:** SPEC-010:R2
- **Depends on:** none
- **Files/Components:** `internal/service/wordlist.go`, `internal/service/passphrase.go`, `internal/service/passphrase_test.go`
- **Verification:** Ran `go test -v -run TestPassphraseGenerator ./internal/service` asserting word count, custom delimiters, capitalization, and cryptographically secure word selection. Evidence: `PASS TestPassphraseGenerator`.

## TASK-035 — Configurable Password Generator and Shannon Entropy Calculator

- [x] Completed
- **Serves:** SPEC-010:R1, SPEC-010:R3
- **Depends on:** none
- **Files/Components:** `internal/service/password_gen.go`, `internal/service/password_gen_test.go`
- **Verification:** Ran `go test -v -run TestPasswordGenerator ./internal/service` asserting character set inclusions, ambiguous character exclusions, Fisher-Yates shuffling, entropy calculations, and strength level evaluations. Evidence: `PASS TestPasswordGenerator`.

## TASK-036 — Secure Timed Clipboard Service

- [x] Completed
- **Serves:** SPEC-010:R4
- **Depends on:** none
- **Files/Components:** `internal/service/clipboard.go`, `internal/service/clipboard_test.go`
- **Verification:** Ran `go test -v -race -run TestClipboardService ./internal/service` validating timed clipboard clear, timer cancellation upon new copy, and immediate manual wipe. Evidence: `PASS TestClipboardService`.
