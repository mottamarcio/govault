package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/spf13/cobra"
)

const (
	Version     = "1.0.0"
	CryptoSuite = "argon2id-hkdf-xchacha20poly1305 (v1)"
)

// AppContext holds shared runtime context and configuration across commands.
type AppContext struct {
	VaultPath string
	JSON      bool
	Quiet     bool
	In        io.Reader
	Out       io.Writer
	Err       io.Writer
	Formatter *OutputFormatter
	Prompter  *PasswordPrompter
	Clipboard service.ClipboardDriver
}

// DefaultVaultPath returns the standard default vault location.
func DefaultVaultPath() string {
	if envPath := os.Getenv("GOVAULT_PATH"); envPath != "" {
		return envPath
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "vault.db"
	}
	return filepath.Join(home, ".govault", "vault.db")
}

// NewRootCommand creates the root `govault` Cobra command tree.
func NewRootCommand(appCtx *AppContext) *cobra.Command {
	if appCtx.In == nil {
		appCtx.In = os.Stdin
	}
	if appCtx.Out == nil {
		appCtx.Out = os.Stdout
	}
	if appCtx.Err == nil {
		appCtx.Err = os.Stderr
	}

	rootCmd := &cobra.Command{
		Use:           "govault",
		Short:         "GoVault: Modern offline password and secrets manager",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if appCtx.VaultPath == "" {
				appCtx.VaultPath = DefaultVaultPath()
			}
			appCtx.Formatter = NewOutputFormatter(appCtx.Out, appCtx.Err, appCtx.JSON, appCtx.Quiet)
			if appCtx.Prompter == nil {
				appCtx.Prompter = DefaultPrompter(appCtx.In, appCtx.Err)
			}
			return nil
		},
	}

	defaultPath := appCtx.VaultPath
	if defaultPath == "" {
		defaultPath = DefaultVaultPath()
	}

	// Persistent global flags
	rootCmd.PersistentFlags().StringVar(&appCtx.VaultPath, "vault", defaultPath, "path to vault database file (default: ~/.govault/vault.db)")
	rootCmd.PersistentFlags().BoolVar(&appCtx.JSON, "json", appCtx.JSON, "output results in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&appCtx.Quiet, "quiet", "q", appCtx.Quiet, "suppress informational stdout messages")

	// Version sub-command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print GoVault version and cryptographic suite",
		RunE: func(cmd *cobra.Command, args []string) error {
			if appCtx.JSON {
				return appCtx.Formatter.PrintJSON(map[string]string{
					"version":      Version,
					"crypto_suite": CryptoSuite,
				})
			}
			appCtx.Formatter.PrintText("GoVault v%s\nCrypto Suite: %s\n", Version, CryptoSuite)
			return nil
		},
	}

	rootCmd.AddCommand(versionCmd)

	// Attach lifecycle commands
	rootCmd.AddCommand(newInitCmd(appCtx))
	rootCmd.AddCommand(newStatusCmd(appCtx))
	rootCmd.AddCommand(newUnlockCmd(appCtx))
	rootCmd.AddCommand(newLockCmd(appCtx))
	rootCmd.AddCommand(newPasswdCmd(appCtx))

	// Attach secret record commands
	rootCmd.AddCommand(newAddCmd(appCtx))
	rootCmd.AddCommand(newGetCmd(appCtx))
	rootCmd.AddCommand(newListCmd(appCtx))
	rootCmd.AddCommand(newSearchCmd(appCtx))
	rootCmd.AddCommand(newEditCmd(appCtx))
	rootCmd.AddCommand(newDeleteCmd(appCtx))
	rootCmd.AddCommand(newTrashCmd(appCtx))

	// Attach utility and maintenance commands
	rootCmd.AddCommand(newGenerateCmd(appCtx))
	rootCmd.AddCommand(newCopyCmd(appCtx))
	rootCmd.AddCommand(newBackupCmd(appCtx))
	rootCmd.AddCommand(newRestoreCmd(appCtx))
	rootCmd.AddCommand(newInspectCmd(appCtx))
	rootCmd.AddCommand(newDoctorCmd(appCtx))
	rootCmd.AddCommand(newTUICmd(appCtx))

	return rootCmd
}

// Execute runs the root CLI command and exits with standard POSIX exit codes.
func Execute() {
	appCtx := &AppContext{}
	rootCmd := NewRootCommand(appCtx)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
