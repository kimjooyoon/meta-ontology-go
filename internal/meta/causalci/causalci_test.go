package causalci

import (
	"encoding/json"
	"testing"
)

func TestCausalSelectionClassifiesThreeScenarioCorpus(t *testing.T) {
	input := fixtureInput(t)
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := Evaluate(raw, input.SourcePath, []byte("package causalci\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err := Verify(raw, input.SourcePath, []byte("package causalci\n"), receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Cases[0].Decision != DecisionSelected || receipt.Cases[1].Decision != DecisionFullFallback || receipt.Cases[2].Decision != DecisionRejected {
		t.Fatalf("decisions = %#v", receipt.Cases)
	}
	if receipt.Metrics.FixedCheckDenominator != FixedCheckDenominator || len(receipt.Indicators) != FixedIndicatorDenominator {
		t.Fatalf("fixed contract = %#v", receipt.Metrics)
	}
}

func TestIndependentVerifierRejectsTamperedChoice(t *testing.T) {
	input := fixtureInput(t)
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := Evaluate(raw, input.SourcePath, []byte("package causalci\n"))
	if err != nil {
		t.Fatal(err)
	}
	receipt.Cases[0].SelectedChecks[0].CheckID = "ci-policy"
	if err := Verify(raw, input.SourcePath, []byte("package causalci\n"), receipt); err == nil {
		t.Fatal("tampered receipt passed independent verification")
	}
}

func fixtureInput(t *testing.T) Input {
	t.Helper()
	checks := make([]Check, 0, FixedCheckDenominator)
	for index, id := range requiredCheckIDs {
		checks = append(checks, Check{ID: id, Ordinal: index + 1, Scope: "canonical-ci", Description: "fixed proof check"})
	}
	coordinate := Coordinate{Stage: "CLAIM_LEDGER", Step: "append", Reason: "SOURCE_ASSERTION"}
	return Input{
		Schema: InputSchema, SourcePath: "examples/causal-ci-selection/main.gooo",
		Operation:        Operation{Producer: "gooo://producer/causal-ci-selection", Consumer: "gooo://consumer/github-actions", MetaOperation: "causal-ci-select", ProofChoice: "CLAIM_IMPACT_REASON", ReadOnly: true},
		Policy:           Policy{Schema: PolicySchema, Checks: checks, FullSuiteID: "full-suite"},
		ClaimTransitions: []ClaimTransition{{Sequence: 1, ClaimID: "claim:source", Before: "UNOBSERVED", After: "OBSERVED", Event: "SOURCE_CLAIM_REGISTERED", Coordinate: coordinate, EvidenceDigest: "source-evidence-1"}},
		Cases: []Case{
			{ID: "selection", ChangedFiles: []string{"internal/meta/causalci/evaluate.go"}, Claims: []Claim{{ID: "claim:source", Question: "does the changed source affect the receipt?", State: "OBSERVED"}}, ImpactEdges: []ImpactEdge{{ID: "E1", From: "internal/meta/causalci/evaluate.go", To: "claim:source", Kind: "CHANGES_CLAIM", Known: true, Reason: "source span binds claim", Coordinate: coordinate}, {ID: "E2", From: "claim:source", To: "check:go-test", Kind: "CLAIM_REQUIRES_CHECK", Known: true, Reason: "receipt behavior requires tests", Coordinate: coordinate}}},
			{ID: "full-fallback", ChangedFiles: []string{"internal/meta/causalci/evaluate.go"}, Claims: []Claim{{ID: "claim:source", Question: "can the changed source owner be resolved?", State: "UNKNOWN"}}, ImpactEdges: []ImpactEdge{{ID: "E1", From: "internal/meta/causalci/evaluate.go", To: "claim:source", Kind: "CHANGES_CLAIM", Known: false, Reason: "OWNER_RESOLUTION_UNAVAILABLE", Coordinate: Coordinate{Stage: "CAUSAL_SELECTION", Step: "trace-impact", Reason: "OWNER_RESOLUTION_UNAVAILABLE"}}}},
			{ID: "rejection", ChangedFiles: []string{"internal/meta/causalci/evaluate.go"}, Claims: []Claim{{ID: "claim:source", Question: "is the check catalog reference valid?", State: "OBSERVED"}}, ImpactEdges: []ImpactEdge{{ID: "E1", From: "internal/meta/causalci/evaluate.go", To: "claim:source", Kind: "CHANGES_CLAIM", Known: true, Reason: "source span binds claim", Coordinate: coordinate}, {ID: "E2", From: "claim:source", To: "check:missing", Kind: "CLAIM_REQUIRES_CHECK", Known: true, Reason: "unregistered check", Coordinate: coordinate}}},
		},
	}
}
