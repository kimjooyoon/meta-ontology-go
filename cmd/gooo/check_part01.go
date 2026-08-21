package main

import (
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
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
	return readRegularFile(filename, maxOutputBytes)
}

// SyntaxSourceParser delegates parsing to the repository syntax contract.
type SyntaxSourceParser struct{}

func (SyntaxSourceParser) ParseFile(filename, source string) (*syntax.File, syntax.Diagnostics) {
	return syntax.ParseFile(filename, source)
}
func runCheck(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	options, err := parseCheckArguments(args)
	if err != nil {
		if jsonMode {
			return reportUsage(true, stdout, stderr, "check", checkUsage)
		}
		fmt.Fprintln(stderr, checkUsage)
		return exitUsage
	}
	return runCheckOptions(options, jsonMode, reader, parser, stdout, stderr)
}
