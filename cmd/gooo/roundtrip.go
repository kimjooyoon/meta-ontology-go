package main

import (
	"fmt"
	"io"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
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

func evaluateRoundTripWithDeadline(file *syntax.File, timeout time.Duration) (roundTripResult, error) {
	if timeout <= 0 {
		return roundTripResult{}, errCommandDeadline
	}
	type evaluation struct {
		result roundTripResult
		err    error
	}
	result := make(chan evaluation, 1)
	go func() {
		value, err := evaluateRoundTrip(file)
		result <- evaluation{result: value, err: err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case value := <-result:
		return value.result, value.err
	case <-timer.C:
		return roundTripResult{}, errCommandDeadline
	}
}

func evaluateRoundTrip(file *syntax.File) (roundTripResult, error) {
	original, err := bidir.Lower(file)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("semantic lowering: %w", err)
	}
	document, err := bidir.DocumentFromSyntax(file)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("document adaptation: %w", err)
	}
	model, err := bidir.Get(document)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("Get: %w", err)
	}
	written, err := bidir.Put(document, model)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("Put: %w", err)
	}
	roundTripped, err := bidir.LowerDocument(written)
	if err != nil {
		return roundTripResult{}, fmt.Errorf("lower written document: %w", err)
	}
	getPutErr := bidir.CheckGetPut(document)
	putGetErr := bidir.CheckPutGet(document, model)
	return roundTripResult{
		original: original, roundTripped: roundTripped,
		equivalent: bidir.EquivalentAfterRoundTrip(original, roundTripped),
		getPut:     getPutErr == nil, putGet: putGetErr == nil,
		getPutErr: getPutErr, putGetErr: putGetErr,
	}, nil
}

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
