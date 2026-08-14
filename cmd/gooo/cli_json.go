package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

const diagnosticSchemaVersion = "gooo/diagnostics/v1"

type cliDiagnostic struct {
	Severity string  `json:"severity"`
	Code     string  `json:"code"`
	Message  string  `json:"message"`
	Span     cliSpan `json:"span"`
}

type cliSpan struct {
	File  string      `json:"file"`
	Start cliPosition `json:"start"`
	End   cliPosition `json:"end"`
}

type cliPosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type jsonReport struct {
	SchemaVersion            string          `json:"schema_version"`
	Command                  string          `json:"command"`
	Status                   string          `json:"status"`
	File                     string          `json:"file,omitempty"`
	Output                   string          `json:"output,omitempty"`
	Manifest                 string          `json:"manifest,omitempty"`
	PreviousGo               string          `json:"previous_go,omitempty"`
	ProtectedBytesEqual      *bool           `json:"protected_bytes_equal,omitempty"`
	SemanticHash             string          `json:"semantic_hash,omitempty"`
	OriginalSemanticHash     string          `json:"original_semantic_hash,omitempty"`
	RoundTrippedSemanticHash string          `json:"round_tripped_semantic_hash,omitempty"`
	Equivalent               *bool           `json:"equivalent,omitempty"`
	GetPut                   *bool           `json:"get_put,omitempty"`
	PutGet                   *bool           `json:"put_get,omitempty"`
	Diagnostics              []cliDiagnostic `json:"diagnostics"`

	Provenance *provenancePublishResponse `json:"provenance,omitempty"`
}

func parseJSONFlag(args []string) (clean []string, jsonMode bool) {
	clean = make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--json" {
			jsonMode = true
			continue
		}
		clean = append(clean, arg)
	}
	return clean, jsonMode
}

func newJSONReport(command, status, filename string, diagnostics []cliDiagnostic) jsonReport {
	if diagnostics == nil {
		diagnostics = []cliDiagnostic{}
	}
	return jsonReport{
		SchemaVersion: diagnosticSchemaVersion,
		Command:       command,
		Status:        status,
		File:          filename,
		Diagnostics:   diagnostics,
	}
}

func writeJSONReport(writer io.Writer, report jsonReport) error {
	payload, err := json.Marshal(report)
	if err != nil {
		return err
	}
	if len(payload)+1 > maxDiagnosticBytes {
		return errDiagnosticLimit
	}
	_, err = fmt.Fprintf(writer, "%s\n", payload)
	return err
}

func sortedSyntaxDiagnostics(diagnostics syntax.Diagnostics) syntax.Diagnostics {
	return canonicalDiagnostics(diagnostics)
}

func syntaxCLIDiagnostics(diagnostics syntax.Diagnostics) []cliDiagnostic {
	sorted := sortedSyntaxDiagnostics(diagnostics)
	result := make([]cliDiagnostic, 0, len(sorted))
	for _, diagnostic := range sorted {
		result = append(result, cliDiagnostic{
			Severity: diagnostic.Severity.String(),
			Code:     string(diagnostic.Code),
			Message:  diagnostic.Message,
			Span:     cliSpanFromSyntax(diagnostic.Span),
		})
	}
	return result
}

func cliSpanFromSyntax(span syntax.Span) cliSpan {
	return cliSpan{
		File:  span.Filename,
		Start: cliPosition{Offset: span.Start.Offset, Line: span.Start.Line, Column: span.Start.Column},
		End:   cliPosition{Offset: span.End.Offset, Line: span.End.Line, Column: span.End.Column},
	}
}

func syntaxFileSpan(file *syntax.File) syntax.Span {
	if file == nil {
		return syntax.Span{}
	}
	return file.Span
}

func errorCLIDiagnostic(filename, code, message string, span syntax.Span) cliDiagnostic {
	if span.Filename == "" {
		span.Filename = filename
	}
	return cliDiagnostic{Severity: "error", Code: code, Message: message, Span: cliSpanFromSyntax(span)}
}

func printSyntaxDiagnostics(stderr io.Writer, diagnostics syntax.Diagnostics) error {
	for _, diagnostic := range sortedSyntaxDiagnostics(diagnostics) {
		if _, err := fmt.Fprintln(stderr, diagnostic.String()); err != nil {
			return err
		}
	}
	return nil
}

func reportFailure(jsonMode bool, stdout, stderr io.Writer, command, filename, code, message string, span syntax.Span) int {
	diagnostic := errorCLIDiagnostic(filename, code, message, span)
	if jsonMode {
		if err := writeJSONReport(stdout, newJSONReport(command, "error", filename, []cliDiagnostic{diagnostic})); err != nil {
			return exitFailure
		}
		return exitFailure
	}
	if filename != "" {
		fmt.Fprintf(stderr, "gooo: %s: %s: %s\n", filename, code, message)
	} else {
		fmt.Fprintf(stderr, "gooo: %s: %s\n", code, message)
	}
	return exitFailure
}

func reportUsage(jsonMode bool, stdout, stderr io.Writer, command, usage string) int {
	if jsonMode {
		report := newJSONReport(command, "error", "", []cliDiagnostic{{
			Severity: "error",
			Code:     "cli.usage",
			Message:  usage,
			Span:     cliSpanFromSyntax(syntax.Span{}),
		}})
		if err := writeJSONReport(stdout, report); err != nil {
			return exitFailure
		}
		return exitUsage
	}
	fmt.Fprintln(stderr, usage)
	return exitUsage
}
