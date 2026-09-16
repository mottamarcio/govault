package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mottamarcio/govault/internal/cli"
	"github.com/mottamarcio/govault/internal/service"
)

func TestGenerateCommand(t *testing.T) {
	t.Run("Generate Random Password with Raw Output", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		appCtx := &cli.AppContext{
			Out: &stdout,
			Err: &stderr,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"generate", "--length", "32", "--raw"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("generate failed: %v", err)
		}

		pass := stdout.String()
		if len(pass) != 32 {
			t.Errorf("expected 32 characters, got %d ('%s')", len(pass), pass)
		}
	})

	t.Run("Generate Passphrase with JSON Output", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		appCtx := &cli.AppContext{
			Out:  &stdout,
			Err:  &stderr,
			JSON: true,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"generate", "--passphrase", "--words", "4", "--delimiter", ".", "--capitalize"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("generate passphrase failed: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, `"type": "passphrase"`) || !strings.Contains(outStr, `"word_count": 4`) {
			t.Errorf("unexpected json output: %s", outStr)
		}
	})

	t.Run("Generate with Invalid Length Fails", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		appCtx := &cli.AppContext{
			Out: &stdout,
			Err: &stderr,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"generate", "--length", "4"})
		err := rootCmd.Execute()
		if err == nil {
			t.Fatal("expected error for length < 8, got nil")
		}
	})
}

func TestCopyCommand(t *testing.T) {
	vaultPath, cleanup := setupTestVault(t, "master-secret-123")
	defer cleanup()

	// Seed record
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
			"--title", "Email Account",
			"--username", "user@example.com",
			"--generate",
			"--length", "20",
		})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("failed to seed record: %v", err)
		}
	}

	t.Run("Copy Password to Clipboard with Auto-Clear", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		clipDriver := service.NewInMemoryClipboardDriver()
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
			Clipboard: clipDriver,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"copy", "Email Account", "--clear-after", "100ms"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("copy command failed: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "Copied to clipboard. Will auto-clear in 100ms") {
			t.Errorf("unexpected copy output: %s", outStr)
		}

		// Verify clipboard contains secret immediately
		copiedText, err := clipDriver.ReadText()
		if err != nil || copiedText == "" {
			t.Fatalf("clipboard should contain copied secret, got err=%v text='%s'", err, copiedText)
		}

		// Wait for auto-clear
		time.Sleep(150 * time.Millisecond)
		clearedText, _ := clipDriver.ReadText()
		if clearedText != "" {
			t.Errorf("clipboard should be cleared after timeout, got '%s'", clearedText)
		}
	})
}

func TestBackupRestoreInspectCommands(t *testing.T) {
	vaultPath, cleanup := setupTestVault(t, "master-secret-123")
	defer cleanup()

	tempDir := filepath.Dir(vaultPath)
	backupPath := filepath.Join(tempDir, "test_backup.gvault")

	// Seed record in active vault
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
			"--title", "Production DB",
			"--username", "dbadmin",
			"--generate",
		})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("failed to seed record: %v", err)
		}
	}

	// 1. Test Backup Export
	t.Run("Create Backup", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"backup", backupPath})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("backup failed: %v", err)
		}

		if _, err := os.Stat(backupPath); err != nil {
			t.Fatalf("backup file was not created: %v", err)
		}
	})

	// 2. Test Inspect (Header only and Authenticated Manifest)
	t.Run("Inspect Backup Header and Manifest", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"inspect", backupPath, "--authenticated"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("inspect failed: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, "Backup ID:") || !strings.Contains(outStr, "Entries Count:   1") {
			t.Errorf("unexpected inspect output: %s", outStr)
		}
	})

	// 3. Test Restore
	t.Run("Restore Backup Atomically", func(t *testing.T) {
		// Wipe or corrupt database first
		os.Remove(vaultPath)

		var stdout, stderr bytes.Buffer
		stdin := strings.NewReader("master-secret-123\n")
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        stdin,
			Out:       &stdout,
			Err:       &stderr,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"restore", backupPath, "--skip-snapshot"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("restore failed: %v", err)
		}

		// Verify record is restored in database
		{
			var outBuf, errBuf bytes.Buffer
			inReader := strings.NewReader("master-secret-123\n")
			checkCtx := &cli.AppContext{
				VaultPath: vaultPath,
				In:        inReader,
				Out:       &outBuf,
				Err:       &errBuf,
			}
			cmd := cli.NewRootCommand(checkCtx)
			cmd.SetArgs([]string{"get", "Production DB", "--field", "username", "--raw"})
			if err := cmd.Execute(); err != nil {
				t.Fatalf("get restored record failed: %v", err)
			}
			if outBuf.String() != "dbadmin" {
				t.Errorf("expected 'dbadmin', got '%s'", outBuf.String())
			}
		}
	})
}

func TestDoctorCommand(t *testing.T) {
	vaultPath, cleanup := setupTestVault(t, "master-secret-123")
	defer cleanup()

	t.Run("Doctor Diagnostics on Healthy Vault", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			Out:       &stdout,
			Err:       &stderr,
			JSON:      true,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"doctor"})
		if err := rootCmd.Execute(); err != nil {
			t.Fatalf("doctor failed on healthy vault: %v", err)
		}

		outStr := stdout.String()
		if !strings.Contains(outStr, `"healthy": true`) {
			t.Errorf("expected healthy true in json output: %s", outStr)
		}
	})
}
