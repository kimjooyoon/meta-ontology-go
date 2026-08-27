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
activity Root(Seed) -> RootState computes "` + root + `|root-observation"
activity Derived(RootState) -> DerivedState computes "claim.edge:requires|derived-integrity"
activity SupportCheck(RootState, DerivedState) -> SupportState computes "claim.edge:supports|support-observation"
activity RequirementCheck(DerivedState, SupportState) -> RequirementState computes "claim.edge:requires|requirement-integrity"
activity ContradictionCheck(RootState, RequirementState) -> ContradictionState computes "claim.edge:contradicts|contradiction-observation"
activity FailureEntailmentCheck(ContradictionState) -> FinalState computes "claim.edge:failure-entailment|failure-observation"
`)
}

func TestGraphUsesDistinctExecutionPropositionsAndTypedEdges(t *testing.T) {
	graph, err := GraphFromSource(fixtureSource("claim.observe:recoverable"), "fixture.gooo")
	if err != nil {
		t.Fatal(err)
	}
	if graph.NodeTotal != ClaimTotal || graph.EdgeTotal != EdgeTotal {
		t.Fatalf("unexpected graph denominator: %+v", graph)
	}
	seen := map[string]bool{}
	for _, claim := range graph.Nodes {
		if claim.Proposition == "" || claim.PropositionDigest == "" || claim.Target.Artifact == "" {
			t.Fatalf("claim is not executable proposition: %+v", claim)
		}
		seen[claim.PropositionDigest] = true
	}
	if len(seen) != ClaimTotal {
		t.Fatalf("propositions are not distinct: %d", len(seen))
	}
}

func TestEvaluateRejectsNonProviderEvidence(t *testing.T) {
	_, err := Evaluate(fixtureSource("claim.observe:recoverable"), "fixture.gooo", EvidenceReceipt{}, nil)
	if err == nil {
		t.Fatal("zero evidence receipt was accepted")
	}
}

func TestTruthTableHasPositiveAndNegativeCasePerEdgeKind(t *testing.T) {
	counts := map[EdgeKind]int{}
	for _, value := range TruthTableCases() {
		counts[value.Kind]++
	}
	for _, kind := range EdgeKinds() {
		if counts[kind] != 2 {
			t.Fatalf("truth table denominator for %s = %d", kind, counts[kind])
		}
	}
}
