package verify

import (
	"strings"
	"testing"

	explanation "github.com/kimjooyoon/meta-ontology-go/internal/meta/minimalcausalexplanation"
)

func TestJudgeUsesPathSetInsteadOfExplanationText(t *testing.T) {
	receipt, err := explanation.Evaluate("examples/minimal-causal-explanation/main.gooo", []byte("package minimalcausalexplanation\nnamespace minimalcausalexplanation\nentity Evidence id \"gooo://evidence\"\nactivity ProduceEvidence(Evidence) -> Evidence\n"), "example/repository", strings.Repeat("b", 40))
	if err != nil {
		t.Fatal(err)
	}
	receipt.Cases[0].ExplanationText = "arbitrary prose cannot add evidence"
	receipt.ReceiptDigest = explanation.ReceiptDigest(receipt)
	judgment, err := Judge(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if judgment.Status != "VERIFIED" || !judgment.PathSetVerified || !judgment.CounterfactualsVerified {
		t.Fatalf("unexpected judgment: %+v", judgment)
	}
}

func TestJudgeRejectsAPathWithRemovedCausalEvidence(t *testing.T) {
	receipt, err := explanation.Evaluate("examples/minimal-causal-explanation/main.gooo", []byte("package minimalcausalexplanation\nnamespace minimalcausalexplanation\nentity Evidence id \"gooo://evidence\"\nactivity ProduceEvidence(Evidence) -> Evidence\n"), "example/repository", strings.Repeat("c", 40))
	if err != nil {
		t.Fatal(err)
	}
	receipt.Cases[0].Paths[0].EvidenceIDs = receipt.Cases[0].Paths[0].EvidenceIDs[:2]
	receipt.ReceiptDigest = explanation.ReceiptDigest(receipt)
	if _, err := Judge(receipt); err == nil {
		t.Fatal("judge accepted an insufficient path")
	}
}
