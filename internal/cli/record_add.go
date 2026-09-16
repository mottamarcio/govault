package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/spf13/cobra"
)

func newAddCmd(appCtx *AppContext) *cobra.Command {
	var (
		recTypeStr  string
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
		Use:   "add",
		Short: "Add a new encrypted secret record",
		RunE: func(cmd *cobra.Command, args []string) error {
			if recTypeStr == "" {
				recTypeStr = "login"
			}
			rawType := strings.ToLower(strings.TrimSpace(recTypeStr))
			if rawType == "apikey" {
				rawType = "api_key"
			}
			recType := domain.RecordType(rawType)
			if !recType.IsValid() {
				return fmt.Errorf("invalid record type '%s': must be login, note, apikey, or custom", recTypeStr)
			}

			// Parse custom fields
			var customFields []domain.Field
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

			prompter := appCtx.Prompter
			if prompter == nil {
				prompter = DefaultPrompter(appCtx.In, appCtx.Err)
			}

			// Construct typed payload
			var payload any
			switch recType {
			case domain.RecordTypeLogin:
				var password string
				if autoGenPass {
					opts := service.DefaultPasswordOptions()
					if passLen >= 8 {
						opts.Length = passLen
					}
					var err error
					password, err = service.GeneratePassword(opts)
					if err != nil {
						return fmt.Errorf("failed to generate password: %w", err)
					}
				} else {
					var err error
					password, err = prompter.ReadPasswordConfirm("Enter secret password")
					if err != nil {
						return err
					}
				}

				loginURI := uri
				if loginURI == "" {
					loginURI = title
				}
				payload = domain.LoginPayload{
					Username:     username,
					Password:     password,
					URI:          loginURI,
					Notes:        notes,
					CustomFields: customFields,
				}

			case domain.RecordTypeNote:
				if content == "" {
					var err error
					content, err = prompter.ReadPrompt("Enter note content")
					if err != nil {
						return err
					}
				}
				payload = domain.NotePayload{
					Content:      content,
					CustomFields: customFields,
				}

			case domain.RecordTypeAPIKey:
				if apiKey == "" {
					var err error
					apiKey, err = prompter.ReadPrompt("Enter API key identifier")
					if err != nil {
						return err
					}
				}
				if apiSecret == "" {
					if autoGenPass {
						opts := service.DefaultPasswordOptions()
						if passLen >= 8 {
							opts.Length = passLen
						}
						var err error
						apiSecret, err = service.GeneratePassword(opts)
						if err != nil {
							return fmt.Errorf("failed to generate secret: %w", err)
						}
					} else {
						var err error
						apiSecret, err = prompter.ReadPasswordConfirm("Enter API secret")
						if err != nil {
							return err
						}
					}
				}
				payload = domain.APIKeyPayload{
					Service:      title,
					Key:          apiKey,
					Secret:       apiSecret,
					Endpoint:     endpoint,
					Notes:        notes,
					CustomFields: customFields,
				}

			case domain.RecordTypeCustom:
				payload = domain.CustomPayload{
					Notes:  notes,
					Fields: customFields,
				}
			}

			return withRecordService(appCtx, func(ctx context.Context, recordSvc *service.RecordService) error {
				input := service.RecordInput{
					Title:   title,
					Type:    recType,
					Tags:    tags,
					Payload: payload,
				}

				rec, err := recordSvc.Create(ctx, input)
				if err != nil {
					return fmt.Errorf("failed to create record: %w", err)
				}

				if appCtx.JSON {
					return appCtx.Formatter.PrintJSON(map[string]any{
						"status":  "created",
						"id":      rec.ID,
						"type":    rec.Type,
						"title":   rec.Title,
						"tags":    rec.Tags,
						"version": rec.Version,
					})
				}

				appCtx.Formatter.PrintText("✓ Created %s secret '%s' (ID: %s)\n", rec.Type, rec.Title, rec.ID)
				return nil
			})
		},
	}

	cmd.Flags().StringVarP(&recTypeStr, "type", "t", "login", "secret type (login, note, apikey, custom)")
	cmd.Flags().StringVar(&title, "title", "", "title or service name of the secret")
	cmd.Flags().StringVarP(&username, "username", "u", "", "username or login identity")
	cmd.Flags().StringVar(&uri, "uri", "", "URL or service endpoint")
	cmd.Flags().StringVar(&notes, "notes", "", "optional notes or remarks")
	cmd.Flags().StringVar(&apiKey, "key", "", "API key identifier (for apikey type)")
	cmd.Flags().StringVar(&apiSecret, "secret", "", "API secret value (for apikey type, or prompted)")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "API endpoint URL (for apikey type)")
	cmd.Flags().StringVar(&content, "content", "", "note content (for note type)")
	cmd.Flags().StringSliceVar(&tags, "tag", nil, "tags to associate with record (can be specified multiple times)")
	cmd.Flags().StringSliceVar(&rawFields, "field", nil, "custom field in key=value format (can be specified multiple times)")
	cmd.Flags().BoolVarP(&autoGenPass, "generate", "g", false, "auto-generate a secure random password")
	cmd.Flags().IntVar(&passLen, "length", 20, "length of auto-generated password")

	return cmd
}
