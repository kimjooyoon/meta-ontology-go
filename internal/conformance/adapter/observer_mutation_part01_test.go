package adapter

import (
	"testing"
)

func attachVerifiedMutation(t *testing.T, observer *NoWriteObserver, request Request) {
	t.Helper()
	evidence, err := newVerifiedMutationEvidence(MutationEvidence{
		Status: MutationEvidenceVerified, Binding: requestObservationBinding(request),
	}, observer.paths)
	if err != nil {
		t.Fatal(err)
	}
	if err := observer.captureVerifiedMutation(evidence); err != nil {
		t.Fatal(err)
	}
}
func attachVerifiedMutationAttempt(t *testing.T, observer *NoWriteObserver, request Request) {
	t.Helper()
	evidence, err := newVerifiedMutationEvidence(MutationEvidence{
		Status: MutationEvidenceVerified, Binding: requestObservationBinding(request),
		Attempts: []MutationAttempt{{
			Path: observer.paths.SourcePath, Operation: "write", Outcome: "rejected",
		}},
	}, observer.paths)
	if err != nil {
		t.Fatal(err)
	}
	if err := observer.captureVerifiedMutation(evidence); err != nil {
		t.Fatal(err)
	}
}
func TestOracleNW001RejectsMissingMutationAttemptProof(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newBareObserver(t, request)
	attachVerifiedWorkflow(t, observer, request)
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleNW001)
}
func TestOracleNW003RejectsPublicMutationClaim(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newBareObserver(t, request)
	attachVerifiedWorkflow(t, observer, request)
	err := observer.CaptureUnverifiedMutation(MutationEvidence{
		Status: MutationEvidenceUnverified, Binding: requestObservationBinding(request),
		Attempts: []MutationAttempt{{Path: observer.paths.SourcePath, Operation: "write", Outcome: "rejected"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleNW003)
}
