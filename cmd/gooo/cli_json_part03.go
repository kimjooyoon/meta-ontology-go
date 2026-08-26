package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
)

func reportFailure(jsonMode bool, stdout, stderr io.Writer, command, filename, code, message string, span syntax.Span) int {
	diagnostic := errorCLIDiagnostic(filename, code, message, span)
	if jsonMode {
		if err := writeJSONReport(stdout, newJSONReport(command, "error", filename, []cliDiagnostic{diagnostic})); err != nil {
			return exitFailure
		}
		return exitFailure
	}
	if filename != "" {
		fmt.Fprintf(stderr, "gooo: %s: %s: %s\n", filename, code, message)
	} else {
		fmt.Fprintf(stderr, "gooo: %s: %s\n", code, message)
	}
	return exitFailure
}
func reportUsage(jsonMode bool, stdout, stderr io.Writer, command, usage string) int {
	if jsonMode {
		report := newJSONReport(command, "error", "", []cliDiagnostic{{
			Severity: "error",
			Code:     "cli.usage",
			Message:  usage,
			Span:     cliSpanFromSyntax(syntax.Span{}),
		}})
		if err := writeJSONReport(stdout, report); err != nil {
			return exitFailure
		}
		return exitUsage
	}
	fmt.Fprintln(stderr, usage)
	return exitUsage
}
