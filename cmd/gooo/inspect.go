package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const graphDumpSchemaVersion = "gooo-graph/v1"

func runInspect(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
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
	ir, err := lowerInspectIR(file, remainingDeadline(deadline))
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: semantic lowering failed: %v\n", filename, err)
		return exitFailure
	}
	payload, err := marshalGraphDump(source, ir)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: graph dump failed: %v\n", filename, err)
		return exitFailure
	}
	if err := writeInspectOutput(stdout, payload); err != nil {
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
	if timeout <= 0 {
		return semantic.IR{}, errCommandDeadline
	}
	result := make(chan inspectLowerResult, 1)
	go func() {
		ir, err := bidir.Lower(file)
		result <- inspectLowerResult{ir: ir, err: err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case lowered := <-result:
		return lowered.ir, lowered.err
	case <-timer.C:
		return semantic.IR{}, errCommandDeadline
	}
}

func marshalGraphDump(source []byte, ir semantic.IR) ([]byte, error) {
	dump := newGraphDump(source, ir)
	payload, err := json.Marshal(dump)
	if err != nil {
		return nil, err
	}
	return append(payload, '\n'), nil
}

func writeInspectOutput(output io.Writer, payload []byte) error {
	for len(payload) > 0 {
		written, err := output.Write(payload)
		if written > 0 {
			payload = payload[written:]
		}
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
	}
	return nil
}
