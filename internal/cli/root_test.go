package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mottamarcio/govault/internal/cli"
)

func TestRootCommandVersionAndFlags(t *testing.T) {
	t.Run("Version command text output", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		appCtx := &cli.AppContext{
			Out: outBuf,
			Err: errBuf,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"version"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		if !strings.Contains(outBuf.String(), "GoVault v1.0.0") {
			t.Fatalf("expected version string, got: %s", outBuf.String())
		}
	})

	t.Run("Version command JSON output", func(t *testing.T) {
		outBuf := new(bytes.Buffer)
		errBuf := new(bytes.Buffer)
		appCtx := &cli.AppContext{
			Out:  outBuf,
			Err:  errBuf,
			JSON: true,
		}

		rootCmd := cli.NewRootCommand(appCtx)
		rootCmd.SetArgs([]string{"version", "--json"})
		err := rootCmd.Execute()
		if err != nil {
			t.Fatalf("unexpected execution error: %v", err)
		}

		if !strings.Contains(outBuf.String(), `"version": "1.0.0"`) {
			t.Fatalf("expected JSON version, got: %s", outBuf.String())
		}
	})

	t.Run("Default vault path resolution", func(t *testing.T) {
		path := cli.DefaultVaultPath()
		if !strings.Contains(path, "vault.db") {
			t.Fatalf("unexpected default vault path: %s", path)
		}
	})
}
