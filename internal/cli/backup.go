package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mottamarcio/govault/internal/backup/format"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

func newBackupCmd(appCtx *AppContext) *cobra.Command {
	var (
		exportPassphrase string
		overwrite        bool
	)

	cmd := &cobra.Command{
		Use:   "backup <output.gvault>",
		Short: "Export the vault to an encrypted portable .gvault backup file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			outputPath := strings.TrimSpace(args[0])
			if outputPath == "" {
				return fmt.Errorf("backup output path is required")
			}

			if _, err := os.Stat(outputPath); err == nil && !overwrite {
				return fmt.Errorf("destination file '%s' already exists (use --overwrite to replace)", outputPath)
			}

			ctx := context.Background()
			db, err := sqlite.Open(appCtx.VaultPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			metaRepo := sqlite.NewMetadataRepository(db)
			vaultSvc := service.NewVaultService(db, metaRepo, appCtx.VaultPath)

			prompter := appCtx.Prompter
			if prompter == nil {
				prompter = DefaultPrompter(appCtx.In, appCtx.Err)
			}

			password, err := prompter.ReadPassword("Enter master password to authorize backup")
			if err != nil {
				return err
			}

			sess, err := vaultSvc.Unlock(ctx, password)
			if err != nil {
				return err
			}
			defer sess.Lock()

			tagRepo := sqlite.NewTagRepository(db)
			recordRepo := sqlite.NewRecordRepository(db, tagRepo)
			historyRepo := sqlite.NewHistoryRepository(db)

			backupSvc := service.NewBackupService(metaRepo, recordRepo, tagRepo, historyRepo, appCtx.VaultPath)

			opts := service.BackupCreateOptions{
				ExportPassphrase: exportPassphrase,
				Overwrite:        overwrite,
			}

			if err := backupSvc.CreateBackup(ctx, sess, outputPath, opts); err != nil {
				return fmt.Errorf("backup creation failed: %w", err)
			}

			if appCtx.JSON {
				return appCtx.Formatter.PrintJSON(map[string]any{
					"status":      "backup_created",
					"output_path": outputPath,
				})
			}

			appCtx.Formatter.PrintText("✓ Vault successfully backed up to '%s'\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&exportPassphrase, "passphrase", "p", "", "custom passphrase to encrypt the backup (defaults to master password)")
	cmd.Flags().BoolVarP(&overwrite, "overwrite", "o", false, "overwrite output file if it already exists")

	return cmd
}

func newRestoreCmd(appCtx *AppContext) *cobra.Command {
	var (
		skipSnapshot bool
	)

	cmd := &cobra.Command{
		Use:   "restore <input.gvault>",
		Short: "Restore an encrypted .gvault backup into the active vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputPath := strings.TrimSpace(args[0])
			if inputPath == "" {
				return fmt.Errorf("backup input path is required")
			}

			if _, err := os.Stat(inputPath); err != nil {
				return fmt.Errorf("backup file '%s' not found: %w", inputPath, err)
			}

			ctx := context.Background()

			prompter := appCtx.Prompter
			if prompter == nil {
				prompter = DefaultPrompter(appCtx.In, appCtx.Err)
			}

			password, err := prompter.ReadPassword("Enter backup encryption passphrase")
			if err != nil {
				return err
			}

			backupSvc := service.NewBackupService(nil, nil, nil, nil, appCtx.VaultPath)
			opts := service.BackupRestoreOptions{
				SkipSnapshot: skipSnapshot,
			}

			if err := backupSvc.RestoreBackup(ctx, inputPath, password, opts); err != nil {
				return fmt.Errorf("restore failed: %w", err)
			}

			if appCtx.JSON {
				return appCtx.Formatter.PrintJSON(map[string]any{
					"status":     "restored",
					"input_path": inputPath,
					"vault_path": appCtx.VaultPath,
				})
			}

			appCtx.Formatter.PrintText("✓ Successfully restored vault from '%s'\n", inputPath)
			return nil
		},
	}

	cmd.Flags().BoolVar(&skipSnapshot, "skip-snapshot", false, "skip creating automatic recovery snapshot before restoring")

	return cmd
}

func newInspectCmd(appCtx *AppContext) *cobra.Command {
	var (
		authenticated bool
	)

	cmd := &cobra.Command{
		Use:   "inspect <backup.gvault>",
		Short: "Inspect metadata and headers of a .gvault backup file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			backupPath := strings.TrimSpace(args[0])
			if backupPath == "" {
				return fmt.Errorf("backup file path is required")
			}

			file, err := os.Open(backupPath)
			if err != nil {
				return fmt.Errorf("failed to open backup file: %w", err)
			}
			defer file.Close()

			inspection, err := format.InspectHeader(file)
			if err != nil {
				return fmt.Errorf("invalid backup header: %w", err)
			}

			var manifest *format.BackupManifest
			if authenticated {
				prompter := appCtx.Prompter
				if prompter == nil {
					prompter = DefaultPrompter(appCtx.In, appCtx.Err)
				}
				password, err := prompter.ReadPassword("Enter backup encryption passphrase for manifest inspection")
				if err != nil {
					return err
				}

				backupSvc := service.NewBackupService(nil, nil, nil, nil, "")
				verifyRes, err := backupSvc.VerifyBackup(context.Background(), backupPath, password)
				if err != nil {
					return fmt.Errorf("failed to decrypt and verify backup: %w", err)
				}
				manifest = verifyRes.Manifest
			}

			if appCtx.JSON {
				res := map[string]any{
					"inspection": inspection,
				}
				if manifest != nil {
					res["manifest"] = manifest
				}
				return appCtx.Formatter.PrintJSON(res)
			}

			appCtx.Formatter.PrintText("Backup ID:       %s\n", inspection.BackupID)
			appCtx.Formatter.PrintText("Vault ID:        %s\n", inspection.VaultID)
			appCtx.Formatter.PrintText("Format Version:  %d\n", inspection.FormatVersion)
			appCtx.Formatter.PrintText("Crypto Suite:    v%d\n", inspection.CryptoSuiteVersion)
			appCtx.Formatter.PrintText("Created At:      %s\n", inspection.CreatedAt.Format("2006-01-02 15:04:05"))
			appCtx.Formatter.PrintText("KDF Algorithm:   %d\n", inspection.KDFAlgorithm)
			appCtx.Formatter.PrintText("Argon2 Time:     %d\n", inspection.KDFTime)
			appCtx.Formatter.PrintText("Argon2 Memory:   %d MB\n", inspection.KDFMemoryMB)
			appCtx.Formatter.PrintText("Argon2 Threads:  %d\n", inspection.KDFThreads)

			if manifest != nil {
				appCtx.Formatter.PrintText("----------------------------------------\n")
				appCtx.Formatter.PrintText("App Version:     %s\n", manifest.AppVersion)
				appCtx.Formatter.PrintText("Entries Count:   %d\n", manifest.EntryCount)
				appCtx.Formatter.PrintText("Histories Count: %d\n", manifest.HistoryCount)
				appCtx.Formatter.PrintText("Tags Count:      %d\n", manifest.TagCount)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&authenticated, "authenticated", "a", false, "prompt for passphrase and decrypt authenticated manifest stats")

	return cmd
}
