package ambiguitybudget

import (
	"strings"
	"testing"
)

func TestEvaluateDerivesIntegerSetsFromLoweredGraph(t *testing.T) {
	receipt := Evaluate(validInput())
	if receipt.ConformanceDecision != "PASS" || receipt.ConformanceResolution != "EXACT" ||
		receipt.SubjectDecision != "MIXED" || receipt.SubjectResolution != "LOWER_RESOLUTION" {
		t.Fatalf("receipt decisions = %s/%s subject=%s/%s", receipt.ConformanceDecision, receipt.ConformanceResolution, receipt.SubjectDecision, receipt.SubjectResolution)
	}
	want := map[string]struct {
		counts     IntegerSet
		class      string
		decision   string
		resolution string
		reason     string
		claimTo    string
	}{
		"zero-ambiguity":        {IntegerSet{1, 0, 1}, "ZERO", "PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT", "DISCHARGED"},
		"boundary-ambiguity":    {IntegerSet{2, 1, 2}, "BOUNDARY", "PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT", "DISCHARGED"},
		"over-budget-ambiguity": {IntegerSet{3, 2, 3}, "OVER", "FAIL_CLOSED", "LOWER_RESOLUTION", "AMBIGUITY_BUDGET_EXCEEDED", "REFUTED"},
		"unknown-ambiguity":     {IntegerSet{2, 0, 2}, "UNKNOWN", "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_COORDINATE_UNOBSERVED", "OPEN"},
	}
	for _, result := range receipt.Cases {
		wantCase, ok := want[result.ID]
		if !ok || result.Counts != wantCase.counts || result.Class != wantCase.class ||
			result.Decision != wantCase.decision || result.Resolution != wantCase.resolution ||
			result.Reason != wantCase.reason || result.Claim.To != wantCase.claimTo {
			t.Fatalf("case %q = %#v, want %#v", result.ID, result, wantCase)
		}
		if result.Claim.From != "OPEN" || result.Claim.Proposition == "" || result.Claim.PropositionDigest == "" || result.Claim.EvidenceDigest == "" {
			t.Fatalf("case %q claim = %#v", result.ID, result.Claim)
		}
	}
	if receipt.BudgetAuthority != "CONTRACT_POLICY" || receipt.BudgetBinding != "ambiguity-budget:budget-policy:v1" ||
		receipt.Summary.Denominator != expectedDenominator() || receipt.Summary.Numerator.IntegerObservationsObserved != 11 ||
		receipt.Summary.Numerator.ClaimsDischarged != 2 || receipt.Summary.Numerator.ClaimsRefuted != 1 ||
		receipt.Summary.Numerator.ClaimsOpen != 1 || receipt.Summary.Numerator.InterventionsSatisfied != 2 ||
		receipt.Summary.Numerator.AuthorityObserved != 0 {
		t.Fatalf("policy/accounting = %#v / %#v", receipt.BudgetPolicy, receipt.Summary)
	}
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
	unknown := receipt.Cases[3]
	if unknown.ID != "unknown-ambiguity" || unknown.UnobservedDimensions[0] != "unresolved_branches" ||
		unknown.Coordinate != (Coordinate{Stage: "AMBIGUITY_OBSERVATION", Step: "unresolved_branches", Reason: "AMBIGUITY_COORDINATE_UNOBSERVED"}) {
		t.Fatalf("unknown coordinate = %#v", unknown)
	}
}

func TestEvaluateClaimLedgerUsesSemanticEvidence(t *testing.T) {
	receipt := Evaluate(validInput())
	if len(receipt.Claims) != len(receipt.Cases) {
		t.Fatalf("claims=%d cases=%d", len(receipt.Claims), len(receipt.Cases))
	}
	for index, result := range receipt.Cases {
		claim := receipt.Claims[index]
		if claim != result.Claim || claim.Stage == "" || claim.Step == "" || claim.Reason == "" || claim.EvidenceDigest == "" ||
			strings.Contains(claim.EvidenceDigest, result.RawSourceDigest) {
			t.Fatalf("claim %d = %#v, case=%#v", index, claim, result)
		}
	}
}

func TestEvaluateInterventionsSeparateSemanticAndNonsemanticChanges(t *testing.T) {
	receipt := Evaluate(validInput())
	if len(receipt.Interventions) != 2 {
		t.Fatalf("interventions = %#v", receipt.Interventions)
	}
	semantic, nonsemantic := receipt.Interventions[0], receipt.Interventions[1]
	if !semantic.Satisfied || semantic.CountsBefore == semantic.CountsAfter || semantic.SemanticDigestBefore == semantic.SemanticDigestAfter ||
		len(semantic.ElementsAfter.CandidateIDs) != 3 || len(semantic.ElementsAfter.EvidencePathIDs) != 3 ||
		semantic.ResolutionAfter != "LOWER_RESOLUTION" || semantic.ClaimBefore.To != "DISCHARGED" || semantic.ClaimAfter.To != "REFUTED" {
		t.Fatalf("semantic intervention = %#v", semantic)
	}
	if !nonsemantic.Satisfied || nonsemantic.SourceDigestBefore == nonsemantic.SourceDigestAfter ||
		nonsemantic.SemanticDigestBefore != nonsemantic.SemanticDigestAfter || nonsemantic.CountsBefore != nonsemantic.CountsAfter ||
		nonsemantic.ClaimBefore != nonsemantic.ClaimAfter {
		t.Fatalf("nonsemantic intervention = %#v", nonsemantic)
	}
}

func TestEvaluateUnknownSourceLowersConformanceAndSubject(t *testing.T) {
	input := validInput()
	input.Source = []byte("package wrong\nnamespace wrong\n")
	receipt := Evaluate(input)
	if receipt.ConformanceDecision != "FAIL_CLOSED" || receipt.ConformanceResolution != "LOWER_RESOLUTION" ||
		receipt.SubjectDecision != "UNKNOWN" || receipt.SubjectResolution != "LOWER_RESOLUTION" ||
		receipt.SubjectCoordinate.Stage != "ambiguity-budget" || receipt.SubjectCoordinate.Step != "observe-source" {
		t.Fatalf("receipt = %#v", receipt)
	}
}

func validInput() Input {
	contract := Contract{
		Schema: ContractSchema, ID: "gooo://ambiguity-budget/contract/v3", SourcePath: "examples/ambiguity-budget/main.gooo",
		SourcePackage: "ambiguitybudget", SourceNamespace: "ambiguitybudget", BudgetActivity: "FixedBudget",
		BudgetPolicy: BudgetPolicy{Schema: PolicySchema, ID: "gooo://ambiguity-budget/policy/v1", Version: "v1", Authority: "CONTRACT_POLICY",
			Dimensions: []BudgetDimension{{ID: "interpretation_candidates", Limit: 2}, {ID: "unresolved_branches", Limit: 1}, {ID: "evidence_paths", Limit: 2}}},
		Denominator: expectedDenominator(), NotClaimed: []string{"NATURAL_LANGUAGE_CONFIDENCE", "PARSE_TREE_PROBABILITY", "SEMANTIC_CORRECTNESS", "INTENT_RECOGNITION"},
		Cases: []CaseContract{
			{ID: "zero-ambiguity", Activity: "ZeroAmbiguity"}, {ID: "boundary-ambiguity", Activity: "BoundaryAmbiguity"},
			{ID: "over-budget-ambiguity", Activity: "OverBudgetAmbiguity"}, {ID: "unknown-ambiguity", Activity: "UnknownAmbiguity"},
		},
		Interventions: []InterventionContract{
			{ID: "semantic-count-crosses-boundary", Kind: "SEMANTIC", TargetActivity: "BoundaryAmbiguity"},
			{ID: "nonsemantic-comment-only", Kind: "NONSEMANTIC", TargetActivity: "BoundaryAmbiguity"},
		},
	}
	return Input{SubjectSHA: strings.Repeat("a", 40), Contract: contract, Source: []byte(sourceFixture), EffectsArtifact: []byte(effectsFixture)}
}

const effectsFixture = `{"schema":"gooo/ambiguity-budget-workspace-effects/v1","version":"v1","tracked_and_untracked":true,"snapshot_before":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","snapshot_after":"sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","repository_writes":0,"write_set_equal":true,"mutation_authority":"UNKNOWN","mutation_authority_resolution":"NOT_OBSERVED"}`

const sourceFixture = `package ambiguitybudget
namespace ambiguitybudget

entity AmbiguityReceipt id "gooo://ambiguity-budget/entity/ambiguity-receipt"

entity ZeroCase id "gooo://ambiguity-budget/case/zero-ambiguity"
entity ZeroCandidateA id "gooo://ambiguity-budget/case/zero-ambiguity/candidate/a"
entity ZeroResolvedBranchA id "gooo://ambiguity-budget/case/zero-ambiguity/branch/resolved/a"
entity ZeroEvidencePathA id "gooo://ambiguity-budget/case/zero-ambiguity/evidence-path/a"
entity ZeroBranchObservation id "gooo://ambiguity-budget/case/zero-ambiguity/branch-observation"

entity BoundaryCase id "gooo://ambiguity-budget/case/boundary-ambiguity"
entity BoundaryCandidateA id "gooo://ambiguity-budget/case/boundary-ambiguity/candidate/a"
entity BoundaryCandidateB id "gooo://ambiguity-budget/case/boundary-ambiguity/candidate/b"
entity BoundaryResolvedBranchA id "gooo://ambiguity-budget/case/boundary-ambiguity/branch/resolved/a"
entity BoundaryUnresolvedBranchA id "gooo://ambiguity-budget/case/boundary-ambiguity/branch/unresolved/a"
entity BoundaryEvidencePathA id "gooo://ambiguity-budget/case/boundary-ambiguity/evidence-path/a"
entity BoundaryEvidencePathB id "gooo://ambiguity-budget/case/boundary-ambiguity/evidence-path/b"
entity BoundaryBranchObservation id "gooo://ambiguity-budget/case/boundary-ambiguity/branch-observation"

entity OverCase id "gooo://ambiguity-budget/case/over-budget-ambiguity"
entity OverCandidateA id "gooo://ambiguity-budget/case/over-budget-ambiguity/candidate/a"
entity OverCandidateB id "gooo://ambiguity-budget/case/over-budget-ambiguity/candidate/b"
entity OverCandidateC id "gooo://ambiguity-budget/case/over-budget-ambiguity/candidate/c"
entity OverResolvedBranchA id "gooo://ambiguity-budget/case/over-budget-ambiguity/branch/resolved/a"
entity OverUnresolvedBranchA id "gooo://ambiguity-budget/case/over-budget-ambiguity/branch/unresolved/a"
entity OverUnresolvedBranchB id "gooo://ambiguity-budget/case/over-budget-ambiguity/branch/unresolved/b"
entity OverEvidencePathA id "gooo://ambiguity-budget/case/over-budget-ambiguity/evidence-path/a"
entity OverEvidencePathB id "gooo://ambiguity-budget/case/over-budget-ambiguity/evidence-path/b"
entity OverEvidencePathC id "gooo://ambiguity-budget/case/over-budget-ambiguity/evidence-path/c"
entity OverBranchObservation id "gooo://ambiguity-budget/case/over-budget-ambiguity/branch-observation"

entity UnknownCase id "gooo://ambiguity-budget/case/unknown-ambiguity"
entity UnknownCandidateA id "gooo://ambiguity-budget/case/unknown-ambiguity/candidate/a"
entity UnknownCandidateB id "gooo://ambiguity-budget/case/unknown-ambiguity/candidate/b"
entity UnknownEvidencePathA id "gooo://ambiguity-budget/case/unknown-ambiguity/evidence-path/a"
entity UnknownEvidencePathB id "gooo://ambiguity-budget/case/unknown-ambiguity/evidence-path/b"

activity FixedBudget() -> AmbiguityReceipt computes "ambiguity-budget:budget-policy:v1"
activity ZeroAmbiguity(ZeroCase) -> AmbiguityReceipt computes "ambiguity-budget:case:zero-ambiguity"
activity BoundaryAmbiguity(BoundaryCase) -> AmbiguityReceipt computes "ambiguity-budget:case:boundary-ambiguity"
activity OverBudgetAmbiguity(OverCase) -> AmbiguityReceipt computes "ambiguity-budget:case:over-budget-ambiguity"
activity UnknownAmbiguity(UnknownCase) -> AmbiguityReceipt computes "ambiguity-budget:case:unknown-ambiguity"

activity ObserveZeroCandidateA(ZeroCase) -> ZeroCandidateA
activity ObserveZeroResolvedBranchA(ZeroCase) -> ZeroResolvedBranchA
activity ObserveZeroBranchObservation(ZeroCase) -> ZeroBranchObservation
activity ObserveZeroEvidencePathA(ZeroCase, ZeroCandidateA) -> ZeroEvidencePathA

activity ObserveBoundaryCandidateA(BoundaryCase) -> BoundaryCandidateA
activity ObserveBoundaryCandidateB(BoundaryCase) -> BoundaryCandidateB
activity ObserveBoundaryResolvedBranchA(BoundaryCase) -> BoundaryResolvedBranchA
activity ObserveBoundaryUnresolvedBranchA(BoundaryCase) -> BoundaryUnresolvedBranchA
activity ObserveBoundaryBranchObservation(BoundaryCase) -> BoundaryBranchObservation
activity ObserveBoundaryEvidencePathA(BoundaryCase, BoundaryCandidateA) -> BoundaryEvidencePathA
activity ObserveBoundaryEvidencePathB(BoundaryCase, BoundaryCandidateB) -> BoundaryEvidencePathB

activity ObserveOverCandidateA(OverCase) -> OverCandidateA
activity ObserveOverCandidateB(OverCase) -> OverCandidateB
activity ObserveOverCandidateC(OverCase) -> OverCandidateC
activity ObserveOverResolvedBranchA(OverCase) -> OverResolvedBranchA
activity ObserveOverUnresolvedBranchA(OverCase) -> OverUnresolvedBranchA
activity ObserveOverUnresolvedBranchB(OverCase) -> OverUnresolvedBranchB
activity ObserveOverBranchObservation(OverCase) -> OverBranchObservation
activity ObserveOverEvidencePathA(OverCase, OverCandidateA) -> OverEvidencePathA
activity ObserveOverEvidencePathB(OverCase, OverCandidateB) -> OverEvidencePathB
activity ObserveOverEvidencePathC(OverCase, OverCandidateC) -> OverEvidencePathC

activity ObserveUnknownCandidateA(UnknownCase) -> UnknownCandidateA
activity ObserveUnknownCandidateB(UnknownCase) -> UnknownCandidateB
activity ObserveUnknownEvidencePathA(UnknownCase, UnknownCandidateA) -> UnknownEvidencePathA
activity ObserveUnknownEvidencePathB(UnknownCase, UnknownCandidateB) -> UnknownEvidencePathB`
