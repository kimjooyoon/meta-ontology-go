package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"time"
)

func runCheckOptions(options checkOptions, jsonMode bool, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	filename := options.filename
	deadline := time.Now().Add(commandDeadline)
	source, err := readSourceWithDeadline(reader, filename, remainingDeadline(deadline))
	if err != nil {
		if jsonMode {
			return reportFailure(true, stdout, stderr, "check", filename, "io.read", err.Error(), syntax.Span{})
		}
		fmt.Fprintf(stderr, "gooo: %s: read error: %v\n", filename, err)
		return exitFailure
	}
	file, diagnostics, err := parseWithDeadline(parser, filename, string(source), remainingDeadline(deadline))
	if err != nil {
		if jsonMode {
			return reportFailure(true, stdout, stderr, "check", filename, "parse", err.Error(), syntax.Span{})
		}
		fmt.Fprintf(stderr, "gooo: %s: parse error: %v\n", filename, err)
		return exitFailure
	}
	if diagnostics.HasErrors() {
		if jsonMode {
			if err := writeJSONReport(stdout, newJSONReport("check", "error", filename, syntaxCLIDiagnostics(diagnostics))); err != nil {
				return exitFailure
			}
			return exitFailure
		}
		if !reportDiagnostics(diagnostics, stderr) {
			return exitFailure
		}
		return exitFailure
	}
	if !jsonMode && !reportDiagnostics(diagnostics, stderr) {
		return exitFailure
	}
	semanticHash, provenanceResponse, code := runSemanticCheck(options, jsonMode, source, file, diagnostics, deadline, stdout, stderr)
	if code != exitOK {
		return code
	}
	return writeCheckResult(jsonMode, filename, semanticHash, provenanceResponse, diagnostics, stdout)
}
