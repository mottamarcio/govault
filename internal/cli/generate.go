package cli

import (
	"fmt"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/spf13/cobra"
)

func newGenerateCmd(appCtx *AppContext) *cobra.Command {
	var (
		length          int
		noUpper         bool
		noDigits        bool
		noSymbols       bool
		noAmbiguous     bool
		passphraseMode  bool
		wordCount       int
		delimiter       string
		capitalizeWords bool
		rawOutput       bool
	)

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate cryptographically secure passwords or Diceware passphrases",
		RunE: func(cmd *cobra.Command, args []string) error {
			if passphraseMode {
				// Diceware Passphrase mode
				if wordCount < 3 || wordCount > 12 {
					return fmt.Errorf("word count must be between 3 and 12 (got %d)", wordCount)
				}

				opts := service.PassphraseOptions{
					WordCount:  wordCount,
					Delimiter:  delimiter,
					Capitalize: capitalizeWords,
					Wordlist:   service.DefaultWordlist,
				}

				passphrase, err := service.GeneratePassphrase(opts)
				if err != nil {
					return fmt.Errorf("failed to generate passphrase: %w", err)
				}

				entropy := service.CalculateEntropy(passphrase)
				strength := service.EvaluateStrength(entropy)

				if rawOutput {
					fmt.Fprint(appCtx.Out, passphrase)
					return nil
				}

				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(map[string]any{
						"type":       "passphrase",
						"secret":     passphrase,
						"word_count": wordCount,
						"delimiter":  delimiter,
						"entropy":    entropy,
						"strength":   string(strength),
					})
				}

				appCtx.Formatter.PrintText("%s\n\n", passphrase)
				appCtx.Formatter.PrintText("Entropy:   %.1f bits\n", entropy)
				appCtx.Formatter.PrintText("Strength:  %s\n", strength)
				return nil
			}

			// Standard random password mode
			if length < 8 || length > 128 {
				return fmt.Errorf("password length must be between 8 and 128 characters (got %d)", length)
			}

			opts := service.PasswordOptions{
				Length:           length,
				IncludeUpper:     !noUpper,
				IncludeLower:     true,
				IncludeDigits:    !noDigits,
				IncludeSymbols:   !noSymbols,
				ExcludeAmbiguous: noAmbiguous,
			}

			password, err := service.GeneratePassword(opts)
			if err != nil {
				return fmt.Errorf("failed to generate password: %w", err)
			}

			entropy := service.CalculateEntropy(password)
			strength := service.EvaluateStrength(entropy)

			if rawOutput {
				fmt.Fprint(appCtx.Out, password)
				return nil
			}

			if appCtx.JSON {
				return appCtx.Formatter.PrintJSON(map[string]any{
					"type":     "password",
					"secret":   password,
					"length":   length,
					"entropy":  entropy,
					"strength": string(strength),
				})
			}

			appCtx.Formatter.PrintText("%s\n\n", password)
			appCtx.Formatter.PrintText("Length:    %d\n", length)
			appCtx.Formatter.PrintText("Entropy:   %.1f bits\n", entropy)
			appCtx.Formatter.PrintText("Strength:  %s\n", strength)
			return nil
		},
	}

	// Password flags
	cmd.Flags().IntVarP(&length, "length", "l", 20, "length of generated password (8-128)")
	cmd.Flags().BoolVar(&noUpper, "no-upper", false, "exclude uppercase letters (A-Z)")
	cmd.Flags().BoolVar(&noDigits, "no-digits", false, "exclude numeric digits (0-9)")
	cmd.Flags().BoolVar(&noSymbols, "no-symbols", false, "exclude special symbols")
	cmd.Flags().BoolVar(&noAmbiguous, "no-ambiguous", false, "exclude ambiguous characters (1lI0Oo)")

	// Passphrase flags
	cmd.Flags().BoolVarP(&passphraseMode, "passphrase", "p", false, "generate Diceware word-based passphrase")
	cmd.Flags().IntVarP(&wordCount, "words", "w", 5, "number of words in passphrase (3-12)")
	cmd.Flags().StringVarP(&delimiter, "delimiter", "d", "-", "delimiter between words")
	cmd.Flags().BoolVarP(&capitalizeWords, "capitalize", "c", false, "capitalize first letter of each word")

	// Output flags
	cmd.Flags().BoolVarP(&rawOutput, "raw", "r", false, "print raw secret string with no labels or formatting")

	return cmd
}
