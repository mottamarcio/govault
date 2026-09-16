package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

var (
	ErrPasswordMismatch = errors.New("passwords do not match")
	ErrEmptyPassword    = errors.New("password cannot be empty")
)

// PasswordPrompter handles secure password input from terminal, stdin, or environment.
type PasswordPrompter struct {
	in     io.Reader
	out    io.Writer
	reader *bufio.Reader
}

// NewPasswordPrompter creates a new PasswordPrompter instance.
func NewPasswordPrompter(in io.Reader, out io.Writer) *PasswordPrompter {
	return &PasswordPrompter{
		in:     in,
		out:    out,
		reader: bufio.NewReader(in),
	}
}

// ReadPassword prompts for a password with echo disabled on terminal, falling back to stdin or environment.
func (p *PasswordPrompter) ReadPassword(prompt string) (string, error) {
	// 1. Check GOVAULT_PASSWORD environment variable
	if envPass := os.Getenv("GOVAULT_PASSWORD"); envPass != "" {
		return envPass, nil
	}

	// 2. Check if stdin is a standard terminal file descriptor
	if f, ok := p.in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		if prompt != "" {
			fmt.Fprint(p.out, prompt+": ")
		}
		bytePass, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(p.out) // Print trailing newline
		if err != nil {
			return "", fmt.Errorf("failed to read password from terminal: %w", err)
		}
		pass := strings.TrimRight(string(bytePass), "\r\n")
		if pass == "" {
			return "", ErrEmptyPassword
		}
		return pass, nil
	}

	// 3. Fallback: Read line from p.in (pipe or buffer)
	if prompt != "" {
		fmt.Fprint(p.out, prompt+": ")
	}
	line, err := p.reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	pass := strings.TrimRight(line, "\r\n")
	if pass == "" {
		return "", ErrEmptyPassword
	}
	return pass, nil
}

// ReadPasswordConfirm prompts for password twice and ensures both match.
func (p *PasswordPrompter) ReadPasswordConfirm(prompt string) (string, error) {
	// If GOVAULT_PASSWORD is set, use directly
	if envPass := os.Getenv("GOVAULT_PASSWORD"); envPass != "" {
		return envPass, nil
	}

	pass1, err := p.ReadPassword(prompt)
	if err != nil {
		return "", err
	}

	pass2, err := p.ReadPassword("Confirm " + strings.ToLower(prompt))
	if err != nil {
		return "", err
	}

	if pass1 != pass2 {
		return "", ErrPasswordMismatch
	}

	return pass1, nil
}

// ReadPrompt reads a non-secret line from terminal/stdin with prompt message.
func (p *PasswordPrompter) ReadPrompt(prompt string) (string, error) {
	if prompt != "" {
		fmt.Fprint(p.out, prompt+": ")
	}
	line, err := p.reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read prompt: %w", err)
	}
	return strings.TrimRight(line, "\r\n"), nil
}

// DefaultPrompter returns a prompter connected to stdin/stderr.
func DefaultPrompter(in io.Reader, out io.Writer) *PasswordPrompter {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stderr
	}
	return NewPasswordPrompter(in, out)
}

// Suppress unused syscall import on linting
var _ = syscall.Stdin
