---
type: plan
for: SPEC-010
status: ready
---

# Implementation Plan: Password Generator & Secure Clipboard Service

## Summary

Design and implement the `PasswordGenerator` (cryptographically secure character-based password generation, embedded EFF Diceware passphrase generation, Shannon entropy calculation, and password strength evaluation) and `ClipboardService` (in-memory clipboard interface, timed auto-clearing, and cancellation) under `internal/service`.

## Repository Context

- Cryptography: `crypto/rand` standard library for cryptographically secure random integers (`crypto/rand.Int`, `crypto/rand.Read`).
- Architecture: `internal/service` provides business logic services consumed by CLI and TUI.
- Zero network dependencies: strictly offline, embedded wordlist.

## Requirement Coverage

- **R1 → Configurable Random Password Generator:**
  - Define `PasswordOptions` struct: `Length` (min 8, max 128, default 20), `IncludeUpper` (bool), `IncludeLower` (bool), `IncludeDigits` (bool), `IncludeSymbols` (bool), `ExcludeAmbiguous` (bool: exclude `1, l, I, 0, O, o`).
  - Implement `GeneratePassword(opts PasswordOptions) (string, error)`:
    - Validate length >= number of enabled character sets.
    - Ensure at least one character from each selected set is included.
    - Fill remaining length from combined character pool using `crypto/rand.Int`.
    - Securely Fisher-Yates shuffle the generated character slice using `crypto/rand`.
- **R2 → Diceware Passphrase Generator:**
  - Embed curated EFF wordlist (or generate 7,776-word standardized list).
  - Define `PassphraseOptions` struct: `WordCount` (min 3, max 12, default 5), `Delimiter` (string, default `-`), `Capitalize` (bool).
  - Implement `GeneratePassphrase(opts PassphraseOptions) (string, error)`:
    - Select random words from wordlist using `crypto/rand.Int`.
    - Join words using specified delimiter.
- **R3 → Password Entropy Calculation:**
  - Implement `CalculateEntropy(secret string) float64`:
    - Determine character pool size $R$ based on presence of lowercase, uppercase, digits, symbols, or wordlist size for passphrases.
    - Calculate Shannon entropy: $E = L \times \log_2(R)$.
  - Implement `EvaluateStrength(entropy float64) StrengthLevel`:
    - Returns `VeryWeak` (< 28 bits), `Weak` (28-35 bits), `Fair` (36-59 bits), `Strong` (60-127 bits), `VeryStrong` (>= 128 bits).
- **R4 → Secure Clipboard Service:**
  - Define `ClipboardDriver` interface for platform clipboard access (`WriteText(string) error`, `ReadText() (string, error)`, `Clear() error`).
  - Implement `ClipboardService`:
    - `Copy(ctx context.Context, text string, timeout time.Duration) error`:
      - Write text to clipboard.
      - Cancel any existing clear timer.
      - Launch background timer to clear clipboard only if content still matches written text.
    - `Clear() error`: Immediate clipboard clearing.

## Architecture

```
internal/service/
├── password_gen.go       # Password generator, Diceware passphrases, and entropy
├── wordlist.go           # Embedded EFF wordlist
├── clipboard.go          # Timed clipboard manager and driver interface
├── password_gen_test.go  # Generator and entropy unit tests
└── clipboard_test.go     # Clipboard service tests
```

## Components Affected

- `internal/service/password_gen.go`
- `internal/service/wordlist.go`
- `internal/service/clipboard.go`
- `internal/service/password_gen_test.go`
- `internal/service/clipboard_test.go`

## Data Changes

- None.

## API Changes

```go
package service

type PasswordOptions struct {
    Length           int
    IncludeUpper     bool
    IncludeLower     bool
    IncludeDigits    bool
    IncludeSymbols   bool
    ExcludeAmbiguous bool
}

type PassphraseOptions struct {
    WordCount  int
    Delimiter  string
    Capitalize bool
}

type StrengthLevel string

const (
    StrengthVeryWeak   StrengthLevel = "Very Weak"
    StrengthWeak       StrengthLevel = "Weak"
    StrengthFair       StrengthLevel = "Fair"
    StrengthStrong     StrengthLevel = "Strong"
    StrengthVeryStrong StrengthLevel = "Very Strong"
)

func GeneratePassword(opts PasswordOptions) (string, error)
func GeneratePassphrase(opts PassphraseOptions) (string, error)
func CalculateEntropy(password string) float64
func EvaluateStrength(entropy float64) StrengthLevel

type ClipboardDriver interface {
    WriteText(text string) error
    ReadText() (string, error)
    Clear() error
}

type ClipboardService struct {
    driver ClipboardDriver
    timer  *time.Timer
    mu     sync.Mutex
}

func NewClipboardService(driver ClipboardDriver) *ClipboardService
func (c *ClipboardService) Copy(ctx context.Context, text string, timeout time.Duration) error
func (c *ClipboardService) Clear() error
```

## Integration Changes

- Ingested by CLI commands (`govault gen`, `govault get --clip`) and Bubble Tea TUI components.

## Implementation Sequence

1. Implement embedded wordlist (`wordlist.go`).
2. Implement password generator, passphrase generator, and entropy calculator (`password_gen.go`).
3. Implement `ClipboardService` with test driver (`clipboard.go`).
4. Write thorough unit tests covering password entropy, options, character set guarantees, Diceware randomness, and clipboard timeout clearing.

## Test Strategy

- **Character Set Test:** Assert generated passwords strictly adhere to options (min/max length, uppercase, digits, symbols, ambiguous exclusion).
- **Passphrase Test:** Assert word counts, delimiters, and capitalization.
- **Entropy Tests:** Table-driven tests validating entropy calculation against known password classes.
- **Clipboard Timed Clear Test:** Copy text with short timeout (e.g. 50ms), wait, verify clipboard is cleared; test timer cancellation on overwrite.

## Risks

- Platform clipboard dependencies on headless Linux/CI; mitigated by abstracting platform clipboard behind `ClipboardDriver` interface with an in-memory driver for unit tests.

## Assumptions

- Standard `crypto/rand` is available on all target OS environments.
