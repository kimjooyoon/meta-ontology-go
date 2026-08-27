package proofchoicealgebra

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestEvaluateProofChoiceCases(t *testing.T) {
	cases := []struct {
		name, wantDecision, wantReason string
	}{
		{name: "foundation", wantDecision: Pass},
		{name: "coherence", wantDecision: Pass},
		{name: "regression", wantDecision: Pass},
		{name: "combined", wantDecision: Pass},
		{name: "missing-choice", wantDecision: FailClosed, wantReason: "PROOF_CHOICE_MISSING"},
		{name: "contradiction", wantDecision: FailClosed, wantReason: "PROOF_CHOICE_CONTRADICTION"},
		{name: "unknown-context", wantDecision: FailClosed, wantReason: "UNKNOWN_CONTEXT"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			path := filepath.Join("..", "..", "..", "examples", "proof-choice-algebra", testCase.name+".gooo")
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			receipt := Evaluate(path, source)
			if receipt.Decision != testCase.wantDecision {
				t.Fatalf("decision = %s, want %s (%s)", receipt.Decision, testCase.wantDecision, receipt.Reason)
			}
			if testCase.wantReason != "" && receipt.Reason != testCase.wantReason {
				t.Fatalf("reason = %s, want %s", receipt.Reason, testCase.wantReason)
			}
			if receipt.FixedDenom != FixedDenominator || receipt.Effects.RepositoryWrites != 0 || receipt.Effects.MutationAuthority {
				t.Fatalf("unsafe or unfixed receipt: %+v", receipt)
			}
		})
	}
}

func TestCombineIsIdempotentAndRejectsConflicts(t *testing.T) {
	claim := Item{Kind: Claim, ID: "claim", Statement: "stable", Choice: Foundation, Producer: "p", Consumer: "c", MetaOperation: "op", Stage: "s", Step: "step", Reason: "r"}
	transition := Transition{ClaimID: "claim", From: "ASSERTED", To: "PERSISTED", Choice: Foundation, Producer: "p", Consumer: "c", MetaOperation: "persist", Stage: "s", Step: "step", Reason: "r", Persistent: true}
	left := Bundle{Items: []Item{claim}, Transitions: []Transition{transition}}
	joined, err := Combine(left, left)
	if err != nil || len(joined.Items) != 1 || len(joined.Transitions) != 1 {
		t.Fatalf("idempotent combination = %+v, err = %v", joined, err)
	}
	conflict := claim
	conflict.Choice = Regression
	if _, err := Combine(left, Bundle{Items: []Item{conflict}, Transitions: []Transition{transition}}); err == nil {
		t.Fatal("conflicting proof choices were accepted")
	}
}

func ExampleCombine() {
	item := Item{Kind: Metric, ID: "metric.example", Statement: "three slots", Choice: Regression, Producer: "p", Consumer: "c", MetaOperation: "measure", Stage: "replay", Step: "compare", Reason: "baseline", Numerator: 3, Denominator: 3}
	bundle, err := Combine(Bundle{Items: []Item{item}}, Bundle{})
	fmt.Println(len(bundle.Items), err == nil)
	// Output: 1 true
}
