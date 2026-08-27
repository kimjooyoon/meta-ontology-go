package minimalcausalexplanation

import (
	"strings"
	"testing"
)

func fixtureSource() []byte {
	return []byte("package minimalcausalexplanation\nnamespace minimalcausalexplanation\nentity Evidence id \"gooo://evidence\"\nactivity ProduceEvidence(Evidence) -> Evidence\n")
}

func TestEvaluateEmitsMinimalCausalExplanationContract(t *testing.T) {
	receipt, err := Evaluate("examples/minimal-causal-explanation/main.gooo", fixtureSource(), "example/repository", strings.Repeat("a", 40))
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Decision != DecisionPass || receipt.Summary.MinimalSufficientPaths != 1 || receipt.Summary.SufficientNonminimalPaths != 1 || receipt.Summary.InsufficientPaths != 1 {
		t.Fatalf("unexpected explanation result: %+v", receipt.Summary)
	}
	if receipt.Summary.CounterfactualTotal != 7 || receipt.Summary.ChangedCounterfactualTotal != 6 {
		t.Fatalf("unexpected counterfactual result: %+v", receipt.Summary)
	}
	if len(receipt.ClaimTransitions) != TransitionTotal || receipt.Preservation.PreservedTotal != ClaimTotal {
		t.Fatalf("unexpected preservation result: %+v", receipt.Preservation)
	}
}
