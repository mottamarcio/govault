package cli

import (
	"context"
	"fmt"

	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
)

// withRecordService is a helper that opens the database, prompts/authenticates the master password,
// creates a RecordService, executes the provided action, and ensures clean closure and zeroization.
func withRecordService(appCtx *AppContext, fn func(ctx context.Context, recordSvc *service.RecordService) error) error {
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
	password, err := prompter.ReadPassword("Enter master password")
	if err != nil {
		return err
	}

	session, err := vaultSvc.Unlock(ctx, password)
	if err != nil {
		return err
	}
	defer session.Lock()

	tagRepo := sqlite.NewTagRepository(db)
	recordRepo := sqlite.NewRecordRepository(db, tagRepo)
	historyRepo := sqlite.NewHistoryRepository(db)

	recordSvc := service.NewRecordService(session, recordRepo, tagRepo, historyRepo)
	return fn(ctx, recordSvc)
}
