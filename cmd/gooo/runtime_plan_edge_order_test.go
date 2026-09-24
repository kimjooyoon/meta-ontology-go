package main

import (
	"encoding/json"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/bidir"
	"github.com/kimjooyoon/meta-ontology-go/internal/syntax"
)

func TestBuildRuntimePlanDataPreservesTypedPlanEdgeOrder(t *testing.T) {
	source := []byte(`package runtimebinding
namespace runtimebinding

entity Integer id "gooo://runtime-binding/entity/integer"

activity ProposeCandidate(Integer) -> Integer computes "int.add:1"
activity RecordIndependentReview(Integer) -> Integer computes "int.add:1"
activity CommitCandidate(Integer) -> Integer computes "int.add:1"

bind ProposeCandidate.result -> RecordIndependentReview.input
bind RecordIndependentReview.result -> CommitCandidate.input
`)
	file, diagnostics := syntax.ParseFile("runtime-plan-edge-order.gooo", string(source))
	if file == nil || diagnostics.HasErrors() {
		t.Fatalf("ParseFile() diagnostics = %v", diagnostics)
	}
	ir, err := bidir.Lower(file)
	if err != nil {
		t.Fatalf("bidir.Lower() error = %v", err)
	}
	document, err := bidir.DocumentFromSyntaxWithEntityFieldsSupport(file, syntax.EntityFieldsV1Support())
	if err != nil {
		t.Fatalf("DocumentFromSyntax() error = %v", err)
	}
	typedPlan, err := bidir.CompileTypedPlan(document)
	if err != nil {
		t.Fatalf("CompileTypedPlan() error = %v", err)
	}
	raw, err := buildRuntimePlanDataWithTypedPlan(source, ir, typedPlan)
	if err != nil {
		t.Fatalf("buildRuntimePlanDataWithTypedPlan() error = %v", err)
	}
	var artifact runtimePlanDocument
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatalf("runtime plan JSON error = %v", err)
	}
	if artifact.TypedPlanDigest != typedPlan.Digest() || len(artifact.BindingEdgeOrder) != len(typedPlan.Edges) {
		t.Fatalf("runtime plan typed identity = %#v, want digest %q and %d edges", artifact, typedPlan.Digest(), len(typedPlan.Edges))
	}
	for index, edge := range typedPlan.Edges {
		want := runtimePlanBindingEdgeKey(edge)
		if artifact.BindingEdgeOrder[index] != want {
			t.Fatalf("runtime plan edge[%d] = %q, want %q", index, artifact.BindingEdgeOrder[index], want)
		}
	}
	if artifact.SemanticHash != ir.StableHash() {
		t.Fatalf("semantic hash = %q, want %q", artifact.SemanticHash, ir.StableHash())
	}
}
