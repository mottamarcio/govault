package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/spf13/cobra"
)

func newGetCmd(appCtx *AppContext) *cobra.Command {
	var (
		showPlaintext bool
		fieldName     string
		rawOutput     bool
	)

	cmd := &cobra.Command{
		Use:     "get <id|title>",
		Aliases: []string{"show"},
		Short:   "Retrieve and view a secret record",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			query := strings.TrimSpace(args[0])
			if query == "" {
				return fmt.Errorf("record ID or title is required")
			}

			return withRecordService(appCtx, func(ctx context.Context, recordSvc *service.RecordService) error {
				// Try fetching by ID first
				rec, err := recordSvc.GetByID(ctx, query)
				if err != nil {
					// Fallback: search by exact title or alias
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

				// Single field extraction mode
				if fieldName != "" {
					val, found := extractFieldValue(rec, fieldName)
					if !found {
						return fmt.Errorf("field '%s' not found in record '%s'", fieldName, rec.Title)
					}

					if rawOutput {
						// Stream exactly raw bytes to stdout without formatting
						fmt.Fprint(appCtx.Out, val)
						return nil
					}

					if appCtx.JSON {
						return appCtx.Formatter.PrintJSON(map[string]any{
							"id":    rec.ID,
							"field": fieldName,
							"value": val,
						})
					}

					appCtx.Formatter.PrintText("%s: %s\n", fieldName, val)
					return nil
				}

				// JSON mode: output full record
				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(rec)
				}

				// Tabular / Key-Value output
				formatRecordView(appCtx, rec, showPlaintext)
				return nil
			})
		},
	}

	cmd.Flags().BoolVarP(&showPlaintext, "show", "s", false, "display sensitive password/secret in plaintext")
	cmd.Flags().StringVar(&fieldName, "field", "", "extract specific field value (e.g. password, username, url)")
	cmd.Flags().BoolVarP(&rawOutput, "raw", "r", false, "output raw field value without labels or newline (for scripting pipelines)")

	return cmd
}

func maskSecret(val string, show bool) string {
	if show {
		return val
	}
	if val == "" {
		return ""
	}
	return "••••••••••••"
}

func extractFieldValue(rec *domain.Record, field string) (string, bool) {
	fieldLower := strings.ToLower(strings.TrimSpace(field))

	switch fieldLower {
	case "id":
		return rec.ID, true
	case "title":
		return rec.Title, true
	case "type":
		return rec.Type.String(), true
	case "tags":
		return strings.Join(rec.Tags, ","), true
	}

	switch p := rec.Payload.(type) {
	case domain.LoginPayload:
		switch fieldLower {
		case "username", "user":
			return p.Username, true
		case "password", "pass":
			return p.Password, true
		case "uri", "url":
			return p.URI, true
		case "notes":
			return p.Notes, true
		}
		for _, f := range p.CustomFields {
			if strings.EqualFold(f.Key, fieldLower) {
				return f.Value, true
			}
		}

	case domain.NotePayload:
		switch fieldLower {
		case "content", "note", "body":
			return p.Content, true
		}
		for _, f := range p.CustomFields {
			if strings.EqualFold(f.Key, fieldLower) {
				return f.Value, true
			}
		}

	case domain.APIKeyPayload:
		switch fieldLower {
		case "service":
			return p.Service, true
		case "key", "api_key":
			return p.Key, true
		case "secret", "api_secret":
			return p.Secret, true
		case "endpoint":
			return p.Endpoint, true
		case "notes":
			return p.Notes, true
		}
		for _, f := range p.CustomFields {
			if strings.EqualFold(f.Key, fieldLower) {
				return f.Value, true
			}
		}

	case domain.CustomPayload:
		switch fieldLower {
		case "notes":
			return p.Notes, true
		}
		for _, f := range p.Fields {
			if strings.EqualFold(f.Key, fieldLower) {
				return f.Value, true
			}
		}
	}

	return "", false
}

func formatRecordView(appCtx *AppContext, rec *domain.Record, show bool) {
	appCtx.Formatter.PrintText("ID:        %s\n", rec.ID)
	appCtx.Formatter.PrintText("Title:     %s\n", rec.Title)
	appCtx.Formatter.PrintText("Type:      %s\n", rec.Type)
	if len(rec.Tags) > 0 {
		appCtx.Formatter.PrintText("Tags:      %s\n", strings.Join(rec.Tags, ", "))
	}
	appCtx.Formatter.PrintText("Version:   %d\n", rec.Version)
	appCtx.Formatter.PrintText("Updated:   %s\n", rec.UpdatedAt.Format("2006-01-02 15:04:05"))
	appCtx.Formatter.PrintText("----------------------------------------\n")

	switch p := rec.Payload.(type) {
	case domain.LoginPayload:
		if p.Username != "" {
			appCtx.Formatter.PrintText("Username:  %s\n", p.Username)
		}
		appCtx.Formatter.PrintText("Password:  %s\n", maskSecret(p.Password, show))
		if p.URI != "" {
			appCtx.Formatter.PrintText("URI:       %s\n", p.URI)
		}
		if p.Notes != "" {
			appCtx.Formatter.PrintText("Notes:     %s\n", p.Notes)
		}
		for _, f := range p.CustomFields {
			val := f.Value
			if f.Masked {
				val = maskSecret(val, show)
			}
			appCtx.Formatter.PrintText("%-10s %s\n", f.Key+":", val)
		}

	case domain.NotePayload:
		appCtx.Formatter.PrintText("Content:\n%s\n", p.Content)
		for _, f := range p.CustomFields {
			appCtx.Formatter.PrintText("%-10s %s\n", f.Key+":", f.Value)
		}

	case domain.APIKeyPayload:
		if p.Service != "" {
			appCtx.Formatter.PrintText("Service:   %s\n", p.Service)
		}
		if p.Key != "" {
			appCtx.Formatter.PrintText("Key:       %s\n", p.Key)
		}
		appCtx.Formatter.PrintText("Secret:    %s\n", maskSecret(p.Secret, show))
		if p.Endpoint != "" {
			appCtx.Formatter.PrintText("Endpoint:  %s\n", p.Endpoint)
		}
		if p.Notes != "" {
			appCtx.Formatter.PrintText("Notes:     %s\n", p.Notes)
		}
		for _, f := range p.CustomFields {
			appCtx.Formatter.PrintText("%-10s %s\n", f.Key+":", f.Value)
		}

	case domain.CustomPayload:
		if p.Notes != "" {
			appCtx.Formatter.PrintText("Notes:     %s\n", p.Notes)
		}
		for _, f := range p.Fields {
			val := f.Value
			if f.Masked {
				val = maskSecret(val, show)
			}
			appCtx.Formatter.PrintText("%-10s %s\n", f.Key+":", val)
		}
	}
}
