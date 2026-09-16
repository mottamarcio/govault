package cli

import (
	"context"
	"fmt"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

func newPasswdCmd(appCtx *AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "passwd",
		Short: "Change the vault master password",
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

			currentPassword, err := prompter.ReadPassword("Enter current master password")
			if err != nil {
				return err
			}

			newPassword, err := prompter.ReadPasswordConfirm("Enter new master password")
			if err != nil {
				return err
			}

			if err := vaultSvc.ChangeMasterPassword(ctx, currentPassword, newPassword); err != nil {
				return err
			}

			if appCtx.JSON {
				return appCtx.Formatter.PrintJSON(map[string]any{
					"status":  "password_changed",
					"success": true,
				})
			}

			appCtx.Formatter.PrintText("✓ Master password changed successfully\n")
			return nil
		},
	}
	return cmd
}
