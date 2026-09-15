package query

import (
	"errors"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func TestInferenceProjectionRejectsCollisionsOrphansAndStale(t *testing.T) {
	path, _ := inferenceQueryFixture(t)
	duplicateRecord := path
	duplicateRecord.Edges = append([]semantic.InferenceEdge(nil), path.Edges...)
	duplicateRecord.Edges[1].RecordID = duplicateRecord.Edges[0].RecordID
	duplicateResult, err := QueryInferencePath(duplicateRecord, inferenceQueryRequest())
	if err == nil || duplicateResult.Complete || !errors.Is(err, semantic.ErrInferencePath) {
		t.Fatalf("duplicate stable ID result = %#v err=%v", duplicateResult, err)
	}
	duplicateEvidence := path
	duplicateEvidence.Evidence = append([]semantic.InferenceEvidence(nil), path.Evidence...)
	duplicateEvidence.Evidence[1].ID = duplicateEvidence.Evidence[0].ID
	duplicateEvidenceResult, err := QueryInferencePath(duplicateEvidence, inferenceQueryRequest())
	if err == nil || duplicateEvidenceResult.Complete || !errors.Is(err, semantic.ErrInferencePath) {
		t.Fatalf("duplicate evidence ID result = %#v err=%v", duplicateEvidenceResult, err)
	}
	orphanEvidence := path
	orphanEvidence.Evidence = append([]semantic.InferenceEvidence(nil), path.Evidence[1:]...)
	orphanResult, orphanErr := QueryInferencePath(orphanEvidence, inferenceQueryRequest())
	if orphanErr == nil || orphanResult.Complete || !errors.Is(orphanErr, semantic.ErrInferencePath) {
		t.Fatalf("orphan evidence result = %#v err=%v", orphanResult, orphanErr)
	}
	collision := path
	collision.Edges = append([]semantic.InferenceEdge(nil), path.Edges...)
	collision.Evidence = append([]semantic.InferenceEvidence(nil), path.Evidence...)
	collision.Evidence[0].ID = collision.Edges[0].RecordID
	collisionResult, collisionErr := QueryInferencePath(collision, inferenceQueryRequest())
	if collisionErr == nil || collisionResult.Complete || !errors.Is(collisionErr, semantic.ErrInferencePath) {
		t.Fatalf("cross-kind collision result = %#v err=%v", collisionResult, collisionErr)
	}
	orphanChainRequest := inferenceQueryRequest()
	orphanChainRequest.Explain = true
	orphanChainRequest.ChainStartID = ID(inferenceQueryID("missing-chain-start").String())
	orphanChainResult, orphanChainErr := projectionForPath(t, path).Execute(orphanChainRequest)
	if orphanChainErr == nil || orphanChainResult.Complete || !errors.Is(orphanChainErr, ErrInferenceChain) {
		t.Fatalf("orphan chain result = %#v err=%v", orphanChainResult, orphanChainErr)
	}
	staleRequest := inferenceQueryRequest()
	staleRequest.RecordID = ID(path.Edges[1].RecordID.String())
	staleRequest.Controls = semantic.InferenceControls{PolicyDigest: inferenceQueryDigest("wrong-policy")}
	staleResult, staleErr := projectionForPath(t, path).Execute(staleRequest)
	if staleErr == nil || staleResult.Complete || !errors.Is(staleErr, ErrInferenceStaleSnapshot) {
		t.Fatalf("stale control request = %#v err=%v", staleResult, staleErr)
	}
}
