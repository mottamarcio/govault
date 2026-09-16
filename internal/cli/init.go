package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

func newInitCmd(appCtx *AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new secure GoVault database",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			// Ensure directory exists with 0700
			dir := filepath.Dir(appCtx.VaultPath)
			if err := os.MkdirAll(dir, 0700); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}

			db, err := sqlite.Open(appCtx.VaultPath)
			if err != nil {
				return fmt.Errorf("failed to open vault database: %w", err)
			}
			defer db.Close()

			if err := db.Migrate(ctx); err != nil {
				return fmt.Errorf("failed to apply migrations: %w", err)
			}

			metaRepo := sqlite.NewMetadataRepository(db)
			vaultSvc := service.NewVaultService(db, metaRepo, appCtx.VaultPath)

			prompter := DefaultPrompter(appCtx.In, appCtx.Err)
			masterPassword, err := prompter.ReadPasswordConfirm("Enter new master password")
			if err != nil {
				return err
			}

			if err := vaultSvc.Init(ctx, masterPassword); err != nil {
				return err
			}

			if appCtx.JSON {
				header, _ := vaultSvc.Inspect(ctx)
				return appCtx.Formatter.PrintJSON(map[string]any{
					"status":     "initialized",
					"vault_id":   header.VaultID,
					"vault_path": appCtx.VaultPath,
				})
			}

			appCtx.Formatter.PrintText("✓ Vault successfully initialized at %s\n", appCtx.VaultPath)
			return nil
		},
	}
	return cmd
}
