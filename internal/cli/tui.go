package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/mottamarcio/govault/internal/storage/sqlite"
	"github.com/mottamarcio/govault/internal/tui"
)

func newTUICmd(appCtx *AppContext) *cobra.Command {
	var timeoutMinutes int

	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Launch interactive Terminal User Interface (TUI)",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			ctx := context.Background()
			if err := db.Migrate(ctx); err != nil {
				return fmt.Errorf("failed to apply migrations: %w", err)
			}

			metaRepo := sqlite.NewMetadataRepository(db)
			tagRepo := sqlite.NewTagRepository(db)
			recordRepo := sqlite.NewRecordRepository(db, tagRepo)
			historyRepo := sqlite.NewHistoryRepository(db)

			timeout := time.Duration(timeoutMinutes) * time.Minute
			if timeout <= 0 {
				timeout = 5 * time.Minute
			}

			app := tui.NewApp(db, metaRepo, recordRepo, tagRepo, historyRepo, appCtx.VaultPath, timeout)

			p := tea.NewProgram(
				app,
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)

			if _, err := p.Run(); err != nil {
				return fmt.Errorf("tui execution error: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&timeoutMinutes, "timeout", 5, "session inactivity auto-lock timeout in minutes")

	return cmd
}
