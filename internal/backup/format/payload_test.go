package format_test

import (
	"errors"
	"testing"
	"time"

	"github.com/mottamarcio/govault/internal/backup/format"
	"github.com/mottamarcio/govault/internal/domain"
)

func TestPayloadSerializationRoundTrip(t *testing.T) {
	loginPayload := domain.LoginPayload{
		Username: "user@example.com",
		Password: "SuperSecretPassword123!",
		URI:      "https://example.com/login",
	}
	loginBytes, err := domain.SerializePayload(loginPayload)
	if err != nil {
		t.Fatalf("failed to serialize login payload: %v", err)
	}

	payload := &format.BackupPayload{
		Version: 1,
		Manifest: format.BackupManifest{
			SourceVaultVersion:  1,
			SourceSchemaVersion: 1,
			AppVersion:          "0.1.0",
			CreatedAtUnix:       time.Now().Unix(),
		},
		Entries: []format.BackupEntry{
			{
				ID:            "entry-2",
				Type:          domain.RecordTypeLogin,
				Title:         "Example Account 2",
				Payload:       loginBytes,
				Version:       1,
				CreatedAtUnix: time.Now().Unix(),
				UpdatedAtUnix: time.Now().Unix(),
				Tags:          []string{"Work"},
			},
			{
				ID:            "entry-1",
				Type:          domain.RecordTypeLogin,
				Title:         "Example Account 1",
				Payload:       loginBytes,
				Version:       1,
				CreatedAtUnix: time.Now().Unix(),
				UpdatedAtUnix: time.Now().Unix(),
				Tags:          []string{"Personal"},
			},
		},
		Histories: []format.BackupHistory{
			{
				HistoryID:     "hist-1",
				RecordID:      "entry-1",
				Version:       1,
				Payload:       loginBytes,
				CreatedAtUnix: time.Now().Unix(),
			},
		},
		Tags: []format.BackupTag{
			{ID: "tag-2", Name: "Work", CreatedAtUnix: time.Now().Unix()},
			{ID: "tag-1", Name: "Personal", CreatedAtUnix: time.Now().Unix()},
		},
	}

	data, err := payload.Marshal()
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	unmarshaled, err := format.UnmarshalPayload(data)
	if err != nil {
		t.Fatalf("failed to unmarshal payload: %v", err)
	}

	if unmarshaled.Manifest.EntryCount != 2 {
		t.Fatalf("expected 2 entries, got %d", unmarshaled.Manifest.EntryCount)
	}
	if unmarshaled.Manifest.HistoryCount != 1 {
		t.Fatalf("expected 1 history record, got %d", unmarshaled.Manifest.HistoryCount)
	}
	if unmarshaled.Manifest.TagCount != 2 {
		t.Fatalf("expected 2 tags, got %d", unmarshaled.Manifest.TagCount)
	}

	// Verify deterministic sorting: entry-1 should come before entry-2
	if unmarshaled.Entries[0].ID != "entry-1" || unmarshaled.Entries[1].ID != "entry-2" {
		t.Fatalf("entries not sorted deterministically: %s, %s", unmarshaled.Entries[0].ID, unmarshaled.Entries[1].ID)
	}
}

func TestPayloadValidationErrors(t *testing.T) {
	t.Run("Duplicate entry ID fails", func(t *testing.T) {
		payload := &format.BackupPayload{
			Version: 1,
			Manifest: format.BackupManifest{
				EntryCount: 2,
			},
			Entries: []format.BackupEntry{
				{ID: "entry-1"},
				{ID: "entry-1"},
			},
		}
		data, err := payload.Marshal()
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		_, err = format.UnmarshalPayload(data)
		if !errors.Is(err, format.ErrCorruptPayload) {
			t.Fatalf("expected ErrCorruptPayload for duplicate entry ID, got %v", err)
		}
	})

	t.Run("Orphaned history fails", func(t *testing.T) {
		payload := &format.BackupPayload{
			Version: 1,
			Manifest: format.BackupManifest{
				EntryCount:   1,
				HistoryCount: 1,
			},
			Entries: []format.BackupEntry{
				{ID: "entry-1"},
			},
			Histories: []format.BackupHistory{
				{HistoryID: "hist-1", RecordID: "missing-entry"},
			},
		}
		data, err := payload.Marshal()
		if err != nil {
			t.Fatalf("failed to marshal: %v", err)
		}

		_, err = format.UnmarshalPayload(data)
		if !errors.Is(err, format.ErrCorruptPayload) {
			t.Fatalf("expected ErrCorruptPayload for orphaned history, got %v", err)
		}
	})
}
