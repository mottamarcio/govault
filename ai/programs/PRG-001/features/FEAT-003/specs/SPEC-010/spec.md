---
id: SPEC-010
type: spec
status: ready
parent: FEAT-003
depends_on:
  - SPEC-001
  - SPEC-007
---

# SPEC-010: Password Generator & Secure Clipboard Service

## Intent

Specify the `PasswordGenService` (configurable random password generator, EFF Diceware passphrase generator, entropy calculation) and `ClipboardService` (safe timed clipboard copy and automatic zero/clear mechanism).

## Requirements

- **R1: Configurable Random Password Generator:** Generate cryptographically secure random passwords based on configurable options: length (default 20, min 8, max 128), character sets (uppercase, lowercase, digits, symbols), and option to exclude ambiguous characters (`1, l, I, 0, O, o`).
- **R2: Diceware Passphrase Generator:** Generate multi-word passphrases using cryptographically secure random selection from an embedded EFF wordlist, with configurable word count (default 5, min 3, max 12) and delimiter (e.g. `-`, ` `, `.`).
- **R3: Password Entropy Calculation:** Calculate Shannon/information entropy (in bits) and strength estimation (Very Weak, Weak, Fair, Strong, Very Strong) for generated and user-provided passwords/passphrases.
- **R4: Secure Clipboard Service:** Provide clipboard copy functionality that automatically clears the clipboard after a configurable duration (default 45 seconds) using a timer/goroutine, with support for immediate manual clearing.

## Acceptance Scenarios

- **Scenario 1: Secure Password Generation**
  - *Given* password options with length 24 and all character sets enabled
  - *When* `GeneratePassword` is called
  - *Then* a 24-character string containing characters from all requested sets is returned with entropy >= 120 bits.
- **Scenario 2: Diceware Passphrase Generation**
  - *Given* passphrase options with 5 words and `-` delimiter
  - *When* `GeneratePassphrase` is called
  - *Then* a string with 5 words joined by `-` is generated with high entropy.
- **Scenario 3: Timed Clipboard Clear**
  - *Given* a sensitive string copied to clipboard with 2-second timeout
  - *When* 2 seconds elapse
  - *Then* the clipboard content is cleared.

## Edge Cases

- Generating password with length shorter than enabled character set count returns an error.
- Clipboard clear timer must cancel and not overwrite clipboard if user copied new content in between.

## Constraints

- Password generation must use `crypto/rand` only; math/rand is strictly forbidden.
- Zero network dependencies.

## Non-Goals

- GUI or TUI widget implementations.
