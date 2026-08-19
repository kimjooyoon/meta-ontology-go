package main

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
	"io"
	"time"
)

type queryOptions struct {
	Kind        semantic.Kind
	KindSet     bool
	IDSelector  bool
	operation   string
	root        string
	target      string
	relation    string
	rule        string
	layer       string
	direction   string
	maxDepth    int
	maxDepthSet bool
	limit       int
	limitSet    bool
}

func runQuery(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	args, jsonMode := parseJSONFlag(args)
	options, filename, usage := parseQueryArguments(args)
	if usage != "" {
		return reportUsage(jsonMode, stdout, stderr, "query", usage)
	}
	deadline := time.Now().Add(commandDeadline)
	source, err := readSourceWithDeadline(reader, filename, remainingDeadline(deadline))
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "io.read", err.Error(), syntax.Span{})
	}
	file, diagnostics, err := parseWithDeadline(parser, filename, string(source), remainingDeadline(deadline))
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "parse", err.Error(), syntaxFileSpan(file))
	}
	if diagnostics.HasErrors() {
		if jsonMode {
			if err := writeJSONReport(stdout, newJSONReport("query", "error", filename, syntaxCLIDiagnostics(diagnostics))); err != nil {
				return exitFailure
			}
			return exitFailure
		}
		if err := printSyntaxDiagnostics(stderr, diagnostics); err != nil {
			return exitFailure
		}
		return exitFailure
	}
	if !jsonMode {
		if err := printSyntaxDiagnostics(stderr, diagnostics); err != nil {
			return exitFailure
		}
	}
	ir, err := lowerInspectIRWith(file, remainingDeadline(deadline), bidir.Lower)
	if err != nil {
		return reportFailure(jsonMode, stdout, stderr, "query", filename, "semantic.lowering", err.Error(), syntaxFileSpan(file))
	}
	return runQueryEngine(options, ir, filename, jsonMode, stdout, stderr)
}
