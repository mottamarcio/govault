package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/spf13/cobra"
)

func newListCmd(appCtx *AppContext) *cobra.Command {
	var (
		recTypeStr   string
		tags         []string
		includeTrash bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List secret records in the vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			var recType domain.RecordType
			if recTypeStr != "" {
				rawType := strings.ToLower(strings.TrimSpace(recTypeStr))
				if rawType == "apikey" {
					rawType = "api_key"
				}
				recType = domain.RecordType(rawType)
			}

			return withRecordService(appCtx, func(ctx context.Context, recordSvc *service.RecordService) error {
				filter := service.SearchFilter{
					Type:         recType,
					Tags:         tags,
					IncludeTrash: includeTrash,
				}

				records, err := recordSvc.Search(ctx, filter)
				if err != nil {
					return fmt.Errorf("failed to list records: %w", err)
				}

				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(records)
				}

				printRecordsTable(appCtx, records)
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&recTypeStr, "type", "t", "", "filter by record type (login, note, apikey, custom)")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "filter by tag (can be specified multiple times)")
	cmd.Flags().BoolVar(&includeTrash, "trash", false, "include soft-deleted items from trash")

	return cmd
}

func newSearchCmd(appCtx *AppContext) *cobra.Command {
	var (
		recTypeStr   string
		tags         []string
		includeTrash bool
	)

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search secret records by title, tag, or content",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}

			var recType domain.RecordType
			if recTypeStr != "" {
				rawType := strings.ToLower(strings.TrimSpace(recTypeStr))
				if rawType == "apikey" {
					rawType = "api_key"
				}
				recType = domain.RecordType(rawType)
			}

			return withRecordService(appCtx, func(ctx context.Context, recordSvc *service.RecordService) error {
				filter := service.SearchFilter{
					Query:        query,
					Type:         recType,
					Tags:         tags,
					IncludeTrash: includeTrash,
				}

				records, err := recordSvc.Search(ctx, filter)
				if err != nil {
					return fmt.Errorf("failed to search records: %w", err)
				}

				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(records)
				}

				printRecordsTable(appCtx, records)
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&recTypeStr, "type", "t", "", "filter by record type (login, note, apikey, custom)")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "filter by tag (can be specified multiple times)")
	cmd.Flags().BoolVar(&includeTrash, "trash", false, "include soft-deleted items from trash")

	return cmd
}

func printRecordsTable(appCtx *AppContext, records []*domain.Record) {
	if len(records) == 0 {
		appCtx.Formatter.PrintText("No records found.\n")
		return
	}

	// POSIX Tabular format with fixed column headers
	fmt.Fprintf(appCtx.Out, "%-36s  %-8s  %-24s  %-16s  %s\n", "ID", "TYPE", "TITLE", "TAGS", "UPDATED")
	fmt.Fprintf(appCtx.Out, "%-36s  %-8s  %-24s  %-16s  %s\n", strings.Repeat("-", 36), strings.Repeat("-", 8), strings.Repeat("-", 24), strings.Repeat("-", 16), strings.Repeat("-", 19))

	for _, r := range records {
		tagsStr := strings.Join(r.Tags, ",")
		if len(tagsStr) > 16 {
			tagsStr = tagsStr[:13] + "..."
		}
		titleStr := r.Title
		if len(titleStr) > 24 {
			titleStr = titleStr[:21] + "..."
		}

		fmt.Fprintf(appCtx.Out, "%-36s  %-8s  %-24s  %-16s  %s\n",
			r.ID,
			r.Type,
			titleStr,
			tagsStr,
			r.UpdatedAt.Format("2006-01-02 15:04:05"),
		)
	}
}
