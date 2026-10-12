package main

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestSemanticContextRetainsWholeSourceOwnedAssembly(t *testing.T) {
	source := starterFiles["main.gooo"]
	dump := decodeGraphDump(t, inspectFixtureOutput(t, source))
	root := ""
	for _, node := range dump.Nodes {
		if node.Kind == "Activity" && node.Name == "Clamp" {
			root = node.ID
		}
	}
	if root == "" {
		t.Fatal("starter Clamp declaration is missing from inspect")
	}
	var out, diagnostics bytes.Buffer
	args := []string{"starter.gooo", root, "--depth", "0", "--max-bytes", "16384"}
	code := runSemanticContext(args, fixtureReader{source: source}, SyntaxSourceParser{}, &out, &diagnostics)
	if code != exitOK {
		t.Fatalf("source-owned assembly context failed: %d %s", code, &diagnostics)
	}
	var report semanticContextReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Nodes) != 1 || report.FactLayer != "deterministic" {
		t.Fatalf("unexpected selected view: %s", &out)
	}
	node := report.Nodes[0]
	if node.Assembly == nil || len(node.Assembly.Cases) != 3 || len(node.Assembly.Choices) != 2 ||
		node.Assembly.MaxAttempts != 4 || node.Assembly.Cases[0].Expected != 0 ||
		node.ValueProgram == "" || node.Assembly.Choices[0].Intent == "" {
		t.Fatalf("assembly intent was shortened: %s", &out)
	}
}

func TestSemanticContextExcludesCandidateRelations(t *testing.T) {
	source := sourceWithRuntimeBinding + "\nentity Proposed id \"billing://entity/proposed\"\n"
	raw, ir, err := readSemanticContext("context.gooo", "", fixtureReader{source: source},
		SyntaxSourceParser{}, time.Now().Add(commandDeadline), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	limits := semanticContextLimits{Depth: 8, MaxNodes: 32, MaxBytes: 16384}
	before, err := marshalSemanticContext(raw, ir, "billing://entity/order", limits)
	if err != nil {
		t.Fatal(err)
	}
	fact := semantic.NewCandidateFact("billing://entity/order", semantic.WasDerivedFrom,
		"billing://entity/proposed", "unaccepted analysis")
	if err := ir.Graph.AddCandidate(fact); err != nil {
		t.Fatal(err)
	}
	after, err := marshalSemanticContext(raw, ir, "billing://entity/order", limits)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("candidate fact changed deterministic context", err)
	}
}

func TestSemanticContextPreservesExplicitMultiInputOrder(t *testing.T) {
	source := `package billing
namespace billing
entity First id "billing://entity/z-first"
entity Second id "billing://entity/a-second"
activity Join(First, Second, First) -> Second
`
	var out, diagnostics bytes.Buffer
	args := []string{"inputs.gooo", "billing://activity/join", "--depth", "0"}
	code := runSemanticContext(args, fixtureReader{source: source}, SyntaxSourceParser{}, &out, &diagnostics)
	if code != exitOK {
		t.Fatalf("multi-input context: %d %s", code, &diagnostics)
	}
	var report semanticContextReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	inputs := report.Nodes[0].InputSequence
	if len(inputs) != 3 || inputs[0] != "billing://entity/z-first" ||
		inputs[1] != "billing://entity/a-second" || inputs[2] != inputs[0] {
		t.Fatalf("source order or repeated input was lost: %s", &out)
	}
}
