package replay

import (
	"os"
	"testing"
)

func TestInputDeclarationsSourceCoherence(t *testing.T) {
	r := ExecuteWithOrderedCollections(os.DirFS("../../../../.."),
		"examples/source-control-reader/inspect.gooo", "VALID", "")
	if r.ObservedDecision != DecisionPass || !r.ASTReplayed || !r.ByteReplayed ||
		!r.SemanticReplayed || !r.GetPut || !r.PutGet {
		t.Fatalf("input declaration source coherence: %+v", r)
	}
}
