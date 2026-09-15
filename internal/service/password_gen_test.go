package service_test

import (
	"strings"
	"testing"

	"github.com/mottamarcio/govault/internal/service"
)

func TestPasswordGenerator(t *testing.T) {
	t.Run("Default password options generate length 20 with all sets", func(t *testing.T) {
		opts := service.DefaultPasswordOptions()
		pass, err := service.GeneratePassword(opts)
		if err != nil {
			t.Fatalf("GeneratePassword failed: %v", err)
		}
		if len(pass) != 20 {
			t.Fatalf("expected length 20, got %d", len(pass))
		}

		entropy := service.CalculateEntropy(pass)
		if entropy < 120 {
			t.Fatalf("expected entropy >= 120, got %f", entropy)
		}

		strength := service.EvaluateStrength(entropy)
		if strength != service.StrengthStrong && strength != service.StrengthVeryStrong {
			t.Fatalf("unexpected strength: %s", strength)
		}
	})

	t.Run("Exclude ambiguous characters", func(t *testing.T) {
		opts := service.PasswordOptions{
			Length:           30,
			IncludeUpper:     true,
			IncludeLower:     true,
			IncludeDigits:    true,
			IncludeSymbols:   false,
			ExcludeAmbiguous: true,
		}

		for i := 0; i < 20; i++ {
			pass, err := service.GeneratePassword(opts)
			if err != nil {
				t.Fatalf("GeneratePassword failed: %v", err)
			}
			for _, r := range pass {
				if strings.ContainsRune("1lI0Oo", r) {
					t.Fatalf("ambiguous char %c found in %s", r, pass)
				}
			}
		}
	})

	t.Run("Entropy and Strength table tests", func(t *testing.T) {
		cases := []struct {
			secret   string
			expected service.StrengthLevel
		}{
			{"1234", service.StrengthVeryWeak},
			{"abcdefgh", service.StrengthFair},
			{"correct-horse-battery-staple", service.StrengthFair},
			{"k9#mP2$xL1@zQ8!v", service.StrengthStrong},
			{"k9#mP2$xL1@zQ8!vA1b2C3d4E5f6G7h8", service.StrengthVeryStrong},
		}

		for _, tc := range cases {
			ent := service.CalculateEntropy(tc.secret)
			str := service.EvaluateStrength(ent)
			if str != tc.expected {
				t.Errorf("secret %q: entropy=%f, got strength %s, want %s", tc.secret, ent, str, tc.expected)
			}
		}
	})

	t.Run("Invalid options fail", func(t *testing.T) {
		_, err := service.GeneratePassword(service.PasswordOptions{Length: 5})
		if err == nil {
			t.Fatal("expected error for Length < 8")
		}

		_, err = service.GeneratePassword(service.PasswordOptions{Length: 200})
		if err == nil {
			t.Fatal("expected error for Length > 128")
		}

		_, err = service.GeneratePassword(service.PasswordOptions{
			Length: 10,
			// No character sets selected
		})
		if err == nil {
			t.Fatal("expected error when no character sets selected")
		}
	})
}
