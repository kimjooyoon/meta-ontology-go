package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const deferredCheckProvenance = "gooo: provenance: deferred (no provenance records attached)"

// SourceReader isolates command I/O from the check pipeline. A future
// workspace or editor host can provide a reader without changing diagnostics.
type SourceReader interface {
	ReadFile(string) ([]byte, error)
}

// SourceParser is the seam for a later semantic lowering stage. Check only
// needs a syntax tree today; callers can replace this adapter when lowering is
// available without changing command or exit-code behavior.
type SourceParser interface {
	ParseFile(string, string) (*syntax.File, syntax.Diagnostics)
}

// OSFileReader reads source files from the local filesystem.
type OSFileReader struct{}

func (OSFileReader) ReadFile(filename string) ([]byte, error) {
	return readRegularFile(filename, maxInputBytes)
}

// SyntaxSourceParser delegates parsing to the repository syntax contract.
type SyntaxSourceParser struct{}

func (SyntaxSourceParser) ParseFile(filename, source string) (*syntax.File, syntax.Diagnostics) {
	return syntax.ParseFile(filename, source)
}

func runCheck(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	semanticMode, filename, ok := checkArguments(args)
	if !ok {
		fmt.Fprintln(stderr, "usage: gooo check [--semantic] <file.gooo>")
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
	if !reportDiagnostics(diagnostics, stderr) {
		return exitFailure
	}
	if diagnostics.HasErrors() {
		return exitFailure
	}
	if semanticMode {
		if _, err := semanticCheckIR(file, remainingDeadline(deadline)); err != nil {
			if !reportSemanticDiagnostic(filename, file, err, stderr) {
				return exitFailure
			}
			return exitFailure
		}
		if _, err := fmt.Fprintln(stderr, deferredCheckProvenance); err != nil {
			return exitFailure
		}
	}
	fmt.Fprintf(stdout, "ok: %s\n", filename)
	return exitOK
}

func checkArguments(args []string) (semanticMode bool, filename string, ok bool) {
	if len(args) == 1 {
		return false, args[0], true
	}
	if len(args) == 2 && args[0] == "--semantic" {
		return true, args[1], true
	}
	return false, "", false
}

func reportSemanticDiagnostic(filename string, file *syntax.File, err error, stderr io.Writer) bool {
	span := syntax.Span{Filename: filename}
	if file != nil {
		span = file.Span
	}
	_, writeErr := fmt.Fprintf(stderr, "%s: error %s: %v\n", span.String(), semanticDiagnosticCode(err), err)
	return writeErr == nil
}

func semanticDiagnosticCode(err error) string {
	if errors.Is(err, errCommandDeadline) {
		return "semantic.deadline"
	}
	if strings.Contains(err.Error(), "unknown declaration") {
		return "semantic.invalid-endpoint"
	}
	if errors.Is(err, semantic.ErrUnknownRelation) {
		return "semantic.invalid-relation"
	}
	if strings.Contains(err.Error(), "cannot connect") || errors.Is(err, semantic.ErrInvalidFact) {
		return "semantic.invalid-kind"
	}
	return "semantic.invalid"
}
