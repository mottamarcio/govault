package screens_test

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/govault/internal/domain"
	"github.com/mottamarcio/govault/internal/service"
	"github.com/mottamarcio/govault/internal/storage/sqlite"
	"github.com/mottamarcio/govault/internal/tui/screens"
	"github.com/mottamarcio/govault/internal/tui/theme"
)

func setupTestRecordService(t *testing.T) (*service.RecordService, *service.Session, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vault.db")

	db, err := sqlite.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}

	ctx := context.Background()
	if err := db.Migrate(ctx); err != nil {
		db.Close()
		t.Fatalf("failed to migrate db: %v", err)
	}

	metaRepo := sqlite.NewMetadataRepository(db)
	tagRepo := sqlite.NewTagRepository(db)
	recordRepo := sqlite.NewRecordRepository(db, tagRepo)
	historyRepo := sqlite.NewHistoryRepository(db)

	vs := service.NewVaultService(db, metaRepo, dbPath)
	if err := vs.Init(ctx, "TestPassword123!"); err != nil {
		db.Close()
		t.Fatalf("failed to init vault: %v", err)
	}

	sess, err := vs.Unlock(ctx, "TestPassword123!")
	if err != nil {
		db.Close()
		t.Fatalf("failed to unlock: %v", err)
	}

	rs := service.NewRecordService(sess, recordRepo, tagRepo, historyRepo)

	cleanup := func() {
		_ = db.Close()
	}

	return rs, sess, cleanup
}

func TestDetailModel_ToggleSecret(t *testing.T) {
	th := theme.DefaultTheme()
	dm := screens.NewDetailModel(th)

	rec := &domain.Record{
		ID:        "rec-123",
		Title:     "AWS Root",
		Type:      domain.RecordTypeLogin,
		Tags:      []string{"cloud", "root"},
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Payload: &domain.LoginPayload{
			Username: "admin@example.com",
			Password: "SuperSecretPassword!",
			URI:      "https://aws.amazon.com",
		},
	}

	dm.SetRecord(rec)
	dm.SetDimensions(60, 20)

	// Initially masked
	viewMasked := dm.View()
	if strings.Contains(viewMasked, "SuperSecretPassword!") {
		t.Error("expected secret to be masked by default")
	}
	if !strings.Contains(viewMasked, "••••••••") {
		t.Error("expected mask bullets in initial detail view")
	}

	// Press 'v' to reveal
	dm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	viewRevealed := dm.View()
	if !strings.Contains(viewRevealed, "SuperSecretPassword!") {
		t.Errorf("expected revealed password in view, got: %s", viewRevealed)
	}
}

func TestHistoryModel_LoadAndBrowse(t *testing.T) {
	rs, _, cleanup := setupTestRecordService(t)
	defer cleanup()

	ctx := context.Background()
	// Create and update a record to generate history
	rec, err := rs.Create(ctx, service.RecordInput{
		Title: "Database Creds",
		Type:  domain.RecordTypeLogin,
		Payload: &domain.LoginPayload{
			Username: "dbuser",
			Password: "InitialPassword",
		},
	})
	if err != nil {
		t.Fatalf("failed to create record: %v", err)
	}

	// Update to v2
	_, err = rs.Update(ctx, rec.ID, service.RecordInput{
		Title:   "Database Creds",
		Type:    domain.RecordTypeLogin,
		Version: rec.Version,
		Payload: &domain.LoginPayload{
			Username: "dbuser",
			Password: "UpdatedPassword",
		},
	})
	if err != nil {
		t.Fatalf("failed to update record: %v", err)
	}

	th := theme.DefaultTheme()
	hm := screens.NewHistoryModel(th, rs)

	err = hm.LoadRecordHistory(rec.ID, rec.Title)
	if err != nil {
		t.Fatalf("failed to load history: %v", err)
	}

	if len(hm.Entries) != 1 {
		t.Fatalf("expected 1 history entry, got %d", len(hm.Entries))
	}

	view := hm.View()
	if !strings.Contains(view, "Revision History: Database Creds") {
		t.Errorf("expected history header, got: %s", view)
	}
	if !strings.Contains(view, "InitialPassword") {
		t.Errorf("expected previous password snapshot, got: %s", view)
	}
}
