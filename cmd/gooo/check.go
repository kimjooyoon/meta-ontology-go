package main

import (
	"errors"
	"fmt"
	"io"
	"sort"
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
	return readRegularFile(filename, maxOutputBytes)
}

// SyntaxSourceParser delegates parsing to the repository syntax contract.
type SyntaxSourceParser struct{}

func (SyntaxSourceParser) ParseFile(filename, source string) (*syntax.File, syntax.Diagnostics) {
	return syntax.ParseFile(filename, source)
}

func runCheck(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	semanticMode, filename, ok := checkArguments(args)
	if !ok {
		if jsonMode {
			return reportUsage(true, stdout, stderr, "check", "usage: gooo check [--semantic] [--json] <file.gooo>")
		}
		fmt.Fprintln(stderr, "usage: gooo check [--semantic] <file.gooo>")
		return exitUsage
	}
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
	var semanticHash string
	if semanticMode {
		ir, err := semanticCheckIR(file, remainingDeadline(deadline))
		if err != nil {
			if jsonMode {
				return reportFailure(true, stdout, stderr, "check", filename, "semantic.lowering", err.Error(), syntaxFileSpan(file))
			}
			if !reportSemanticDiagnostic(filename, file, err, stderr) {
				return exitFailure
			}
			return exitFailure
		}
		semanticHash = ir.StableHash()
		if !jsonMode {
			if _, err := fmt.Fprintln(stderr, deferredCheckProvenance); err != nil {
				return exitFailure
			}
		}
	}
	if jsonMode {
		report := newJSONReport("check", "ok", filename, syntaxCLIDiagnostics(diagnostics))
		report.SemanticHash = semanticHash
		if err := writeJSONReport(stdout, report); err != nil {
			return exitFailure
		}
		return exitOK
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
	output, formatErr := formatSemanticDiagnostics(span, err)
	if formatErr != nil {
		fmt.Fprintln(stderr, "gooo: diagnostic resource limit exceeded")
		return false
	}
	_, writeErr := stderr.Write(output)
	return writeErr == nil
}

type semanticDiagnostic struct {
	Code    string
	Message string
}

func canonicalSemanticDiagnostics(err error) []semanticDiagnostic {
	var validation *semantic.ValidationErrors
	if errors.As(err, &validation) && len(validation.Issues) > 0 {
		issues := append([]semantic.ValidationIssue(nil), validation.Issues...)
		sort.Slice(issues, func(i, j int) bool {
			left, right := issues[i], issues[j]
			if left.Code != right.Code {
				return left.Code < right.Code
			}
			if left.Message != right.Message {
				return left.Message < right.Message
			}
			if left.Subject != right.Subject {
				return left.Subject < right.Subject
			}
			return left.Object < right.Object
		})
		result := make([]semanticDiagnostic, 0, len(issues))
		for _, issue := range issues {
			result = append(result, semanticDiagnostic{
				Code: semanticValidationCode(issue.Code), Message: issue.Message,
			})
		}
		return result
	}
	return []semanticDiagnostic{{Code: semanticDiagnosticCode(err), Message: err.Error()}}
}

func formatSemanticDiagnostics(span syntax.Span, err error) ([]byte, error) {
	diagnostics := canonicalSemanticDiagnostics(err)
	if len(diagnostics) > maxDiagnosticCount {
		return nil, errDiagnosticLimit
	}
	var output strings.Builder
	for _, diagnostic := range diagnostics {
		line := fmt.Sprintf("%s: error %s: %s\n", span.String(), diagnostic.Code, diagnostic.Message)
		if output.Len()+len(line) > maxDiagnosticBytes {
			return nil, errDiagnosticLimit
		}
		output.WriteString(line)
	}
	return []byte(output.String()), nil
}

func semanticValidationCode(code string) string {
	code = strings.TrimSpace(code)
	if code == "" {
		return "semantic.invalid"
	}
	if strings.HasPrefix(code, "semantic.") {
		return code
	}
	switch code {
	case "relation-kind":
		return "semantic.invalid-kind"
	case "missing-subject", "missing-object":
		return "semantic.invalid-endpoint"
	case "unknown-relation":
		return "semantic.invalid-relation"
	}
	return "semantic." + code
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
