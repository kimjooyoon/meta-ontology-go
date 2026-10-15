package replay

import (
	"os"
	"testing"
)

func TestRecordedExplanationSourceCoherence(t *testing.T) {
	repository := os.DirFS("../../../../..")
	result := ExecuteWithOrderedCollections(repository, "examples/recorded-use-origins/inspect.gooo", "VALID", "")
	if result.ObservedDecision != DecisionPass || !result.ASTReplayed || !result.ByteReplayed ||
		!result.SemanticReplayed || !result.GetPut || !result.PutGet {
		t.Fatalf("recorded-use policy and explanation lost V5 source coherence: %+v", result)
	}
	t.Log("new Explain source: AST, bytes, semantics, GetPut and PutGet; no native program/model execution")
}
