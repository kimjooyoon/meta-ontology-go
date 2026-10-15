package replay

import (
	"os"
	"testing"
)

func TestRecordedRDFSourceCoherence(t *testing.T) {
	result := ExecuteWithOrderedCollections(os.DirFS("../../../../.."),
		"examples/recorded-use-origins/inspect.gooo", "VALID", "")
	if result.ObservedDecision != DecisionPass || !result.ASTReplayed || !result.ByteReplayed ||
		!result.SemanticReplayed || !result.GetPut || !result.PutGet {
		t.Fatalf("source-owned RDF identities lost V5 coherence: %+v", result)
	}
	t.Log("new RDF source: AST, bytes, semantics, GetPut and PutGet; no native/model execution")
}
