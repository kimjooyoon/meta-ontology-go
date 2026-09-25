package semantic

import (
	"errors"
	"strings"
	"testing"
)

func TestInferencePathChainRejectsOrphanAndAmbiguousPaths(t *testing.T) {
	start := MustIdentity("inference-test://chain/start")
	middle := MustIdentity("inference-test://chain/middle")
	end := MustIdentity("inference-test://chain/end")
	first := inferenceEdgeFixture(InferenceDeterministicDerivation, "chain-first")
	first.SubjectID, first.ObjectID = start, middle
	second := inferenceEdgeFixture(InferenceDerivedProjection, "chain-second")
	second.SubjectID, second.ObjectID = middle, end
	chain, err := NewInferencePathChain(second, first)
	if err != nil || len(chain.Edges) != 2 || chain.Edges[0].SubjectID != start {
		t.Fatalf("valid unordered chain was not reconstructed: chain=%#v err=%v", chain, err)
	}
	branch := inferenceEdgeFixture(InferenceDeterministicDerivation, "chain-branch")
	branch.SubjectID, branch.ObjectID = start, end
	_, err = NewInferencePathChain(first, branch)
	if err == nil || !strings.Contains(err.Error(), "path_ambiguity") {
		t.Fatalf("ambiguous chain error = %v", err)
	}
	cycleA := inferenceEdgeFixture(InferenceDeterministicDerivation, "chain-cycle-a")
	cycleA.SubjectID = MustIdentity("inference-test://chain/cycle-a")
	cycleA.ObjectID = MustIdentity("inference-test://chain/cycle-b")
	cycleB := inferenceEdgeFixture(InferenceDeterministicDerivation, "chain-cycle-b")
	cycleB.SubjectID, cycleB.ObjectID = cycleA.ObjectID, cycleA.SubjectID
	_, err = NewInferencePathChain(first, cycleA, cycleB)
	if err == nil || !strings.Contains(err.Error(), "path_orphan") {
		t.Fatalf("orphan chain error = %v", err)
	}
}
func TestInferencePathErrorsAreFailClosed(t *testing.T) {
	bad := InferencePathV1{Version: InferencePathSchemaVersion, Edges: []InferenceEdge{{Kind: "UNKNOWN"}}}
	if err := bad.Validate(); err == nil || !errors.Is(err, ErrInferencePath) {
		t.Fatalf("unknown state error = %v, want ErrInferencePath", err)
	}
}
