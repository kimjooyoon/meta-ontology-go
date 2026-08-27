package ambiguitybudget

import (
	"strings"
	"testing"
)

func TestEvaluateObservesExecutableIntegerSets(t *testing.T) {
	receipt := Evaluate(validInput())
	if receipt.ConformanceDecision != "PASS" || receipt.ConformanceResolution != "EXACT" ||
		receipt.SubjectDecision != "MIXED" || receipt.SubjectResolution != "LOWER_RESOLUTION" {
		t.Fatalf("receipt decisions = %s/%s subject=%s/%s", receipt.ConformanceDecision, receipt.ConformanceResolution, receipt.SubjectDecision, receipt.SubjectResolution)
	}
	want := map[string][4]string{
		"zero-ambiguity":        {"PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT", "EXACT"},
		"boundary-ambiguity":    {"PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT", "EXACT"},
		"over-budget-ambiguity": {"FAIL_CLOSED", "LOWER_RESOLUTION", "AMBIGUITY_BUDGET_EXCEEDED", "LOWER_RESOLUTION"},
		"unknown-ambiguity":     {"UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_INPUT_UNKNOWN", "OPEN"},
	}
	for _, result := range receipt.Cases {
		wantCase, ok := want[result.ID]
		if !ok || [4]string{result.Decision, result.Resolution, result.Reason, result.Claim.To} != wantCase {
			t.Fatalf("case %q = %#v, want %v", result.ID, result, wantCase)
		}
	}
	if receipt.Budget != expectedBudget() || receipt.Summary.FixedDenominator != 2 ||
		receipt.Summary.IntegerDimensions != 3 || receipt.Summary.InterventionsTotal != 2 {
		t.Fatalf("budget/summary = %#v / %#v", receipt.Budget, receipt.Summary)
	}
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluatePreservesClaimProvenance(t *testing.T) {
	receipt := Evaluate(validInput())
	if len(receipt.Claims) != len(receipt.Cases) {
		t.Fatalf("claims=%d cases=%d", len(receipt.Claims), len(receipt.Cases))
	}
	for index, result := range receipt.Cases {
		claim := receipt.Claims[index]
		if claim != result.Claim || claim.Stage == "" || claim.Step == "" || claim.Reason == "" || claim.EvidenceDigest == "" {
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
	if !semantic.Satisfied || semantic.CountsBefore == semantic.CountsAfter ||
		semantic.SemanticDigestBefore == semantic.SemanticDigestAfter || semantic.ResolutionAfter != "LOWER_RESOLUTION" {
		t.Fatalf("semantic intervention = %#v", semantic)
	}
	if !nonsemantic.Satisfied || nonsemantic.SourceDigestBefore == nonsemantic.SourceDigestAfter ||
		nonsemantic.SemanticDigestBefore != nonsemantic.SemanticDigestAfter || nonsemantic.CountsBefore != nonsemantic.CountsAfter {
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
		Schema: ContractSchema, ID: "ambiguity-budget", SourcePath: "main.gooo",
		SourcePackage: "ambiguitybudget", SourceNamespace: "ambiguitybudget", BudgetActivity: "FixedBudget",
		FixedDenominator: FixedDenominator, NotClaimed: []string{"NATURAL_LANGUAGE_CONFIDENCE", "PARSE_TREE_PROBABILITY", "SEMANTIC_CORRECTNESS", "INTENT_RECOGNITION"},
		Cases: []CaseContract{
			{ID: "zero-ambiguity", Activity: "ZeroAmbiguity"},
			{ID: "boundary-ambiguity", Activity: "BoundaryAmbiguity"},
			{ID: "over-budget-ambiguity", Activity: "OverBudgetAmbiguity"},
			{ID: "unknown-ambiguity", Activity: "UnknownAmbiguity"},
		},
		Interventions: []InterventionContract{
			{ID: "semantic-count-crosses-boundary", Kind: "SEMANTIC", TargetActivity: "BoundaryAmbiguity"},
			{ID: "nonsemantic-comment-only", Kind: "NONSEMANTIC", TargetActivity: "BoundaryAmbiguity"},
		},
	}
	return Input{SubjectSHA: strings.Repeat("a", 40), Contract: contract, Source: []byte(sourceFixture)}
}

const sourceFixture = `package ambiguitybudget
namespace ambiguitybudget
entity Candidate id "gooo://ambiguity-budget/entity/candidate"
entity Branch id "gooo://ambiguity-budget/entity/branch"
entity Evidence id "gooo://ambiguity-budget/entity/evidence"
entity Receipt id "gooo://ambiguity-budget/entity/receipt"
activity FixedBudget() -> Receipt computes "ambiguity-budget:budget:2,1,2"
activity ZeroAmbiguity() -> Receipt computes "ambiguity-budget:case:zero-ambiguity:ZERO:KNOWN:1,0,1"
activity BoundaryAmbiguity() -> Receipt computes "ambiguity-budget:case:boundary-ambiguity:BOUNDARY:KNOWN:2,1,2"
activity OverBudgetAmbiguity() -> Receipt computes "ambiguity-budget:case:over-budget-ambiguity:OVER:KNOWN:3,2,3"
activity UnknownAmbiguity() -> Receipt computes "ambiguity-budget:case:unknown-ambiguity:UNKNOWN:UNKNOWN:2,1,2"`
