package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

func runClaimDependencies(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	options, err := parseClaimDependencyArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	deadline := time.Now().Add(commandDeadline)
	source, err := readSourceWithDeadline(reader, options.filename, remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: read error: %v\n", options.filename, err)
		return exitFailure
	}
	file, diagnostics, err := parseWithDeadline(parser, options.filename, string(source), remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: parse error: %v\n", options.filename, err)
		return exitFailure
	}
	if !reportDiagnostics(diagnostics, stderr) || diagnostics.HasErrors() {
		return exitFailure
	}
	report := resolveClaimDependencies(options.filename, source, file)
	payload, err := json.Marshal(report)
	if err != nil || len(payload)+1 > maxGraphDumpBytes {
		fmt.Fprintf(stderr, "gooo: %s: claim dependency output failed\n", options.filename)
		return exitFailure
	}
	if err := writeInspectOutput(stdout, append(payload, '\n'), deadline); err != nil {
		fmt.Fprintf(stderr, "gooo: claim dependency output: %v\n", err)
		return exitFailure
	}
	if report.Decision != claimDependencyObserved {
		return exitFailure
	}
	return exitOK
}
