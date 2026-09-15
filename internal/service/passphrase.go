package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// PassphraseOptions specifies options for generating a Diceware passphrase.
type PassphraseOptions struct {
	WordCount  int
	Delimiter  string
	Capitalize bool
	Wordlist   []string
}

// DefaultPassphraseOptions returns default options (5 words, hyphen delimiter).
func DefaultPassphraseOptions() PassphraseOptions {
	return PassphraseOptions{
		WordCount:  5,
		Delimiter:  "-",
		Capitalize: false,
		Wordlist:   DefaultWordlist,
	}
}

// GeneratePassphrase generates a cryptographically secure multi-word passphrase.
func GeneratePassphrase(opts PassphraseOptions) (string, error) {
	if opts.WordCount < 3 {
		return "", errors.New("word count must be at least 3")
	}
	if opts.WordCount > 12 {
		return "", errors.New("word count cannot exceed 12")
	}

	wordlist := opts.Wordlist
	if len(wordlist) == 0 {
		wordlist = DefaultWordlist
	}

	totalWords := big.NewInt(int64(len(wordlist)))
	selected := make([]string, opts.WordCount)

	for i := 0; i < opts.WordCount; i++ {
		idx, err := rand.Int(rand.Reader, totalWords)
		if err != nil {
			return "", fmt.Errorf("failed to generate random word index: %w", err)
		}
		word := wordlist[idx.Int64()]
		if opts.Capitalize && len(word) > 0 {
			word = strings.ToUpper(word[:1]) + word[1:]
		}
		selected[i] = word
	}

	return strings.Join(selected, opts.Delimiter), nil
}
