package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/mottamarcio/govault/internal/service"
)

func TestClipboardService(t *testing.T) {
	ctx := context.Background()
	driver := service.NewInMemoryClipboardDriver()
	svc := service.NewClipboardService(driver)

	t.Run("Immediate copy and read", func(t *testing.T) {
		err := svc.Copy(ctx, "secret-token-123", 0)
		if err != nil {
			t.Fatalf("Copy failed: %v", err)
		}

		got, err := svc.Read()
		if err != nil || got != "secret-token-123" {
			t.Fatalf("Read got %q, want secret-token-123", got)
		}

		_ = svc.Clear()
	})

	t.Run("Timed auto-clear", func(t *testing.T) {
		timeout := 50 * time.Millisecond
		err := svc.Copy(ctx, "temporary-secret", timeout)
		if err != nil {
			t.Fatalf("Copy failed: %v", err)
		}

		got, _ := svc.Read()
		if got != "temporary-secret" {
			t.Fatalf("expected temporary-secret immediately, got %q", got)
		}

		// Wait for timer to expire
		time.Sleep(100 * time.Millisecond)

		gotAfter, _ := svc.Read()
		if gotAfter != "" {
			t.Fatalf("expected clipboard to be cleared, got %q", gotAfter)
		}
	})

	t.Run("Overwriting content cancels previous timer without clearing new content", func(t *testing.T) {
		// 1. Copy first text with 100ms timeout
		_ = svc.Copy(ctx, "first-secret", 100*time.Millisecond)

		// 2. Wait 30ms and copy second text with 300ms timeout
		time.Sleep(30 * time.Millisecond)
		_ = svc.Copy(ctx, "second-secret", 300*time.Millisecond)

		// 3. Wait until first timeout would have expired (at 120ms total)
		time.Sleep(90 * time.Millisecond)

		// Second secret must still be in clipboard
		got, _ := svc.Read()
		if got != "second-secret" {
			t.Fatalf("expected second-secret to remain, got %q", got)
		}

		// 4. Wait for second timeout to expire
		time.Sleep(250 * time.Millisecond)
		gotFinal, _ := svc.Read()
		if gotFinal != "" {
			t.Fatalf("expected clipboard to be cleared after second timeout, got %q", gotFinal)
		}
	})
}
