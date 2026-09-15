package adapter

import (
	"os"
	"strings"
	"testing"
)

func TestPublicMutationVerifiedStatusCannotForgeObserverProof(t *testing.T) {
	request := sampleRequest(StatusFail)
	observer := newBareObserver(t, request)
	err := observer.CaptureUnverifiedMutation(MutationEvidence{
		Status: MutationEvidenceVerified, Binding: requestObservationBinding(request),
	})
	if err == nil || !strings.Contains(err.Error(), OracleNW003) {
		t.Fatalf("verified public mutation claim was accepted: %v", err)
	}
	if _, err := observer.CaptureRejected(RejectionCancelled); err != nil {
		t.Fatal(err)
	}
}
func TestMutationAttemptCaptureRejectsForeignPath(t *testing.T) {
	request := sampleRequest(StatusFail)
	observer := newBareObserver(t, request)
	outside := observer.paths.TempRoot + ".outside"
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(outside)
	_, err := newVerifiedMutationEvidence(MutationEvidence{
		Status: MutationEvidenceVerified, Binding: requestObservationBinding(request),
		Attempts: []MutationAttempt{{Path: outside, Operation: "write", Outcome: "rejected"}},
	}, observer.paths)
	if err == nil {
		t.Fatal("foreign mutation path was accepted")
	}
}
