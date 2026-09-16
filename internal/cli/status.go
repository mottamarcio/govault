package cli

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
	"github.com/spf13/cobra"
)

func newStatusCmd(appCtx *AppContext) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Display vault initialization and lock status",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if _, err := os.Stat(appCtx.VaultPath); os.IsNotExist(err) {
				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(map[string]any{
						"status":      "uninitialized",
						"vault_path":  appCtx.VaultPath,
						"initialized": false,
					})
				}
				appCtx.Formatter.PrintText("Status: Uninitialized (no database at %s)\n", appCtx.VaultPath)
				return nil
			}

			db, err := sqlite.Open(appCtx.VaultPath)
			if err != nil {
				return fmt.Errorf("failed to open database: %w", err)
			}
			defer db.Close()

			metaRepo := sqlite.NewMetadataRepository(db)
			vaultSvc := service.NewVaultService(db, metaRepo, appCtx.VaultPath)

			header, err := vaultSvc.Inspect(ctx)
			if errors.Is(err, service.ErrVaultNotInitialized) {
				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(map[string]any{
						"status":      "uninitialized",
						"vault_path":  appCtx.VaultPath,
						"initialized": false,
					})
				}
				appCtx.Formatter.PrintText("Status: Uninitialized\n")
				return nil
			}
			if err != nil {
				return err
			}

			if appCtx.JSON {
				return appCtx.Formatter.PrintJSON(map[string]any{
					"status":         "locked",
					"initialized":    true,
					"vault_id":       header.VaultID,
					"schema_version": header.SchemaVersion,
					"crypto_suite":   header.CryptoSuite,
					"vault_path":     appCtx.VaultPath,
					"created_at":     header.CreatedAt.Format("2006-01-02 15:04:05 MST"),
				})
			}

			appCtx.Formatter.PrintText("Vault ID:        %s\n", header.VaultID)
			appCtx.Formatter.PrintText("Status:          Locked\n")
			appCtx.Formatter.PrintText("Schema Version:  %d\n", header.SchemaVersion)
			appCtx.Formatter.PrintText("Crypto Suite:    %s\n", header.CryptoSuite)
			appCtx.Formatter.PrintText("Database Path:   %s\n", appCtx.VaultPath)
			appCtx.Formatter.PrintText("Created At:      %s\n", header.CreatedAt.Format("2006-01-02 15:04:05 MST"))
			return nil
		},
	}
	return cmd
}
