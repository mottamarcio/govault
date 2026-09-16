package cli

import (
	"context"
	"fmt"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

func newUnlockCmd(appCtx *AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "Unlock the vault and verify master password",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			db, err := sqlite.Open(appCtx.VaultPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			metaRepo := sqlite.NewMetadataRepository(db)
			vaultSvc := service.NewVaultService(db, metaRepo, appCtx.VaultPath)

			prompter := DefaultPrompter(appCtx.In, appCtx.Err)
			password, err := prompter.ReadPassword("Enter master password")
			if err != nil {
				return err
			}

			session, err := vaultSvc.Unlock(ctx, password)
			if err != nil {
				return err
			}

			vaultID, _ := session.VaultID()
			if appCtx.JSON {
				return appCtx.Formatter.PrintJSON(map[string]any{
					"status":   "unlocked",
					"vault_id": vaultID,
				})
			}

			appCtx.Formatter.PrintText("✓ Vault successfully authenticated and unlocked\n")
			return nil
		},
	}
	return cmd
}

func newLockCmd(appCtx *AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Lock the vault session and wipe memory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if appCtx.JSON {
				return appCtx.Formatter.PrintJSON(map[string]any{
					"status": "locked",
				})
			}
			appCtx.Formatter.PrintText("✓ Vault is locked\n")
			return nil
		},
	}
	return cmd
}
