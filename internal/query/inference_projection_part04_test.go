package query

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"testing"
)

func TestInferenceProjectionRejectsInvalidAuthorityAndChains(t *testing.T) {
	path, edges := inferenceQueryFixture(t)
	lift := path
	lift.Edges = append([]semantic.InferenceEdge(nil), path.Edges...)
	lift.Edges[4].AcceptanceReceipt = ""
	invalidResult, err := QueryInferencePath(lift, inferenceQueryRequest())
	if err == nil || invalidResult.Complete || invalidResult.Status != ResponseError || len(invalidResult.Edges) != 0 {
		t.Fatalf("invalid lift result = %#v err=%v", invalidResult, err)
	}
	stale := path
	stale.Edges = append([]semantic.InferenceEdge(nil), path.Edges...)
	stale.Evidence = append([]semantic.InferenceEvidence(nil), path.Evidence...)
	stale.Evidence[0].After.Semantic = inferenceQueryDigest("stale")
	staleResult, err := QueryInferencePath(stale, inferenceQueryRequest())
	if err == nil || staleResult.Complete || !errors.Is(err, semantic.ErrInferencePath) {
		t.Fatalf("stale path result = %#v err=%v", staleResult, err)
	}
	ambiguous := semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion}
	first := edges[0]
	second := edges[1]
	second.SubjectID = first.SubjectID
	for _, edge := range []semantic.InferenceEdge{first, second} {
		evidenceID := edge.Evidence[0].ID
		evidence := semantic.InferenceEvidence{ID: evidenceID, Digest: edge.Evidence[0].Digest, Before: edge.Before, After: edge.After, Controls: edge.Controls}
		ambiguous.Edges = append(ambiguous.Edges, edge)
		ambiguous.Evidence = append(ambiguous.Evidence, evidence)
	}
	ambiguousRequest := inferenceQueryRequest()
	ambiguousRequest.Explain = true
	ambiguousResult, err := QueryInferencePath(ambiguous, ambiguousRequest)
	if err == nil || ambiguousResult.Complete || !errors.Is(err, ErrInferenceChain) || len(ambiguousResult.Edges) != 0 {
		t.Fatalf("ambiguous chain result = %#v err=%v", ambiguousResult, err)
	}
	cycle := semantic.InferencePathV1{Version: semantic.InferencePathSchemaVersion}
	cycleA := edges[0]
	cycleB := edges[1]
	cycleA.SubjectID, cycleA.ObjectID = inferenceQueryID("cycle/a"), inferenceQueryID("cycle/b")
	cycleB.SubjectID, cycleB.ObjectID = inferenceQueryID("cycle/b"), inferenceQueryID("cycle/a")
	for _, edge := range []semantic.InferenceEdge{cycleA, cycleB} {
		evidenceID := edge.Evidence[0].ID
		cycle.Edges = append(cycle.Edges, edge)
		cycle.Evidence = append(cycle.Evidence, semantic.InferenceEvidence{ID: evidenceID, Digest: edge.Evidence[0].Digest, Before: edge.Before, After: edge.After, Controls: edge.Controls})
	}
	cycleRequest := inferenceQueryRequest()
	cycleRequest.Explain = true
	cycleResult, err := QueryInferencePath(cycle, cycleRequest)
	if err == nil || cycleResult.Complete || !errors.Is(err, ErrInferenceChain) {
		t.Fatalf("cycle chain result = %#v err=%v", cycleResult, err)
	}
}
func projectionForPath(t *testing.T, path semantic.InferencePathV1) *InferenceProjection {
	t.Helper()
	projection, err := NewInferenceProjection(path)
	if err != nil {
		t.Fatal(err)
	}
	return projection
}
