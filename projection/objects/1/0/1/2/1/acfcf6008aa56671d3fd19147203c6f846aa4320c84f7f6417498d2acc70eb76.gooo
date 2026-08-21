package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRunAnalyzeSemanticInvalidInputIsDeferred(t *testing.T) {
	const semanticInvalid = `package billing
namespace billing
entity Order id "billing://entity/order"
activity PayOrder(Missing) -> Order
`
	output, code, stderr := analyzeFixtureOutput(t, semanticInvalid)
	if code != exitFailure || len(stderr) != 0 {
		t.Fatalf("semantic plan result = code %d, stderr=%q", code, stderr)
	}
	plan := decodeFixPlan(t, output)
	if plan.Status != fixPlanSemanticInvalid || plan.IR.Status != "unavailable" || len(plan.Diagnostics) != 1 {
		t.Fatalf("unexpected semantic plan: %#v", plan)
	}
	diagnostic := plan.Diagnostics[0]
	if diagnostic.Phase != "semantic" || diagnostic.Code != "semantic.lowering" || diagnostic.Span.File != "fixture.gooo" || diagnostic.Status != "deferred" || diagnostic.Applicability != "not-evaluated" {
		t.Fatalf("unexpected semantic diagnostic: %#v", diagnostic)
	}
	if plan.GraphHash != "" || plan.GraphPatch.Status != "deferred" || plan.Workspace.Status != "deferred" {
		t.Fatalf("semantic-invalid plan exposed unavailable output: %#v", plan)
	}
}
func TestRunAnalyzePermutationPreservesCanonicalPlan(t *testing.T) {
	first := decodeFixPlan(t, analyzeFixtureBytes(t, sourceOrderA))
	second := decodeFixPlan(t, analyzeFixtureBytes(t, sourceOrderB))
	if first.SourceDigest == second.SourceDigest || first.GraphHash != second.GraphHash || first.IR.SemanticDigest != second.IR.SemanticDigest {
		t.Fatalf("source permutation changed semantic plan identity: first=%#v second=%#v", first, second)
	}
	first.SourceDigest, second.SourceDigest = "", ""
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatalf("canonical plan differs by source order:\nfirst=%s\nsecond=%s", firstJSON, secondJSON)
	}
}
