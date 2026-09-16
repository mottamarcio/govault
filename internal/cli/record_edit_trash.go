package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/spf13/cobra"
)

func newEditCmd(appCtx *AppContext) *cobra.Command {
	var (
		title       string
		username    string
		uri         string
		notes       string
		apiKey      string
		apiSecret   string
		endpoint    string
		content     string
		tags        []string
		rawFields   []string
		autoGenPass bool
		passLen     int
	)

	cmd := &cobra.Command{
		Use:   "edit <id|title>",
		Short: "Edit an existing secret record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return fmt.Errorf("record ID or title is required")
			}

			return withRecordService(appCtx, func(ctx context.Context, recordSvc *service.RecordService) error {
				// Fetch existing record
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

				// Apply title update if provided
				newTitle := rec.Title
				if cmd.Flags().Changed("title") {
					newTitle = title
				}

				// Apply tags update if provided
				newTags := rec.Tags
				if cmd.Flags().Changed("tag") {
					newTags = tags
				}

				// Parse custom fields if provided
				var customFields []domain.Field
				if cmd.Flags().Changed("field") {
					for _, f := range rawFields {
						parts := strings.SplitN(f, "=", 2)
						if len(parts) != 2 {
							return fmt.Errorf("invalid field format '%s': expected key=value", f)
						}
						customFields = append(customFields, domain.Field{
							Key:    strings.TrimSpace(parts[0]),
							Value:  parts[1],
							Masked: false,
						})
					}
				}

				prompter := appCtx.Prompter
				if prompter == nil {
					prompter = DefaultPrompter(appCtx.In, appCtx.Err)
				}

				var newPayload any
				switch p := rec.Payload.(type) {
				case domain.LoginPayload:
					newUsername := p.Username
					if cmd.Flags().Changed("username") {
						newUsername = username
					}
					newURI := p.URI
					if cmd.Flags().Changed("uri") {
						newURI = uri
					}
					newNotes := p.Notes
					if cmd.Flags().Changed("notes") {
						newNotes = notes
					}
					newPassword := p.Password
					if autoGenPass {
						opts := service.DefaultPasswordOptions()
						if passLen >= 8 {
							opts.Length = passLen
						}
						newPassword, err = service.GeneratePassword(opts)
						if err != nil {
							return fmt.Errorf("failed to generate password: %w", err)
						}
					}
					fields := p.CustomFields
					if cmd.Flags().Changed("field") {
						fields = customFields
					}

					newPayload = domain.LoginPayload{
						Username:     newUsername,
						Password:     newPassword,
						URI:          newURI,
						Notes:        newNotes,
						CustomFields: fields,
					}

				case domain.NotePayload:
					newContent := p.Content
					if cmd.Flags().Changed("content") {
						newContent = content
					}
					fields := p.CustomFields
					if cmd.Flags().Changed("field") {
						fields = customFields
					}

					newPayload = domain.NotePayload{
						Content:      newContent,
						CustomFields: fields,
					}

				case domain.APIKeyPayload:
					newKey := p.Key
					if cmd.Flags().Changed("key") {
						newKey = apiKey
					}
					newSecret := p.Secret
					if cmd.Flags().Changed("secret") {
						newSecret = apiSecret
					} else if autoGenPass {
						opts := service.DefaultPasswordOptions()
						if passLen >= 8 {
							opts.Length = passLen
						}
						newSecret, err = service.GeneratePassword(opts)
						if err != nil {
							return fmt.Errorf("failed to generate secret: %w", err)
						}
					}
					newEndpoint := p.Endpoint
					if cmd.Flags().Changed("endpoint") {
						newEndpoint = endpoint
					}
					newNotes := p.Notes
					if cmd.Flags().Changed("notes") {
						newNotes = notes
					}
					fields := p.CustomFields
					if cmd.Flags().Changed("field") {
						fields = customFields
					}

					newPayload = domain.APIKeyPayload{
						Service:      newTitle,
						Key:          newKey,
						Secret:       newSecret,
						Endpoint:     newEndpoint,
						Notes:        newNotes,
						CustomFields: fields,
					}

				case domain.CustomPayload:
					newNotes := p.Notes
					if cmd.Flags().Changed("notes") {
						newNotes = notes
					}
					fields := p.Fields
					if cmd.Flags().Changed("field") {
						fields = customFields
					}

					newPayload = domain.CustomPayload{
						Notes:  newNotes,
						Fields: fields,
					}
				}

				input := service.RecordInput{
					Title:   newTitle,
					Type:    rec.Type,
					Tags:    newTags,
					Payload: newPayload,
				}

				updated, err := recordSvc.Update(ctx, rec.ID, input)
				if err != nil {
					return fmt.Errorf("failed to update record: %w", err)
				}

				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(map[string]any{
						"status":  "updated",
						"id":      updated.ID,
						"type":    updated.Type,
						"title":   updated.Title,
						"version": updated.Version,
					})
				}

				appCtx.Formatter.PrintText("✓ Updated secret '%s' (Version: %d)\n", updated.Title, updated.Version)
				return nil
			})
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "update title")
	cmd.Flags().StringVarP(&username, "username", "u", "", "update username")
	cmd.Flags().StringVar(&uri, "uri", "", "update URI")
	cmd.Flags().StringVar(&notes, "notes", "", "update notes")
	cmd.Flags().StringVar(&apiKey, "key", "", "update API key identifier")
	cmd.Flags().StringVar(&apiSecret, "secret", "", "update API secret")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "update API endpoint")
	cmd.Flags().StringVar(&content, "content", "", "update note content")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "update tags (replaces current tags)")
	cmd.Flags().StringSliceVar(&rawFields, "field", nil, "update custom fields in key=value format (replaces current fields)")
	cmd.Flags().BoolVarP(&autoGenPass, "generate", "g", false, "regenerate password with random secure characters")
	cmd.Flags().IntVar(&passLen, "length", 20, "length of regenerated password")

	return cmd
}

func newDeleteCmd(appCtx *AppContext) *cobra.Command {
	var permanent bool

	cmd := &cobra.Command{
		Use:     "delete <id|title>",
		Aliases: []string{"rm"},
		Short:   "Delete a secret record (moves to trash by default)",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return fmt.Errorf("record ID or title is required")
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

				if permanent {
					// Soft delete first then purge
					if err := recordSvc.Delete(ctx, rec.ID); err != nil {
						return fmt.Errorf("failed to delete record: %w", err)
					}
					if err := recordSvc.PurgeTrash(ctx); err != nil {
						return fmt.Errorf("failed to purge record: %w", err)
					}

					if appCtx.JSON {
						return appCtx.Formatter.PrintJSON(map[string]any{
							"status": "permanently_deleted",
							"id":     rec.ID,
						})
					}

					appCtx.Formatter.PrintText("✓ Permanently purged secret '%s' (ID: %s)\n", rec.Title, rec.ID)
					return nil
				}

				if err := recordSvc.Delete(ctx, rec.ID); err != nil {
					return fmt.Errorf("failed to delete record: %w", err)
				}

				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(map[string]any{
						"status": "trashed",
						"id":     rec.ID,
					})
				}

				appCtx.Formatter.PrintText("✓ Moved secret '%s' to trash (ID: %s)\n", rec.Title, rec.ID)
				return nil
			})
		},
	}

	cmd.Flags().BoolVar(&permanent, "permanent", false, "permanently delete the record immediately")
	return cmd
}

func newTrashCmd(appCtx *AppContext) *cobra.Command {
	trashCmd := &cobra.Command{
		Use:   "trash",
		Short: "Manage soft-deleted secret records",
	}

	// trash list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all items in the trash",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRecordService(appCtx, func(ctx context.Context, recordSvc *service.RecordService) error {
				records, err := recordSvc.ListTrash(ctx)
				if err != nil {
					return fmt.Errorf("failed to list trash: %w", err)
				}

				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(records)
				}

				printRecordsTable(appCtx, records)
				return nil
			})
		},
	}

	// trash restore <id>
	restoreCmd := &cobra.Command{
		Use:   "restore <id|title>",
		Short: "Restore a record from the trash",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			return withRecordService(appCtx, func(ctx context.Context, recordSvc *service.RecordService) error {
				// Find target record in trash
				trashItems, err := recordSvc.ListTrash(ctx)
				if err != nil {
					return fmt.Errorf("failed to inspect trash: %w", err)
				}

				var target *domain.Record
				for _, r := range trashItems {
					if r.ID == query || strings.EqualFold(r.Title, query) {
						target = r
						break
					}
				}

				if target == nil {
					return fmt.Errorf("item '%s' not found in trash", query)
				}

				if err := recordSvc.Restore(ctx, target.ID); err != nil {
					return fmt.Errorf("failed to restore record: %w", err)
				}

				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(map[string]any{
						"status": "restored",
						"id":     target.ID,
					})
				}

				appCtx.Formatter.PrintText("✓ Restored secret '%s' from trash\n", target.Title)
				return nil
			})
		},
	}

	// trash purge
	purgeCmd := &cobra.Command{
		Use:   "purge",
		Short: "Permanently remove all records currently in the trash",
		RunE: func(cmd *cobra.Command, args []string) error {
			return withRecordService(appCtx, func(ctx context.Context, recordSvc *service.RecordService) error {
				if err := recordSvc.PurgeTrash(ctx); err != nil {
					return fmt.Errorf("failed to purge trash: %w", err)
				}

				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(map[string]any{
						"status": "trash_purged",
					})
				}

				appCtx.Formatter.PrintText("✓ All items in trash permanently purged\n")
				return nil
			})
		},
	}

	trashCmd.AddCommand(listCmd)
	trashCmd.AddCommand(restoreCmd)
	trashCmd.AddCommand(purgeCmd)

	return trashCmd
}
