package screens_test

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/tui/screens"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

func TestEditorModel_CreateAndValidate(t *testing.T) {
	rs, _, cleanup := setupTestRecordService(t)
	defer cleanup()

	th := theme.DefaultTheme()
	m := screens.NewEditorModel(th, rs)
	m.SetDimensions(80, 24)

	m.OpenCreate(domain.RecordTypeLogin)
	if !m.Active {
		t.Fatalf("expected editor to be active")
	}

	// Try saving with empty title -> validation error
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.ValidationError == "" {
		t.Fatalf("expected validation error on empty title")
	}

	// Set valid fields
	m.TitleInput.SetValue("GitHub Work")
	m.UsernameInput.SetValue("octocat")
	m.SecretInput.SetValue("s3cr3tP@ss")
	m.URIInput.SetValue("https://github.com")
	m.TagsInput.SetValue("work, vcs")
	m.NotesArea.SetValue("Main corporate login")

	// Add custom field
	m.AddCustomFieldWithValues("2FA Recovery", "ABCDEF-123456")

	// Save
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatalf("expected non-nil cmd on successful save")
	}
	saveMsg, ok := cmd().(screens.EditorSaveMsg)
	if !ok || saveMsg.Record == nil {
		t.Fatalf("expected valid EditorSaveMsg, got %+v", saveMsg)
	}

	if saveMsg.Record.Title != "GitHub Work" || saveMsg.Record.Version != 1 {
		t.Fatalf("unexpected saved record: %+v", saveMsg.Record)
	}

	// Verify custom fields persisted
	loginP, ok := saveMsg.Record.Payload.(*domain.LoginPayload)
	if !ok {
		loginVal := saveMsg.Record.Payload.(domain.LoginPayload)
		loginP = &loginVal
	}
	if len(loginP.CustomFields) != 1 || loginP.CustomFields[0].Key != "2FA Recovery" {
		t.Fatalf("expected custom fields in saved payload: %+v", loginP)
	}
}

func TestEditorModel_EditExistingRecord(t *testing.T) {
	rs, _, cleanup := setupTestRecordService(t)
	defer cleanup()

	th := theme.DefaultTheme()
	m := screens.NewEditorModel(th, rs)
	m.SetDimensions(80, 24)

	// Create initial record
	rec, err := rs.Create(context.Background(), service.RecordInput{
		Title: "AWS Prod",
		Type:  domain.RecordTypeLogin,
		Tags:  []string{"cloud"},
		Payload: &domain.LoginPayload{
			Username: "admin",
			Password: "oldpassword",
			URI:      "https://aws.amazon.com",
		},
	})
	if err != nil {
		t.Fatalf("failed to create record: %v", err)
	}

	// Open edit
	m.OpenEdit(rec)
	if m.Mode != screens.EditorModeEdit || m.TitleInput.Value() != "AWS Prod" {
		t.Fatalf("expected edit mode with AWS Prod")
	}

	// Update password and tags
	m.SecretInput.SetValue("newSecr3tPassword!")
	m.TagsInput.SetValue("cloud, aws, prod")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd == nil {
		t.Fatalf("expected non-nil cmd on save")
	}
	saveMsg := cmd().(screens.EditorSaveMsg)
	if saveMsg.Record.Version != 2 {
		t.Fatalf("expected version 2 after update, got %d", saveMsg.Record.Version)
	}
	if len(saveMsg.Record.Tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(saveMsg.Record.Tags))
	}
}

func TestEditorModel_CustomFieldsAndDiscard(t *testing.T) {
	rs, _, cleanup := setupTestRecordService(t)
	defer cleanup()

	th := theme.DefaultTheme()
	m := screens.NewEditorModel(th, rs)
	m.SetDimensions(80, 24)

	m.OpenCreate(domain.RecordTypeCustom)
	m.TitleInput.SetValue("Wi-Fi Settings")
	m.AddCustomFieldWithValues("SSID", "Office5G")
	m.AddCustomFieldWithValues("PSK", "supersecretwifi")

	// Try duplicate key
	m.AddCustomFieldWithValues("SSID", "Duplicate")
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.ValidationError == "" {
		t.Fatalf("expected duplicate key validation error")
	}

	// Remove last custom field
	m.FocusIndex = int(screens.FocusCustomStart) + 4 // 3rd custom field key
	m.RemoveFocusedCustomField()
	if len(m.CustomFields) != 2 {
		t.Fatalf("expected 2 custom fields after deletion, got %d", len(m.CustomFields))
	}

	// Discard changes test
	m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.DiscardModal.Active {
		t.Fatalf("expected discard confirmation modal to be active")
	}

	// Confirm discard
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	if m.Active {
		t.Fatalf("expected editor to reset and become inactive")
	}
}
