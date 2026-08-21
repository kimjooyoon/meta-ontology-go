package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"time"
)

func runRoundTrip(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	if len(args) != 1 {
		return reportUsage(jsonMode, stdout, stderr, "roundtrip", "usage: gooo roundtrip [--json] <file.gooo>")
	}
	filename := args[0]
	deadline := time.Now().Add(commandDeadline)
	source, err := readSourceWithDeadline(reader, filename, remainingDeadline(deadline))
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "roundtrip", filename, "io.read", err.Error(), syntax.Span{})
	}
	file, diagnostics, err := parseWithDeadline(parser, filename, string(source), remainingDeadline(deadline))
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "roundtrip", filename, "parse", err.Error(), syntaxFileSpan(file))
	}
	if diagnostics.HasErrors() {
		if jsonMode {
			if err := writeJSONReport(stdout, newJSONReport("roundtrip", "error", filename, syntaxCLIDiagnostics(diagnostics))); err != nil {
				return exitFailure
			}
			return exitFailure
		}
		if err := printSyntaxDiagnostics(stderr, diagnostics); err != nil {
			return exitFailure
		}
		return exitFailure
	}
	if !jsonMode {
		if err := printSyntaxDiagnostics(stderr, diagnostics); err != nil {
			return exitFailure
		}
	}
	result, err := evaluateRoundTripWithDeadline(file, remainingDeadline(deadline))
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "roundtrip", filename, "bidir.roundtrip", err.Error(), syntaxFileSpan(file))
	}
	if !result.equivalent || !result.getPut || !result.putGet {
		return reportRoundTripMismatch(filename, file, result, jsonMode, stdout, stderr)
	}
	if jsonMode {
		report := newJSONReport("roundtrip", "ok", filename, syntaxCLIDiagnostics(diagnostics))
		report.OriginalSemanticHash = result.original.StableHash()
		report.RoundTrippedSemanticHash = result.roundTripped.StableHash()
		report.Equivalent = &result.equivalent
		report.GetPut, report.PutGet = &result.getPut, &result.putGet
		if err := writeJSONReport(stdout, report); err != nil {
			return exitFailure
		}
		return exitOK
	}
	fmt.Fprintf(stdout, "roundtrip: %s equivalent (%s)\n", filename, result.original.StableHash())
	return exitOK
}

type roundTripResult struct {
	original, roundTripped     semantic.IR
	equivalent, getPut, putGet bool
	getPutErr, putGetErr       error
}
