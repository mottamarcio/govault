package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
	"unicode"
)

const (
	UpperChars     = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	LowerChars     = "abcdefghijklmnopqrstuvwxyz"
	DigitChars     = "0123456789"
	SymbolChars    = "!@#$%^&*()-_=+[]{}|;:,.<>/?"
	AmbiguousChars = "1lI0Oo"
)

// PasswordOptions specifies configurable options for random password generation.
type PasswordOptions struct {
	Length           int
	IncludeUpper     bool
	IncludeLower     bool
	IncludeDigits    bool
	IncludeSymbols   bool
	ExcludeAmbiguous bool
}

// DefaultPasswordOptions returns standard secure defaults (length 20, all char sets enabled).
func DefaultPasswordOptions() PasswordOptions {
	return PasswordOptions{
		Length:           20,
		IncludeUpper:     true,
		IncludeLower:     true,
		IncludeDigits:    true,
		IncludeSymbols:   true,
		ExcludeAmbiguous: false,
	}
}

// StrengthLevel categorizes password strength.
type StrengthLevel string

const (
	StrengthVeryWeak   StrengthLevel = "Very Weak"
	StrengthWeak       StrengthLevel = "Weak"
	StrengthFair       StrengthLevel = "Fair"
	StrengthStrong     StrengthLevel = "Strong"
	StrengthVeryStrong StrengthLevel = "Very Strong"
)

// GeneratePassword produces a cryptographically secure random password adhering to PasswordOptions.
func GeneratePassword(opts PasswordOptions) (string, error) {
	if opts.Length < 8 {
		return "", errors.New("password length must be at least 8 characters")
	}
	if opts.Length > 128 {
		return "", errors.New("password length cannot exceed 128 characters")
	}

	var pools []string
	var guaranteedChars []byte
	var fullPool strings.Builder

	filter := func(set string) string {
		if !opts.ExcludeAmbiguous {
			return set
		}
		var b strings.Builder
		for _, r := range set {
			if !strings.ContainsRune(AmbiguousChars, r) {
				b.WriteRune(r)
			}
		}
		return b.String()
	}

	if opts.IncludeUpper {
		s := filter(UpperChars)
		if len(s) > 0 {
			pools = append(pools, s)
			fullPool.WriteString(s)
		}
	}
	if opts.IncludeLower {
		s := filter(LowerChars)
		if len(s) > 0 {
			pools = append(pools, s)
			fullPool.WriteString(s)
		}
	}
	if opts.IncludeDigits {
		s := filter(DigitChars)
		if len(s) > 0 {
			pools = append(pools, s)
			fullPool.WriteString(s)
		}
	}
	if opts.IncludeSymbols {
		s := filter(SymbolChars)
		if len(s) > 0 {
			pools = append(pools, s)
			fullPool.WriteString(s)
		}
	}

	if len(pools) == 0 {
		return "", errors.New("at least one character set must be selected")
	}
	if opts.Length < len(pools) {
		return "", fmt.Errorf("length %d is too short to include all %d selected character sets", opts.Length, len(pools))
	}

	// 1. Pick at least one guaranteed character from each chosen pool
	for _, p := range pools {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(p))))
		if err != nil {
			return "", fmt.Errorf("crypto rand failed: %w", err)
		}
		guaranteedChars = append(guaranteedChars, p[idx.Int64()])
	}

	// 2. Fill remaining characters from the combined pool
	combined := fullPool.String()
	combinedLen := big.NewInt(int64(len(combined)))
	remainingCount := opts.Length - len(guaranteedChars)

	result := make([]byte, opts.Length)
	copy(result, guaranteedChars)

	for i := 0; i < remainingCount; i++ {
		idx, err := rand.Int(rand.Reader, combinedLen)
		if err != nil {
			return "", fmt.Errorf("crypto rand failed: %w", err)
		}
		result[len(guaranteedChars)+i] = combined[idx.Int64()]
	}

	// 3. Fisher-Yates shuffle using crypto/rand
	for i := len(result) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", fmt.Errorf("crypto rand shuffle failed: %w", err)
		}
		result[i], result[j.Int64()] = result[j.Int64()], result[i]
	}

	return string(result), nil
}

// CalculateEntropy calculates the Shannon/information entropy in bits.
func CalculateEntropy(secret string) float64 {
	if len(secret) == 0 {
		return 0.0
	}

	// Check if passphrase (words separated by delimiters)
	words := strings.FieldsFunc(secret, func(r rune) bool {
		return r == '-' || r == ' ' || r == '_' || r == '.'
	})

	if len(words) >= 3 && isAllLettersOnly(words) {
		// Diceware entropy: L * log2(wordlist_size)
		return float64(len(words)) * math.Log2(float64(len(DefaultWordlist)))
	}

	var hasLower, hasUpper, hasDigits, hasSymbols bool
	for _, r := range secret {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigits = true
		default:
			hasSymbols = true
		}
	}

	poolSize := 0
	if hasLower {
		poolSize += 26
	}
	if hasUpper {
		poolSize += 26
	}
	if hasDigits {
		poolSize += 10
	}
	if hasSymbols {
		poolSize += 32
	}

	if poolSize == 0 {
		return 0.0
	}

	return float64(len(secret)) * math.Log2(float64(poolSize))
}

func isAllLettersOnly(words []string) bool {
	for _, w := range words {
		for _, r := range w {
			if !unicode.IsLetter(r) {
				return false
			}
		}
	}
	return true
}

// EvaluateStrength converts an entropy value into a qualitative strength level.
func EvaluateStrength(entropy float64) StrengthLevel {
	switch {
	case entropy < 28:
		return StrengthVeryWeak
	case entropy < 36:
		return StrengthWeak
	case entropy < 60:
		return StrengthFair
	case entropy < 128:
		return StrengthStrong
	default:
		return StrengthVeryStrong
	}
}
