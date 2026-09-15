package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"time"
)

func readGenerateInput(options generateOptions, reader SourceReader, parser SourceParser, jsonMode bool, stdout, stderr io.Writer, deadline time.Time) (generateInput, int) {
	source, err := readSourceWithDeadline(reader, options.filename, remainingDeadline(deadline))
	if err != nil {
		if jsonMode {
			return generateInput{}, reportFailure(true, stdout, stderr, "generate", options.filename, "io.read", err.Error(), syntax.Span{})
		}
		fmt.Fprintf(stderr, "gooo: %s: read error: %v\n", options.filename, err)
		return generateInput{}, exitFailure
	}
	file, diagnostics, err := parseWithDeadline(parser, options.filename, string(source), remainingDeadline(deadline))
	if err != nil {
		if jsonMode {
			return generateInput{}, reportFailure(true, stdout, stderr, "generate", options.filename, "parse", err.Error(), syntaxFileSpan(file))
		}
		fmt.Fprintf(stderr, "gooo: %s: parse error: %v\n", options.filename, err)
		return generateInput{}, exitFailure
	}
	if diagnostics.HasErrors() {
		if jsonMode {
			if err := writeJSONReport(stdout, newJSONReport("generate", "error", options.filename, syntaxCLIDiagnostics(diagnostics))); err != nil {
				return generateInput{}, exitFailure
			}
		} else if !reportDiagnostics(diagnostics, stderr) {
			return generateInput{}, exitFailure
		}
		return generateInput{}, exitFailure
	}
	if !jsonMode && !reportDiagnostics(diagnostics, stderr) {
		return generateInput{}, exitFailure
	}
	var previousGo []byte
	if options.previousGo != "" {
		previousGo, err = readPreviousWithDeadline(reader, options.previousGo, remainingDeadline(deadline))
		if err != nil {
			if jsonMode {
				return generateInput{}, reportFailure(true, stdout, stderr, "generate", options.previousGo, "io.read-previous-go", err.Error(), syntax.Span{})
			}
			fmt.Fprintf(stderr, "gooo: %s: read previous Go error: %v\n", options.previousGo, err)
			return generateInput{}, exitFailure
		}
	}
	return generateInput{source: source, file: file, diagnostics: diagnostics, previousGo: previousGo}, exitOK
}
