package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

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
	return os.ReadFile(filename)
}

// SyntaxSourceParser delegates parsing to the repository syntax contract.
type SyntaxSourceParser struct{}

func (SyntaxSourceParser) ParseFile(filename, source string) (*syntax.File, syntax.Diagnostics) {
	return syntax.ParseFile(filename, source)
}

func runCheck(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: gooo check <file.gooo>")
		return exitUsage
	}
	filename := args[0]
	source, err := reader.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(stderr, "gooo: %s: read error: %v\n", filename, err)
		return exitFailure
	}
	_, diagnostics := parser.ParseFile(filename, string(source))
	for _, diagnostic := range diagnostics.SortBySpan() {
		fmt.Fprintln(stderr, diagnostic.String())
	}
	if diagnostics.HasErrors() {
		return exitFailure
	}
	fmt.Fprintf(stdout, "ok: %s\n", filename)
	return exitOK
}
