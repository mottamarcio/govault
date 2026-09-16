package cli

import (
	"encoding/json"
	"fmt"
	"io"
)

// OutputFormatter manages formatted printing to stdout/stderr in tabular or JSON format.
type OutputFormatter struct {
	out   io.Writer
	err   io.Writer
	json  bool
	quiet bool
}

// NewOutputFormatter creates a new OutputFormatter.
func NewOutputFormatter(out, err io.Writer, jsonOutput, quiet bool) *OutputFormatter {
	return &OutputFormatter{
		out:   out,
		err:   err,
		json:  jsonOutput,
		quiet: quiet,
	}
}

// PrintJSON prints v marshaled as indented JSON to stdout.
func (f *OutputFormatter) PrintJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(f.out, string(data))
	return err
}

// PrintText prints text to stdout unless quiet is enabled.
func (f *OutputFormatter) PrintText(formatStr string, args ...any) {
	if !f.quiet {
		fmt.Fprintf(f.out, formatStr, args...)
	}
}

// PrintError prints error messages to stderr.
func (f *OutputFormatter) PrintError(formatStr string, args ...any) {
	fmt.Fprintf(f.err, "error: "+formatStr+"\n", args...)
}
