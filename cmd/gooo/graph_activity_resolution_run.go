package main

import (
	"encoding/json"
	"fmt"
	queryengine "github.com/kimjooyoon/meta-ontology-go/internal/query"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"io"
	"time"
)

func runGraphActivityResolution(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	filename, selector, err := parseGraphActivityResolutionArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	deadline := time.Now().Add(commandDeadline)
	source, err := readSourceWithDeadline(reader, filename, remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: read error: %v\n", filename, err)
		return exitFailure
	}
	file, diagnostics, err := parseWithDeadline(parser, filename, string(source), remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: parse error: %v\n", filename, err)
		return exitFailure
	}
	if !reportDiagnostics(diagnostics, stderr) || diagnostics.HasErrors() {
		return exitFailure
	}
	ir, err := lowerInspectIR(file, remainingDeadline(deadline))
	if err != nil {
		reportSemanticDiagnostic(filename, file, err, stderr)
		return exitFailure
	}
	graph, err := queryengine.FromSemanticIR(ir)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: activity graph failed: %v\n", filename, err)
		return exitFailure
	}
	resolution, err := graph.ResolveActivityCardinality(selector)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: activity resolution failed: %v\n", filename, err)
		return exitFailure
	}
	resolution.Subject.SourceDigest = semantic.StableHash(source)
	resolution.Subject.SourceStatus = "bound"
	resolution.Subject.SourceFile = filename
	payload, err := json.Marshal(resolution)
	if err != nil || len(payload)+1 > maxGraphDumpBytes {
		fmt.Fprintf(stderr, "gooo: %s: activity resolution output failed\n", filename)
		return exitFailure
	}
	if err := writeInspectOutput(stdout, append(payload, '\n'), deadline); err != nil {
		fmt.Fprintf(stderr, "gooo: activity resolution output: %v\n", err)
		return exitFailure
	}
	if resolution.Decision != queryengine.ActivityResolutionClosed {
		return exitFailure
	}
	return exitOK
}
