package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
)

// Only the public dump route selects V1; injected/generic inspectors stay deferred.
func runPublicGraph(args []string, reader SourceReader, stdout, stderr io.Writer) int {
	if len(args) != 2 || args[0] != "dump" {
		return runGraph(args, reader, SyntaxSourceParser{}, stdout, stderr)
	}
	filename := args[1]
	deadline := time.Now().Add(commandDeadline)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	source, err := readSourceWithDeadline(reader, filename, remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: read error: %v\n", filename, err)
		return exitFailure
	}
	file, diagnostics, err := parseWithDeadline(EntityFieldsCLIParser{}, filename, string(source), remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: parse error: %v\n", filename, err)
		return exitFailure
	}
	if !reportDiagnostics(diagnostics, stderr) || diagnostics.HasErrors() {
		return exitFailure
	}
	ir, err := bidir.LowerContextWithEntityFieldsSupport(ctx, file, bidir.EntityFieldsV1Support())
	if ctx.Err() != nil {
		err = errCommandDeadline
	}
	if err == nil {
		err = ir.Validate()
	}
	if err != nil {
		reportSemanticDiagnostic(filename, file, err, stderr)
		return exitFailure
	}
	payload, err := marshalFieldGraphDump(source, ir)
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
