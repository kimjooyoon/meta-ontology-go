package workfrontier

import (
	"testing"
)

func TestR4BindingsRequireCanonicalPayloadProof(t *testing.T) {
	input := r4FixtureInput(t, "acyclic")
	if got := EvaluateR4(input); got.Status != R4StatusPass {
		t.Fatalf("valid bindings = %#v", got)
	}
	alternate := r4FixtureInput(t, "acyclic")
	alternate.States[0].Status = "PASS"
	projections, err := r4ProjectionBytes(alternate)
	if err != nil {
		t.Fatal(err)
	}

	stale := input
	stale.SnapshotDigest = r4BindingDigest(string(projections.snapshot))
	got := EvaluateR4(stale)
	if got.Status != R4StatusUnknown || got.Reason != R4ReasonSnapshotBindingMismatch || len(got.SelectedIDs) != 0 {
		t.Fatalf("stale binding = %#v", got)
	}

	mutated := input
	mutated.SnapshotPayload = string(projections.snapshot)
	got = EvaluateR4(mutated)
	if got.Status != R4StatusUnknown || got.Reason != R4ReasonSnapshotBindingMismatch || len(got.SelectedIDs) != 0 {
		t.Fatalf("mutated payload = %#v", got)
	}

	missing := input
	missing.PolicyPayload = ""
	got = EvaluateR4(missing)
	if got.Status != R4StatusUnknown || got.Reason != R4ReasonRequiredInputMissing || len(got.SelectedIDs) != 0 {
		t.Fatalf("missing payload = %#v", got)
	}

	malformed := input
	malformed.RegistryPayload = `{"fixture":"r4","fixture":"duplicate"}`
	got = EvaluateR4(malformed)
	if got.Status != R4StatusFailClosed || got.Reason != R4ReasonMalformedBinding || len(got.SelectedIDs) != 0 {
		t.Fatalf("duplicate payload field = %#v", got)
	}
}
func TestR4RootRelocationChangesReachableBinding(t *testing.T) {
	input := r4FixtureInput(t, "acyclic")
	relocated := input
	relocated.RootObligationIDs = []string{"obligation/child"}
	first, err := AnalyzeR4Graph(input)
	if err != nil {
		t.Fatal(err)
	}
	second, err := AnalyzeR4Graph(relocated)
	if err != nil {
		t.Fatal(err)
	}
	if first.GraphDigest == second.GraphDigest || first.SCCDigest == second.SCCDigest {
		t.Fatalf("root relocation did not change reachable digests: %#v %#v", first, second)
	}
	if got := EvaluateR4(relocated); len(got.SelectedIDs) != 0 {
		t.Fatalf("relocated root selected paths without its declared frontier: %#v", got)
	}
}
