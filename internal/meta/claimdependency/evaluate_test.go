package claimdependency

import "testing"

func fixtureSource(marker string) []byte {
	return []byte("package claimdependency\nnamespace claimdependency\nentity Integer id \"gooo://claim-dependency/entity/integer\"\nactivity Root(Integer) -> Integer computes \"" + marker + "\"\nactivity Derived(Integer) -> Integer computes \"int.add:2\"\n")
}

func TestDirectUnknownPreservesFailureResponsibility(t *testing.T) {
	receipt, err := Evaluate(fixtureSource("int.unknown:1"), "examples/language-claim-dependency/unknown.gooo", CaseDirectUnknown)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Metrics.FixedClaimTotal != 6 || receipt.Metrics.FixedEdgeTotal != 8 || receipt.Metrics.DirectUnknownClaimTotal != 1 || receipt.Metrics.DependencyBlockedClaimTotal != 5 || receipt.Metrics.ObservedBlockingEdgeTotal != 8 || receipt.Metrics.MaximumCausePathDepth != 2 {
		t.Fatalf("unexpected direct unknown metrics: %+v", receipt.Metrics)
	}
	if receipt.Resolutions[0].FailureResponsibility != "LOCAL_PRODUCER" || receipt.Resolutions[1].FailureResponsibility != "UPSTREAM_CLAIM" {
		t.Fatalf("failure responsibility was not separated: %+v", receipt.Resolutions[:2])
	}
	want := []string{"gooo.claim.dependency.source-observed.v1", "gooo.claim.dependency.producer-bound.v1", "gooo.claim.dependency.decision-replay-bound.v1"}
	got := receipt.Resolutions[5].CausePath
	if len(got) != len(want) {
		t.Fatalf("decision path length: got %d want %d (%v)", len(got), len(want), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("decision path: got %v want %v", got, want)
		}
	}
}

func TestRefutationAndRecoveryUseDistinctTransitions(t *testing.T) {
	refuted, err := Evaluate(fixtureSource("int.add:-1"), "examples/language-claim-dependency/refuted.gooo", CaseRefuted)
	if err != nil {
		t.Fatal(err)
	}
	if refuted.Metrics.RefutedClaimTotal != 6 || refuted.Metrics.DirectRefutedClaimTotal != 1 || refuted.Metrics.DependencyRefutedClaimTotal != 5 || refuted.Metrics.ObservedRefutingEdgeTotal != 8 || refuted.Transitions[ClaimTotal].After != "REFUTED" {
		t.Fatalf("unexpected refuted result: %+v", refuted.Metrics)
	}
	recovered, err := Evaluate(fixtureSource("int.add:1"), "examples/language-claim-dependency/main.gooo", CaseRecovered)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Metrics.DischargedClaimTotal != 6 || recovered.Metrics.DependencyRecoveredTotal != 5 || recovered.Metrics.ObservedRecoveryEdgeTotal != 5 || recovered.Decision.Value != "PASS" {
		t.Fatalf("unexpected recovered result: %+v", recovered.Metrics)
	}
	if recovered.Transitions[ClaimTotal].Before != "OPEN" || recovered.Transitions[ClaimTotal].After != "DISCHARGED" || recovered.Transitions[ClaimTotal+1].Event != "DEPENDENCY_RECOVERED" {
		t.Fatalf("recovery transition vocabulary changed: %+v", recovered.Transitions[ClaimTotal:ClaimTotal+2])
	}
}

func TestRejectsSourceWithoutControlledMarker(t *testing.T) {
	if _, err := Evaluate(fixtureSource("int.add:2"), "examples/language-claim-dependency/main.gooo", CaseDirectUnknown); err == nil {
		t.Fatal("source without direct-unknown marker was accepted")
	}
}
