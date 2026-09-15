package domain_test

import (
	"strings"
	"testing"
	"time"

	"github.com/mottamarcio/govault/internal/domain"
)

func TestRecordValidation(t *testing.T) {
	validRecord := func() domain.Record {
		return domain.Record{
			ID:        "550e8400-e29b-41d4-a716-446655440000",
			VaultID:   "vault-123",
			Title:     "My Gmail Account",
			Type:      domain.RecordTypeLogin,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	t.Run("Valid Record passes", func(t *testing.T) {
		r := validRecord()
		if err := r.Validate(); err != nil {
			t.Fatalf("expected valid record, got error: %v", err)
		}
	})

	t.Run("Empty ID fails", func(t *testing.T) {
		r := validRecord()
		r.ID = ""
		if err := r.Validate(); err != domain.ErrInvalidUUID {
			t.Fatalf("expected ErrInvalidUUID, got: %v", err)
		}
	})

	t.Run("Empty VaultID fails", func(t *testing.T) {
		r := validRecord()
		r.VaultID = "   "
		if err := r.Validate(); err != domain.ErrInvalidUUID {
			t.Fatalf("expected ErrInvalidUUID, got: %v", err)
		}
	})

	t.Run("Empty Title fails", func(t *testing.T) {
		r := validRecord()
		r.Title = "   "
		if err := r.Validate(); err != domain.ErrEmptyTitle {
			t.Fatalf("expected ErrEmptyTitle, got: %v", err)
		}
	})

	t.Run("Title too long fails", func(t *testing.T) {
		r := validRecord()
		r.Title = strings.Repeat("A", 256)
		if err := r.Validate(); err != domain.ErrTitleTooLong {
			t.Fatalf("expected ErrTitleTooLong, got: %v", err)
		}
	})

	t.Run("Invalid RecordType fails", func(t *testing.T) {
		r := validRecord()
		r.Type = "unknown_type"
		if err := r.Validate(); err != domain.ErrInvalidRecordType {
			t.Fatalf("expected ErrInvalidRecordType, got: %v", err)
		}
	})
}

func TestPayloadValidation(t *testing.T) {
	t.Run("LoginPayload with unique fields passes", func(t *testing.T) {
		p := domain.LoginPayload{
			CustomFields: []domain.Field{
				{Key: "PIN", Value: "123"},
				{Key: "Backup Phone", Value: "+123456789"},
			},
		}
		if err := p.Validate(); err != nil {
			t.Fatalf("expected valid payload, got: %v", err)
		}
	})

	t.Run("LoginPayload with duplicate field keys fails", func(t *testing.T) {
		p := domain.LoginPayload{
			CustomFields: []domain.Field{
				{Key: "PIN", Value: "123"},
				{Key: "pin", Value: "456"},
			},
		}
		if err := p.Validate(); err != domain.ErrDuplicateFieldKey {
			t.Fatalf("expected ErrDuplicateFieldKey, got: %v", err)
		}
	})

	t.Run("NotePayload with empty field key fails", func(t *testing.T) {
		p := domain.NotePayload{
			CustomFields: []domain.Field{
				{Key: "   ", Value: "val"},
			},
		}
		if err := p.Validate(); err != domain.ErrEmptyKey {
			t.Fatalf("expected ErrEmptyKey, got: %v", err)
		}
	})

	t.Run("APIKeyPayload with duplicate field keys fails", func(t *testing.T) {
		p := domain.APIKeyPayload{
			CustomFields: []domain.Field{
				{Key: "Env", Value: "prod"},
				{Key: "env", Value: "staging"},
			},
		}
		if err := p.Validate(); err != domain.ErrDuplicateFieldKey {
			t.Fatalf("expected ErrDuplicateFieldKey, got: %v", err)
		}
	})

	t.Run("CustomPayload with valid fields passes", func(t *testing.T) {
		p := domain.CustomPayload{
			Fields: []domain.Field{
				{Key: "Port", Value: "8080"},
				{Key: "Host", Value: "localhost"},
			},
		}
		if err := p.Validate(); err != nil {
			t.Fatalf("expected valid payload, got: %v", err)
		}
	})
}

func TestTagValidation(t *testing.T) {
	t.Run("Valid Tag passes", func(t *testing.T) {
		tag := domain.Tag{
			ID:        "tag-1",
			VaultID:   "vault-1",
			Name:      "Production",
			CreatedAt: time.Now(),
		}
		if err := tag.Validate(); err != nil {
			t.Fatalf("expected valid tag, got: %v", err)
		}
	})

	t.Run("Empty Tag Name fails", func(t *testing.T) {
		tag := domain.Tag{
			ID:      "tag-1",
			VaultID: "vault-1",
			Name:    "   ",
		}
		if err := tag.Validate(); err != domain.ErrEmptyTagName {
			t.Fatalf("expected ErrEmptyTagName, got: %v", err)
		}
	})

	t.Run("Tag Name too long fails", func(t *testing.T) {
		tag := domain.Tag{
			ID:      "tag-1",
			VaultID: "vault-1",
			Name:    strings.Repeat("T", 65),
		}
		if err := tag.Validate(); err != domain.ErrTagNameTooLong {
			t.Fatalf("expected ErrTagNameTooLong, got: %v", err)
		}
	})
}
