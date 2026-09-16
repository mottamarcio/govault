package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mottamarcio/govault/internal/cli"
)

func setupTestVault(t *testing.T, password string) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "govault-record-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	vaultPath := filepath.Join(tempDir, "vault.db")

	var stdout, stderr bytes.Buffer
	stdin := strings.NewReader(password + "\n" + password + "\n")
	appCtx := &cli.AppContext{
		VaultPath: vaultPath,
		In:        stdin,
		Out:       &stdout,
		Err:       &stderr,
	}

	rootCmd := cli.NewRootCommand(appCtx)
	rootCmd.SetArgs([]string{"init"})
	if err := rootCmd.Execute(); err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("init failed: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}
	return vaultPath, cleanup
}

func TestRecordAdd(t *testing.T) {
	vaultPath, cleanup := setupTestVault(t, "master-secret-123")
	defer cleanup()

	t.Run("Add Login with Auto-Generated Password", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{
			"add",
			"--type", "login",
			"--title", "GitHub",
			"--username", "alice",
			"--uri", "https://github.com",
			"--tag", "dev",
			"--tag", "work",
			"--field", "pin=9988",
			"--generate",
			"--length", "24",
		})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("add login command failed: %v (stderr: %s)", err, stderr.String())
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "Created login secret 'GitHub'") {
			t.Errorf("unexpected add output: %s", outStr)
		}
	})

	t.Run("Add APIKey with Explicit Prompt and JSON Output", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		// 1: API secret
		// 2: API secret confirmation
		// 3: master password to unlock vault
		stdin := strings.NewReader("sk-test-secret-999\nsk-test-secret-999\nmaster-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
			JSON:      true,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{
			"add",
			"--type", "apikey",
			"--title", "Stripe API",
			"--key", "pk_live_12345",
			"--tag", "payments",
			"--json",
		})

		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("add apikey command failed: %v (stderr: %s)", err, stderr.String())
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, `"status": "created"`) || !strings.Contains(outStr, `"title": "Stripe API"`) {
			t.Errorf("unexpected json output: %s", outStr)
		}
	})
}

func TestRecordGet(t *testing.T) {
	vaultPath, cleanup := setupTestVault(t, "master-secret-123")
	defer cleanup()

	// Pre-populate with a login record
	{
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{
			"add",
			"--type", "login",
			"--title", "AWS Console",
			"--username", "admin",
			"--generate",
			"--length", "16",
		})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("failed to seed record: %v", err)
		}
	}

	t.Run("Get Record Masked by Default", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"get", "AWS Console"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("get command failed: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "Title:     AWS Console") {
			t.Errorf("expected title in output: %s", outStr)
		}
		if !strings.Contains(outStr, "Password:  ••••••••••••") {
			t.Errorf("expected password to be masked: %s", outStr)
		}
	})

	t.Run("Get Record with Explicit --show", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"get", "AWS Console", "--show"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("get --show command failed: %v", err)
		}

		outStr := stdout.String()
		if strings.Contains(outStr, "••••••••••••") {
			t.Errorf("expected unmasked password with --show: %s", outStr)
		}
	})

	t.Run("Get Single Field with Raw Pipeline Output", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"get", "AWS Console", "--field", "username", "--raw"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("get raw username failed: %v", err)
		}

		if stdout.String() != "admin" {
			t.Errorf("expected exactly 'admin', got '%s'", stdout.String())
		}
	})

	t.Run("Get Non-Existent Record Returns Error", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"get", "Non-Existent-Service"})
		err := rootCmd.Execute()
		if err == nil {
			t.Fatal("expected error for non-existent record, got nil")
		}
	})
}

func TestRecordListAndSearch(t *testing.T) {
	vaultPath, cleanup := setupTestVault(t, "master-secret-123")
	defer cleanup()

	// Seed multiple records
	recordsToSeed := [][]string{
		{"add", "--type", "login", "--title", "GitHub Pro", "--username", "alice", "--tag", "dev", "--tag", "vcs", "--generate"},
		{"add", "--type", "login", "--title", "GitLab Enterprise", "--username", "alice", "--tag", "dev", "--generate"},
		{"add", "--type", "note", "--title", "Prod Infrastructure", "--content", "Secret servers 10.0.0.1", "--tag", "ops", "--tag", "prod"},
	}

	for _, args := range recordsToSeed {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs(args)
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("failed to seed record with args %v: %v", args, err)
		}
	}

	t.Run("List Tabular Records", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"list"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("list command failed: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "GitHub Pro") || !strings.Contains(outStr, "GitLab Enterprise") || !strings.Contains(outStr, "Secret servers 10.0.0.1") {
			t.Errorf("list output missing seeded records: %s", outStr)
		}
	})

	t.Run("List Filtered by Tag and Type", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
			JSON:      true,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"list", "--type", "login", "--tag", "vcs", "--json"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("list filtered command failed: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "GitHub Pro") {
			t.Errorf("expected GitHub Pro in output: %s", outStr)
		}
		if strings.Contains(outStr, "GitLab Enterprise") || strings.Contains(outStr, "Prod Infrastructure") {
			t.Errorf("unexpected matches in filtered list: %s", outStr)
		}
	})

	t.Run("Search Query Relevance", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"search", "GitLab"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("search command failed: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "GitLab Enterprise") {
			t.Errorf("expected GitLab in search output: %s", outStr)
		}
		if strings.Contains(outStr, "Prod Infrastructure") {
			t.Errorf("did not expect unrelated record in search output: %s", outStr)
		}
	})
}

func TestRecordEditDeleteTrash(t *testing.T) {
	vaultPath, cleanup := setupTestVault(t, "master-secret-123")
	defer cleanup()

	// 1. Add record
	{
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{
			"add",
			"--type", "login",
			"--title", "Slack Workspace",
			"--username", "bob",
			"--generate",
		})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("failed to add initial record: %v", err)
		}
	}

	// 2. Edit record
	t.Run("Edit Record Details", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{
			"edit", "Slack Workspace",
			"--username", "bob_senior",
			"--tag", "work",
			"--tag", "chat",
		})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("edit failed: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "Updated secret 'Slack Workspace'") || !strings.Contains(outStr, "Version: 2") {
			t.Errorf("unexpected edit output: %s", outStr)
		}
	})

	// 3. Delete to Trash
	t.Run("Delete to Trash", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}
		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"delete", "Slack Workspace"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("delete failed: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "Moved secret 'Slack Workspace' to trash") {
			t.Errorf("unexpected delete output: %s", outStr)
		}
	})

	// 4. Verify in Trash List and Absent from Active List
	t.Run("Trash List and Restore", func(t *testing.T) {
		// Verify active list does not show it
		{
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader("master-secret-123\n")
			appCtx := &cli.AppContext{
				VaultPath: vaultPath,
				In:        stdin,
				Out:       &stdout,
				Err:       &stderr,
			}
			rootCmd := cli.NewRootCommand(appCtx)
			rootCmd.SetArgs([]string{"list"})
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("list failed: %v", err)
			}
			if strings.Contains(stdout.String(), "Slack Workspace") {
				t.Errorf("deleted record should not be visible in active list")
			}
		}

		// Check trash list
		{
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader("master-secret-123\n")
			appCtx := &cli.AppContext{
				VaultPath: vaultPath,
				In:        stdin,
				Out:       &stdout,
				Err:       &stderr,
			}
			rootCmd := cli.NewRootCommand(appCtx)
			rootCmd.SetArgs([]string{"trash", "list"})
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("trash list failed: %v", err)
			}
			if !strings.Contains(stdout.String(), "Slack Workspace") {
				t.Errorf("expected deleted record in trash list: %s", stdout.String())
			}
		}

		// Restore
		{
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader("master-secret-123\n")
			appCtx := &cli.AppContext{
				VaultPath: vaultPath,
				In:        stdin,
				Out:       &stdout,
				Err:       &stderr,
			}
			rootCmd := cli.NewRootCommand(appCtx)
			rootCmd.SetArgs([]string{"trash", "restore", "Slack Workspace"})
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("trash restore failed: %v", err)
			}
			if !strings.Contains(stdout.String(), "Restored secret 'Slack Workspace' from trash") {
				t.Errorf("unexpected restore output: %s", stdout.String())
			}
		}

		// Verify record is back in active get
		{
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader("master-secret-123\n")
			appCtx := &cli.AppContext{
				VaultPath: vaultPath,
				In:        stdin,
				Out:       &stdout,
				Err:       &stderr,
			}
			rootCmd := cli.NewRootCommand(appCtx)
			rootCmd.SetArgs([]string{"get", "Slack Workspace", "--field", "username", "--raw"})
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("get restored record failed: %v", err)
			}
			if stdout.String() != "bob_senior" {
				t.Errorf("expected updated username 'bob_senior', got '%s'", stdout.String())
			}
		}
	})

	// 5. Permanent deletion / Trash purge
	t.Run("Trash Purge", func(t *testing.T) {
		// Delete again
		{
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader("master-secret-123\n")
			appCtx := &cli.AppContext{
				VaultPath: vaultPath,
				In:        stdin,
				Out:       &stdout,
				Err:       &stderr,
			}
			rootCmd := cli.NewRootCommand(appCtx)
			rootCmd.SetArgs([]string{"delete", "Slack Workspace"})
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("delete failed: %v", err)
			}
		}

		// Purge trash
		{
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader("master-secret-123\n")
			appCtx := &cli.AppContext{
				VaultPath: vaultPath,
				In:        stdin,
				Out:       &stdout,
				Err:       &stderr,
			}
			rootCmd := cli.NewRootCommand(appCtx)
			rootCmd.SetArgs([]string{"trash", "purge"})
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("trash purge failed: %v", err)
			}
			if !strings.Contains(stdout.String(), "All items in trash permanently purged") {
				t.Errorf("unexpected purge output: %s", stdout.String())
			}
		}

		// Trash list should now be empty
		{
			var stdout, stderr bytes.Buffer
			stdin := strings.NewReader("master-secret-123\n")
			appCtx := &cli.AppContext{
				VaultPath: vaultPath,
				In:        stdin,
				Out:       &stdout,
				Err:       &stderr,
			}
			rootCmd := cli.NewRootCommand(appCtx)
			rootCmd.SetArgs([]string{"trash", "list"})
			if err := rootCmd.Execute(); err != nil {
				t.Fatalf("trash list after purge failed: %v", err)
			}
			if !strings.Contains(stdout.String(), "No records found") {
				t.Errorf("expected no records found in trash after purge: %s", stdout.String())
			}
		}
	})
}
