package cli_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/mottamarcio/govault/internal/cli"
)

func TestPromptPassword(t *testing.T) {
	t.Run("Reads single password from stdin buffer", func(t *testing.T) {
		inBuf := bytes.NewBufferString("my-secret-password\n")
		outBuf := new(bytes.Buffer)
		prompter := cli.NewPasswordPrompter(inBuf, outBuf)

		pass, err := prompter.ReadPassword("Enter password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pass != "my-secret-password" {
			t.Fatalf("expected my-secret-password, got: %s", pass)
		}
	})

	t.Run("Reads and confirms matching passwords", func(t *testing.T) {
		inBuf := bytes.NewBufferString("first-pass\nfirst-pass\n")
		outBuf := new(bytes.Buffer)
		prompter := cli.NewPasswordPrompter(inBuf, outBuf)

		pass, err := prompter.ReadPasswordConfirm("Enter new password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pass != "first-pass" {
			t.Fatalf("expected first-pass, got: %s", pass)
		}
	})

	t.Run("Fails on password mismatch", func(t *testing.T) {
		inBuf := bytes.NewBufferString("first-pass\nsecond-mismatch\n")
		outBuf := new(bytes.Buffer)
		prompter := cli.NewPasswordPrompter(inBuf, outBuf)

		_, err := prompter.ReadPasswordConfirm("Enter new password")
		if err != cli.ErrPasswordMismatch {
			t.Fatalf("expected ErrPasswordMismatch, got: %v", err)
		}
	})

	t.Run("Falls back to GOVAULT_PASSWORD environment variable", func(t *testing.T) {
		os.Setenv("GOVAULT_PASSWORD", "env-secret-123")
		defer os.Unsetenv("GOVAULT_PASSWORD")

		inBuf := new(bytes.Buffer)
		outBuf := new(bytes.Buffer)
		prompter := cli.NewPasswordPrompter(inBuf, outBuf)

		pass, err := prompter.ReadPassword("Enter password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pass != "env-secret-123" {
			t.Fatalf("expected env-secret-123, got: %s", pass)
		}
	})

	t.Run("Fails on empty password", func(t *testing.T) {
		inBuf := bytes.NewBufferString("\n")
		outBuf := new(bytes.Buffer)
		prompter := cli.NewPasswordPrompter(inBuf, outBuf)

		_, err := prompter.ReadPassword("Enter password")
		if err != cli.ErrEmptyPassword {
			t.Fatalf("expected ErrEmptyPassword, got: %v", err)
		}
	})
}
