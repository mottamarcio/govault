package cli_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mottamarcio/govault/internal/cli"
)

func TestVaultLifecycle(t *testing.T) {
	tempDir := t.TempDir()
	vaultPath := filepath.Join(tempDir, "vault.db")

	// 1. Status on uninitialized vault
	t.Run("Status on non-existent vault", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			Out:       outBuf,
			Err:       errBuf,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"status"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("status on uninitialized should not error: %v", err)
		}
		if !strings.Contains(outBuf.String(), "Uninitialized") {
			t.Fatalf("expected Uninitialized status, got: %s", outBuf.String())
		}
	})

	// 2. Init vault
	t.Run("Init vault successfully", func(t *testing.T) {
		inBuf := bytes.NewBufferString("MasterSecret123!\nMasterSecret123!\n")
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        inBuf,
			Out:       outBuf,
			Err:       errBuf,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"init"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("init failed: %v", err)
		}
		if !strings.Contains(outBuf.String(), "successfully initialized") {
			t.Fatalf("expected success message, got: %s", outBuf.String())
		}
	})

	// 3. Status on initialized vault (text and JSON)
	t.Run("Status on initialized vault", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			Out:       outBuf,
			Err:       errBuf,
			JSON:      true,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"status", "--json"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("status failed: %v", err)
		}
		if !strings.Contains(outBuf.String(), `"status": "locked"`) {
			t.Fatalf("expected locked JSON status, got: %s", outBuf.String())
		}
	})

	// 4. Unlock vault
	t.Run("Unlock vault", func(t *testing.T) {
		inBuf := bytes.NewBufferString("MasterSecret123!\n")
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        inBuf,
			Out:       outBuf,
			Err:       errBuf,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"unlock"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("unlock failed: %v", err)
		}
		if !strings.Contains(outBuf.String(), "successfully authenticated") {
			t.Fatalf("expected success message, got: %s", outBuf.String())
		}
	})

	// 5. Change password
	t.Run("Change master password with passwd", func(t *testing.T) {
		inBuf := bytes.NewBufferString("MasterSecret123!\nNewMasterSecret456!\nNewMasterSecret456!\n")
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        inBuf,
			Out:       outBuf,
			Err:       errBuf,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"passwd"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("passwd failed: %v", err)
		}
		if !strings.Contains(outBuf.String(), "changed successfully") {
			t.Fatalf("expected success message, got: %s", outBuf.String())
		}
	})

	// 6. Verify new password unlocks
	t.Run("Unlock with new password", func(t *testing.T) {
		inBuf := bytes.NewBufferString("NewMasterSecret456!\n")
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		appCtx := &cli.AppContext{
			VaultPath: vaultPath,
			In:        inBuf,
			Out:       outBuf,
			Err:       errBuf,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"unlock"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("unlock with new password failed: %v", err)
		}
	})
}
