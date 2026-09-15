package service_test

import (
	"strings"
	"testing"

	"github.com/mottamarcio/govault/internal/service"
)

func TestPassphraseGenerator(t *testing.T) {
	t.Run("Default passphrase options generate 5 words separated by hyphen", func(t *testing.T) {
		opts := service.DefaultPassphraseOptions()
		pass, err := service.GeneratePassphrase(opts)
		if err != nil {
			t.Fatalf("GeneratePassphrase failed: %v", err)
		}

		parts := strings.Split(pass, "-")
		if len(parts) != 5 {
			t.Fatalf("expected 5 words, got %d in %q", len(parts), pass)
		}
		for _, w := range parts {
			if len(w) == 0 {
				t.Fatal("empty word encountered")
			}
		}
	})

	t.Run("Custom delimiter and word count", func(t *testing.T) {
		opts := service.PassphraseOptions{
			WordCount:  4,
			Delimiter:  " ",
			Capitalize: true,
		}
		pass, err := service.GeneratePassphrase(opts)
		if err != nil {
			t.Fatalf("GeneratePassphrase failed: %v", err)
		}

		parts := strings.Split(pass, " ")
		if len(parts) != 4 {
			t.Fatalf("expected 4 words, got %d in %q", len(parts), pass)
		}

		for _, w := range parts {
			firstChar := string(w[0])
			if firstChar != strings.ToUpper(firstChar) {
				t.Fatalf("expected capitalized word, got %s", w)
			}
		}
	})

	t.Run("WordCount out of bounds fails", func(t *testing.T) {
		_, err := service.GeneratePassphrase(service.PassphraseOptions{WordCount: 2})
		if err == nil {
			t.Fatal("expected error for WordCount < 3")
		}

		_, err = service.GeneratePassphrase(service.PassphraseOptions{WordCount: 15})
		if err == nil {
			t.Fatal("expected error for WordCount > 12")
		}
	})
}
