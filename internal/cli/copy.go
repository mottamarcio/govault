package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/spf13/cobra"
)

func newCopyCmd(appCtx *AppContext) *cobra.Command {
	var (
		fieldName  string
		clearAfter time.Duration
	)

	cmd := &cobra.Command{
		Use:     "copy <id|title>",
		Aliases: []string{"clip"},
		Short:   "Copy a secret field to system clipboard with auto-clear timer",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return fmt.Errorf("record ID or title is required")
			}

			if clearAfter <= 0 {
				clearAfter = 45 * time.Second
			}

			return withRecordService(appCtx, func(ctx context.Context, recordSvc *service.RecordService) error {
				rec, err := recordSvc.GetByID(ctx, query)
				if err != nil {
					records, listErr := recordSvc.List(ctx, false)
					if listErr != nil {
						return fmt.Errorf("record '%s' not found: %w", query, err)
					}
					var found *domain.Record
					for _, r := range records {
						if strings.EqualFold(r.Title, query) || r.ID == query {
							found = r
							break
						}
					}
					if found == nil {
						return fmt.Errorf("record '%s' not found", query)
					}
					rec = found
				}

				// If no specific field requested, default to main sensitive field (password/secret)
				targetField := fieldName
				if targetField == "" {
					switch rec.Type {
					case domain.RecordTypeLogin:
						targetField = "password"
					case domain.RecordTypeAPIKey:
						targetField = "secret"
					case domain.RecordTypeNote:
						targetField = "content"
					default:
						targetField = "password"
					}
				}

				val, found := extractFieldValue(rec, targetField)
				if !found {
					return fmt.Errorf("field '%s' not found in record '%s'", targetField, rec.Title)
				}

				clipDriver := appCtx.Clipboard
				if clipDriver == nil {
					clipDriver = service.NewInMemoryClipboardDriver()
				}

				clipSvc := service.NewClipboardService(clipDriver)
				if err := clipSvc.Copy(ctx, val, clearAfter); err != nil {
					return fmt.Errorf("failed to copy to clipboard (headless environment? use --raw or stdout): %w", err)
				}

				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(map[string]any{
						"status":      "copied",
						"id":          rec.ID,
						"field":       targetField,
						"clear_after": clearAfter.String(),
					})
				}

				appCtx.Formatter.PrintText("✓ Copied to clipboard. Will auto-clear in %s.\n", clearAfter)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&fieldName, "field", "", "specific field to copy (defaults to password/secret/content)")
	cmd.Flags().DurationVar(&clearAfter, "clear-after", 45*time.Second, "auto-clear timeout duration (e.g. 45s, 1m, 30s)")

	return cmd
}
