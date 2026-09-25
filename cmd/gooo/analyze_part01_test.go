package main

import (
	"bytes"
	"testing"
)

func TestRunAnalyzeValidFixPlanContract(t *testing.T) {
	output, code, stderr := analyzeFixtureOutput(t, sourceOrderA)
	if code != exitOK || len(stderr) != 0 {
		t.Fatalf("analyze result = code %d, stderr=%q", code, stderr)
	}
	plan := decodeFixPlan(t, output)
	if plan.SchemaVersion != fixPlanSchemaVersion || plan.Status != fixPlanReady || plan.SourceDigest == "" || plan.GraphHash == "" {
		t.Fatalf("incomplete fix plan identity: %#v", plan)
	}
	if !bytes.Contains(output, []byte(`"diagnostics":[]`)) || bytes.Contains(output, []byte(`"diagnostics":null`)) {
		t.Fatalf("valid fix plan diagnostics is not an empty JSON array: %s", output)
	}
	if plan.IR.Status != "available" || plan.IR.SemanticDigest == "" || len(plan.Diagnostics) != 0 {
		t.Fatalf("unexpected valid plan state: %#v", plan)
	}
	if plan.Evidence.Status != "missing" || plan.Provenance.Status != "missing" {
		t.Fatalf("evidence status = %#v, provenance = %#v", plan.Evidence, plan.Provenance)
	}
	if plan.Repairs.Status != "deferred" || plan.GraphPatch.Status != "deferred" || plan.Workspace.Status != "deferred" || plan.SemanticLoop.Status != "deferred" || plan.Lowering.Status != "deferred" || plan.Output.Status != "deferred" {
		t.Fatalf("write or loop status was not deferred: %#v", plan)
	}
	wantAuthorities := graphAuthorities{
		GoooSource: "authoritative", SemanticIR: "authoritative", Handwritten: "authoritative",
		Provenance: "authoritative", Graph: "derived",
	}
	if plan.Authorities != wantAuthorities {
		t.Fatalf("authorities = %#v, want %#v", plan.Authorities, wantAuthorities)
	}
}
func TestRunAnalyzeSyntaxDiagnosticsAreStableAndOrdered(t *testing.T) {
	const malformed = "package billing\nentity Broken id \"x\" @"
	first, firstCode, firstErr := analyzeFixtureOutput(t, malformed)
	second, secondCode, secondErr := analyzeFixtureOutput(t, malformed)
	if firstCode != exitFailure || secondCode != exitFailure || len(firstErr) != 0 || len(secondErr) != 0 {
		t.Fatalf("syntax plan result = %d/%d, stderr=%q/%q", firstCode, secondCode, firstErr, secondErr)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("repeated syntax plan differs:\nfirst=%s\nsecond=%s", first, second)
	}
	plan := decodeFixPlan(t, first)
	if plan.Status != fixPlanSyntaxInvalid || plan.IR.Status != "unavailable" || len(plan.Diagnostics) == 0 {
		t.Fatalf("unexpected syntax plan: %#v", plan)
	}
	previousOffset := -1
	for _, diagnostic := range plan.Diagnostics {
		if diagnostic.Phase != "syntax" || diagnostic.RepairID == "" || diagnostic.Status != "deferred" || diagnostic.Applicability != "potential" {
			t.Fatalf("unexpected syntax diagnostic: %#v", diagnostic)
		}
		if diagnostic.Span.File != "fixture.gooo" || diagnostic.Span.Start.Offset < previousOffset {
			t.Fatalf("diagnostics are not source ordered: %#v", plan.Diagnostics)
		}
		previousOffset = diagnostic.Span.Start.Offset
	}
}
