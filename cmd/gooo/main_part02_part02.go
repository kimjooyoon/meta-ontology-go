package main

import (
	"errors"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"io"
	"time"
)

func runAnalyzeDelta(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	o, err := parseAnalyzeDeltaArguments(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitUsage
	}
	deadline := time.Now().Add(2 * commandDeadline)
	authority, model, err := readAnalyzeAuthority(o.authority, reader, parser, deadline)
	if err != nil {
		return reportAnalyzeDeltaError(stderr, o.authority, "read authority", err)
	}
	sources, err := readAnalyzeSources(o.goFiles, reader, model, authority, deadline)
	if err != nil {
		return reportAnalyzeDeltaError(stderr, "", "read Go input", err)
	}
	registry, err := analyzeRegistry(authority, sources)
	if err != nil {
		return reportAnalyzeDeltaError(stderr, "", "build symbol registry", err)
	}
	policy, err := analyzeMappingPolicy()
	if err != nil {
		return reportAnalyzeDeltaError(stderr, "", "build mapping policy", err)
	}
	adapted, err := analyzer.AnalyzeAndAdaptSemantic(analyzer.SourceSemanticAdapterInput{Base: authority, Sources: sources, Registry: registry, Policy: policy, Producer: semantic.GoHostedCompilerID, EvidenceKind: semantic.CompilerRunEvidence, ToolchainIdentity: analyzeDeltaToolchain})
	if err != nil {
		return reportAnalyzeDeltaError(stderr, "", "adapt semantic delta", err)
	}
	if !semantic.CompareIR(adapted.IR, authority).SemanticEqual {
		return reportAnalyzeDeltaError(stderr, "", "reconcile signature facts", errors.New("deterministic signature facts disagree with DSL authority"))
	}
	payload, err := marshalAnalyzeDelta(adapted.NormalizedDelta, authority, adapted.IR)
	if err != nil {
		return reportAnalyzeDeltaError(stderr, "", "marshal semantic delta", err)
	}
	if err := writeInspectOutput(stdout, payload, deadline); err != nil {
		return reportAnalyzeDeltaError(stderr, "", "write semantic delta", err)
	}
	return exitOK
}
