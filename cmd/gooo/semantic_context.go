package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/assemblyspec"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const semanticContextUsage = "usage: gooo context <file.gooo> <stable-id> [--depth 0..8] " +
	"[--max-nodes 1..256] [--max-bytes 1024..1048576] [--expect-source <digest>]"

type semanticContextLimits struct {
	Depth    int `json:"depth"`
	MaxNodes int `json:"max_nodes"`
	MaxBytes int `json:"max_bytes"`
}

type semanticContextNode struct {
	graphNode
	Depth         int                `json:"depth"`
	ValueProgram  string             `json:"value_program,omitempty"`
	InputSequence []semantic.ID      `json:"input_sequence,omitempty"`
	Assembly      *assemblyspec.Spec `json:"assembly,omitempty"`
}

type semanticContextReport struct {
	Schema               string                `json:"schema"`
	SourceDigest         string                `json:"source_digest"`
	SemanticDigest       string                `json:"semantic_digest"`
	Root                 string                `json:"root"`
	FactLayer            string                `json:"fact_layer"`
	Limits               semanticContextLimits `json:"limits"`
	Nodes                []semanticContextNode `json:"nodes"`
	Relations            []graphRelation       `json:"relations"`
	RuntimeBindings      []graphRuntimeBinding `json:"runtime_bindings"`
	Truncated            bool                  `json:"truncated"`
	Reasons              []string              `json:"reasons"`
	OmittedAdjacentNodes int                   `json:"omitted_adjacent_nodes"`
	NextRoot             string                `json:"next_root,omitempty"`
	ContextDigest        string                `json:"context_digest,omitempty"`
}

func runSemanticContext(args []string, reader SourceReader, parser SourceParser, stdout, stderr io.Writer) int {
	limits, expected, err := parseSemanticContextOptions(args)
	if err != nil {
		fmt.Fprintln(stderr, "gooo context:", err)
		return exitUsage
	}
	deadline := time.Now().Add(commandDeadline)
	source, ir, err := readSemanticContext(args[0], expected, reader, parser, deadline, stderr)
	if err == nil {
		var payload []byte
		payload, err = marshalSemanticContext(source, ir, args[1], limits)
		if err == nil {
			err = writeInspectOutput(stdout, payload, deadline)
		}
	}
	if err != nil {
		fmt.Fprintln(stderr, "gooo context:", err)
		return exitFailure
	}
	return exitOK
}

func marshalSemanticContext(source []byte, ir semantic.IR, root string, limits semanticContextLimits) ([]byte, error) {
	dump := newGraphDump(source, ir)
	dump.Relations = graphRelations(ir.Graph.DeterministicFacts())
	index := newSemanticContextIndex(dump, ir.Graph.Nodes())
	if _, found := index.nodes[root]; !found {
		return nil, fmt.Errorf("stable ID %q is absent; inspect declarations with gooo inspect", root)
	}
	order, depths := index.walk(root, limits)
	for count := len(order); count > 0; count-- {
		report := index.project(dump, root, limits, order[:count], depths)
		canonical, err := json.Marshal(report)
		if err != nil {
			return nil, err
		}
		report.ContextDigest = semantic.StableHash(canonical)
		payload, err := json.Marshal(report)
		if err != nil {
			return nil, err
		}
		if len(payload)+1 <= limits.MaxBytes {
			return append(payload, '\n'), nil
		}
	}
	if limits.MaxBytes == maxGraphDumpBytes {
		return nil, fmt.Errorf("root context exceeds the supported %d-byte output limit", maxGraphDumpBytes)
	}
	return nil, fmt.Errorf("root context exceeds --max-bytes %d; increase the byte limit", limits.MaxBytes)
}
