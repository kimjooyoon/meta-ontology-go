package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
)

func reportRoundTripMismatch(filename string, file *syntax.File, result roundTripResult, jsonMode bool, stdout, stderr io.Writer) int {
	message := "semantic round-trip is not equivalent"
	if result.getPutErr != nil {
		message = result.getPutErr.Error()
	} else if result.putGetErr != nil {
		message = result.putGetErr.Error()
	}
	if !jsonMode {
		fmt.Fprintf(stderr, "gooo: %s: round-trip failed: %s\n", filename, message)
		return exitFailure
	}
	report := newJSONReport("roundtrip", "error", filename, []cliDiagnostic{
		errorCLIDiagnostic(filename, "bidir.roundtrip", message, syntaxFileSpan(file)),
	})
	report.OriginalSemanticHash = result.original.StableHash()
	report.RoundTrippedSemanticHash = result.roundTripped.StableHash()
	report.Equivalent = &result.equivalent
	if err := writeJSONReport(stdout, report); err != nil {
		return exitFailure
	}
	return exitFailure
}
