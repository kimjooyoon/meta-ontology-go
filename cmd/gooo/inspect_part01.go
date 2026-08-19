package main

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"time"
)

const graphDumpSchemaVersion = "gooo-graph/v1"
const maxGraphDumpBytes = 1 << 20

var errGraphDumpLimit = errors.New("graph dump resource limit exceeded")

func runInspect(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	return runInspectWithLowerer(args, reader, parser, stdout, stderr, bidir.Lower)
}
func runInspectWithLowerer(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer, lower func(*syntax.File) (semantic.IR, error)) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: gooo inspect <file.gooo>")
		return exitUsage
	}
	filename := args[0]
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
	ir, err := lowerInspectIRWith(file, remainingDeadline(deadline), lower)
	if err != nil {
		if !reportSemanticDiagnostic(filename, file, err, stderr) {
			return exitFailure
		}
		return exitFailure
	}
	payload, err := marshalGraphDump(source, ir)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: graph dump failed: %v\n", filename, err)
		return exitFailure
	}
	if err := writeInspectOutput(stdout, payload, deadline); err != nil {
		fmt.Fprintf(stderr, "gooo: graph output: %v\n", err)
		return exitFailure
	}
	return exitOK
}

type inspectLowerResult struct {
	ir  semantic.IR
	err error
}

func lowerInspectIR(file *syntax.File, timeout time.Duration) (semantic.IR, error) {
	return lowerInspectIRWith(file, timeout, bidir.Lower)
}
