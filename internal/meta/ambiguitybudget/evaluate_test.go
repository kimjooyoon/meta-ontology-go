package ambiguitybudget

import (
	"strings"
	"testing"
)

func TestEvaluateZeroBoundaryOverAndUnknownCases(t *testing.T) {
	input := validInput()
	receipt := Evaluate(input)
	if receipt.Decision != "PASS" || receipt.Resolution != "EXACT" || receipt.Summary.CasesSatisfied != 4 {
		t.Fatalf("receipt = %#v", receipt)
	}
	want := map[string][3]string{
		"zero":     {"PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT"},
		"boundary": {"PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT"},
		"over":     {"FAIL_CLOSED", "LOWER_RESOLUTION", "AMBIGUITY_BUDGET_EXCEEDED"},
		"unknown":  {"UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_INPUT_UNKNOWN"},
	}
	for _, result := range receipt.Cases {
		wantCase, ok := want[result.ID]
		if !ok || [3]string{result.Decision, result.Resolution, result.Reason} != wantCase {
			t.Fatalf("case %q = %#v, want %v", result.ID, result, wantCase)
		}
	}
	if receipt.Summary.ZeroAmbiguityCases != 1 || receipt.Summary.BoundaryCases != 1 ||
		receipt.Summary.OverBudgetCases != 1 || receipt.Summary.UnknownCases != 1 || receipt.Summary.LowerResolutionCases != 2 {
		t.Fatalf("summary = %#v", receipt.Summary)
	}
	if err := Validate(receipt); err != nil {
		t.Fatal(err)
	}
}

func TestEvaluatePreservesClaimTransitions(t *testing.T) {
	receipt := Evaluate(validInput())
	if len(receipt.Claims) != len(receipt.Cases) {
		t.Fatalf("claims=%d cases=%d", len(receipt.Claims), len(receipt.Cases))
	}
	for index, result := range receipt.Cases {
		if receipt.Claims[index] != result.Claim || result.Claim.To != result.Resolution || result.Claim.Reason != result.Reason {
			t.Fatalf("claim %d = %#v, case=%#v", index, receipt.Claims[index], result)
		}
	}
}

func TestEvaluateUnknownSourceLowersResolution(t *testing.T) {
	input := validInput()
	input.Source = []byte("package wrong\nnamespace wrong\n")
	receipt := Evaluate(input)
	if receipt.Decision != "UNKNOWN" || receipt.Resolution != "LOWER_RESOLUTION" ||
		receipt.Coordinate.Stage != "ambiguity-budget" || receipt.Coordinate.Step != "observe-source" {
		t.Fatalf("receipt = %#v", receipt)
	}
}

func validInput() Input {
	contract := Contract{
		Schema: ContractSchema, ID: "ambiguity-budget", SourcePath: "main.gooo",
		SourcePackage: "ambiguitybudget", SourceNamespace: "ambiguitybudget", SourceEntities: 4, SourceActivities: 4,
		Budget:           IntegerSet{InterpretationCandidates: 2, UnresolvedBranches: 1, EvidencePaths: 2},
		FixedDenominator: FixedDenominator, NotClaimed: []string{"NATURAL_LANGUAGE_CONFIDENCE", "PARSE_TREE_PROBABILITY", "SEMANTIC_CORRECTNESS", "INTENT_RECOGNITION"},
		Cases: []CaseSpec{
			caseSpec("zero", "ZERO", "KNOWN", IntegerSet{1, 0, 1}, "PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT"),
			caseSpec("boundary", "BOUNDARY", "KNOWN", IntegerSet{2, 1, 2}, "PASS", "EXACT", "AMBIGUITY_BUDGET_WITHIN_LIMIT"),
			caseSpec("over", "OVER", "KNOWN", IntegerSet{3, 2, 3}, "FAIL_CLOSED", "LOWER_RESOLUTION", "AMBIGUITY_BUDGET_EXCEEDED"),
			caseSpec("unknown", "UNKNOWN", "UNKNOWN", IntegerSet{2, 1, 2}, "UNKNOWN", "LOWER_RESOLUTION", "AMBIGUITY_INPUT_UNKNOWN"),
		},
	}
	return Input{SubjectSHA: strings.Repeat("a", 40), Contract: contract, Source: []byte(sourceFixture)}
}

func caseSpec(id, class, state string, counts IntegerSet, decision, resolution, reason string) CaseSpec {
	return CaseSpec{ID: id, Class: class, InputState: state, Counts: counts,
		ExpectedDecision: decision, ExpectedResolution: resolution, ExpectedReason: reason,
		Coordinate: Coordinate{Stage: "ambiguity-budget", Step: id, Reason: reason},
		Claim:      ClaimTransition{CaseID: id, From: "AMBIGUITY_OBSERVED", To: resolution, Reason: reason}}
}

const sourceFixture = `package ambiguitybudget
namespace ambiguitybudget
entity Candidate id "gooo://ambiguity-budget/entity/candidate"
entity Branch id "gooo://ambiguity-budget/entity/branch"
entity Evidence id "gooo://ambiguity-budget/entity/evidence"
entity Receipt id "gooo://ambiguity-budget/entity/receipt"
activity Measure(Candidate, Branch) -> Receipt
activity CountCandidates(Candidate) -> Receipt
activity CountBranches(Branch) -> Receipt
activity RecordEvidence(Evidence) -> Receipt`
