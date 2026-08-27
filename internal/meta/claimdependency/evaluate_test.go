package claimdependency

import "testing"

func fixtureSource(root string) []byte {
	return []byte(`package claimdependency
namespace claimdependency
entity Seed id "gooo://claim-dependency/entity/seed"
entity RootState id "gooo://claim-dependency/entity/root-state"
entity DerivedState id "gooo://claim-dependency/entity/derived-state"
entity SupportState id "gooo://claim-dependency/entity/support-state"
entity RequirementState id "gooo://claim-dependency/entity/requirement-state"
entity ContradictionState id "gooo://claim-dependency/entity/contradiction-state"
entity FinalState id "gooo://claim-dependency/entity/final-state"
activity Root(Seed) -> RootState computes "` + root + `"
activity Derived(RootState) -> DerivedState computes "claim.edge:requires"
activity SupportCheck(RootState, DerivedState) -> SupportState computes "claim.edge:supports"
activity RequirementCheck(DerivedState, SupportState) -> RequirementState computes "claim.edge:requires"
activity ContradictionCheck(RootState, RequirementState) -> ContradictionState computes "claim.edge:contradicts"
activity FailureEntailmentCheck(ContradictionState) -> FinalState computes "claim.edge:failure-entailment"
`)
}

func TestUnknownKeepsDependentSupportAndRequirementClaimsOpen(t *testing.T) {
	source := fixtureSource("claim.observe:recoverable")
	observation, err := ObservationForSource(source, "fixture.gooo", ObservationUnknown, "")
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := Evaluate(source, "fixture.gooo", observation, nil)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Metrics.FixedClaimTotal != 6 || receipt.Metrics.FixedEdgeTotal != 8 || receipt.Metrics.DirectUnknownClaimTotal != 1 || receipt.Metrics.DependencyBlockedClaimTotal != 5 || receipt.Metrics.ObservedBlockingEdgeTotal != 8 || receipt.Metrics.ObservedRefutingEdgeTotal != 0 {
		t.Fatalf("unexpected UNKNOWN metrics: %+v", receipt.Metrics)
	}
	if receipt.Resolutions[0].State != "OPEN" || receipt.Resolutions[1].State != "OPEN" || receipt.Resolutions[1].FailureResponsibility != "UPSTREAM_CLAIM" {
		t.Fatalf("UNKNOWN responsibility was not preserved: %+v", receipt.Resolutions[:2])
	}
}

func TestOnlyExplicitRefutingEdgesPropagateRefuted(t *testing.T) {
	source := fixtureSource("claim.observe:contradiction")
	observation, err := ObservationForSource(source, "fixture.gooo", ObservationContradiction, "explicit contradiction evidence")
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := Evaluate(source, "fixture.gooo", observation, nil)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Metrics.DirectRefutedClaimTotal != 1 || receipt.Metrics.DependencyRefutedClaimTotal != 2 || receipt.Metrics.OpenClaimTotal != 3 || receipt.Metrics.ObservedRefutingEdgeTotal != 2 || receipt.Metrics.ObservedBlockingEdgeTotal != 5 {
		t.Fatalf("unexpected refutation metrics: %+v", receipt.Metrics)
	}
	if receipt.Resolutions[1].State != "OPEN" || receipt.Resolutions[4].State != "REFUTED" || receipt.Resolutions[5].State != "REFUTED" {
		t.Fatalf("typed edge propagation was not selective: %+v", receipt.Resolutions)
	}
}

func TestRecoveryAppendsPriorUnknownLedger(t *testing.T) {
	source := fixtureSource("claim.observe:recoverable")
	unknownObservation, err := ObservationForSource(source, "unknown.gooo", ObservationUnknown, "")
	if err != nil {
		t.Fatal(err)
	}
	unknown, err := Evaluate(source, "unknown.gooo", unknownObservation, nil)
	if err != nil {
		t.Fatal(err)
	}
	recoveryObservation, err := ObservationForSource(source, "main.gooo", ObservationEvidence, "accepted evidence")
	if err != nil {
		t.Fatal(err)
	}
	recovered, err := Evaluate(source, "main.gooo", recoveryObservation, &unknown)
	if err != nil {
		t.Fatal(err)
	}
	if recovered.Metrics.TransitionTotal != 18 || recovered.Metrics.AppendOnlyTransitionTotal != 6 || recovered.Metrics.DischargedClaimTotal != 6 || recovered.Metrics.ObservedRecoveryEdgeTotal != 8 || recovered.PriorReceiptDigest == "" || recovered.PreviousTransitionDigest != unknown.TransitionHeadDigest {
		t.Fatalf("recovery did not append the prior ledger: %+v", recovered.Metrics)
	}
	for index := range unknown.Transitions {
		if recovered.Transitions[index] != unknown.Transitions[index] {
			t.Fatalf("recovery rewrote transition %d", index+1)
		}
	}
}

func TestSourceWithoutObservationPredicateFailsClosed(t *testing.T) {
	source := fixtureSource("int.add:-1")
	if _, err := ObservationForSource(source, "fixture.gooo", ObservationContradiction, "evidence"); err == nil {
		t.Fatal("non-claim source predicate was accepted")
	}
}
