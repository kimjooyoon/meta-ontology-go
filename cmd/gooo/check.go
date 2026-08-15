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

func runSemanticCheck(options checkOptions, jsonMode bool, source []byte, file *syntax.File, diagnostics syntax.Diagnostics, deadline time.Time, stdout, stderr io.Writer) (string, *provenancePublishResponse, int) {
	if !options.semantic {
		return "", nil, exitOK
	}
	ir, err := semanticCheckIR(file, remainingDeadline(deadline))
	if err != nil {
		if jsonMode {
			return "", nil, reportFailure(true, stdout, stderr, "check", options.filename, "semantic.lowering", err.Error(), syntaxFileSpan(file))
		}
		if !reportSemanticDiagnostic(options.filename, file, err, stderr) {
			return "", nil, exitFailure
		}
		return "", nil, exitFailure
	}
	semanticHash := ir.StableHash()
	if options.provenanceStore == "" {
		if !jsonMode {
			if _, err := fmt.Fprintln(stderr, deferredCheckProvenance); err != nil {
				return semanticHash, nil, exitFailure
			}
		}
		return semanticHash, nil, exitOK
	}
	response, err := publishSemanticCheckProvenance(options.filename, source, file, ir, options.provenanceStore)
	if err == nil {
		return semanticHash, &response, exitOK
	}
	response, sealErr := rejectSemanticCheckProvenance(response, err)
	if sealErr != nil {
		return semanticHash, nil, exitFailure
	}
	if jsonMode {
		report := newJSONReport("check", "error", options.filename, syntaxCLIDiagnostics(diagnostics))
		report.SemanticHash = semanticHash
		report.Provenance = &response
		if writeErr := writeJSONReport(stdout, report); writeErr != nil {
			return semanticHash, nil, exitFailure
		}
		return semanticHash, nil, exitFailure
	}
	fmt.Fprintf(stdout, "ok: %s\nprovenance: rejected\n", options.filename)
	fmt.Fprintf(stderr, "gooo: %s: %s: %v\n", options.filename, provenanceErrorCode(err), err)
	return semanticHash, nil, exitFailure
}

func writeCheckResult(jsonMode bool, filename, semanticHash string, provenanceResponse *provenancePublishResponse, diagnostics syntax.Diagnostics, stdout io.Writer) int {
	if jsonMode {
		report := newJSONReport("check", "ok", filename, syntaxCLIDiagnostics(diagnostics))
		report.SemanticHash = semanticHash
		report.Provenance = provenanceResponse
		if err := writeJSONReport(stdout, report); err != nil {
			return exitFailure
		}
		return exitOK
	}
	fmt.Fprintf(stdout, "ok: %s\n", filename)
	if provenanceResponse != nil {
		fmt.Fprintf(stdout, "provenance: %s records=%d store_digest=%s\n", provenanceResponse.Status, len(provenanceResponse.Records), provenanceResponse.StoreDigest)
	}
	return exitOK
}

const checkUsage = "usage: gooo check [--semantic] [--provenance-store <ledger.jsonl>] [--json] <file.gooo>"

type checkOptions struct {
	semantic        bool
	filename        string
	provenanceStore string
}

func parseCheckArguments(args []string) (checkOptions, error) {
	options := checkOptions{}
	for index := 0; index < len(args); index++ {
		switch args[index] {
		case "--semantic":
			if options.semantic {
				return checkOptions{}, errors.New(checkUsage)
			}
			options.semantic = true
		case "--provenance-store":
			if options.provenanceStore != "" || index+1 >= len(args) || args[index+1] == "" || strings.HasPrefix(args[index+1], "-") {
				return checkOptions{}, errors.New(checkUsage)
			}
			options.provenanceStore = args[index+1]
			index++
		default:
			if strings.HasPrefix(args[index], "-") || options.filename != "" {
				return checkOptions{}, errors.New(checkUsage)
			}
			options.filename = args[index]
		}
	}
	if options.filename == "" || (options.provenanceStore != "" && !options.semantic) {
		return checkOptions{}, errors.New(checkUsage)
	}
	return options, nil
}

func checkArguments(args []string) (semanticMode bool, filename string, ok bool) {
	options, err := parseCheckArguments(args)
	if err != nil || options.provenanceStore != "" {
		return false, "", false
	}
	return options.semantic, options.filename, true
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
