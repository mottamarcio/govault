package service_test

import (
	"testing"
	"time"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
)

func TestSearchFilter(t *testing.T) {
	now := time.Now()
	records := []*domain.Record{
		{
			ID:        "rec-1",
			Title:     "Google Account",
			Type:      domain.RecordTypeLogin,
			Tags:      []string{"work", "email"},
			UpdatedAt: now.Add(-10 * time.Minute),
			Payload: domain.LoginPayload{
				Username: "dev@company.com",
				URI:      "https://accounts.google.com",
				Notes:    "Corporate SSO",
			},
		},
		{
			ID:        "rec-2",
			Title:     "AWS Root",
			Type:      domain.RecordTypeAPIKey,
			Tags:      []string{"infra", "cloud"},
			UpdatedAt: now.Add(-5 * time.Minute),
			Payload: domain.APIKeyPayload{
				Service: "Amazon Web Services",
				Key:     "AKIAEXAMPLE",
			},
		},
		{
			ID:        "rec-3",
			Title:     "Database Backup Notes",
			Type:      domain.RecordTypeNote,
			Tags:      []string{"infra"},
			UpdatedAt: now,
			Payload: domain.NotePayload{
				Content: "Restore steps with pg_dump and google cloud backup",
			},
		},
		{
			ID:        "rec-4",
			Title:     "Old Deleted Secret",
			Type:      domain.RecordTypeCustom,
			Tags:      []string{"trash"},
			DeletedAt: &now,
			Payload: domain.CustomPayload{
				Notes: "Deprecated credentials",
			},
		},
	}

	t.Run("Query matches title with high relevance", func(t *testing.T) {
		results := service.FilterAndRankRecords(records, service.SearchFilter{
			Query: "Google",
		})

		if len(results) != 2 {
			t.Fatalf("expected 2 matches for 'Google', got %d", len(results))
		}
		// Exact title prefix match should rank first
		if results[0].ID != "rec-1" {
			t.Fatalf("expected rec-1 to rank highest, got %s", results[0].ID)
		}
		if results[1].ID != "rec-3" {
			t.Fatalf("expected rec-3 to rank second, got %s", results[1].ID)
		}
	})

	t.Run("Filter by Type", func(t *testing.T) {
		results := service.FilterAndRankRecords(records, service.SearchFilter{
			Type: domain.RecordTypeAPIKey,
		})

		if len(results) != 1 || results[0].ID != "rec-2" {
			t.Fatalf("expected rec-2, got %+v", results)
		}
	})

	t.Run("Filter by Tags", func(t *testing.T) {
		results := service.FilterAndRankRecords(records, service.SearchFilter{
			Tags: []string{"infra"},
		})

		if len(results) != 2 {
			t.Fatalf("expected 2 matches for tag 'infra', got %d", len(results))
		}
	})

	t.Run("Deleted records excluded by default", func(t *testing.T) {
		results := service.FilterAndRankRecords(records, service.SearchFilter{
			Query: "Deleted",
		})
		if len(results) != 0 {
			t.Fatalf("expected 0 results, got %d", len(results))
		}

		resultsTrash := service.FilterAndRankRecords(records, service.SearchFilter{
			Query:        "Deleted",
			IncludeTrash: true,
		})
		if len(resultsTrash) != 1 || resultsTrash[0].ID != "rec-4" {
			t.Fatalf("expected rec-4 when IncludeTrash is true, got %+v", resultsTrash)
		}
	})
}
