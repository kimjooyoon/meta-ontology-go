package adapter

import (
	"reflect"
	"testing"
)

func TestOracleNW004RejectsAttemptWithUnchangedFilesystem(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newBareObserver(t, request)
	attachVerifiedWorkflow(t, observer, request)
	before := observer.before
	attachVerifiedMutationAttempt(t, observer, request)
	observation, err := observer.CaptureRejected(RejectionClosed)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(observation.Before, before) ||
		!reflect.DeepEqual(observation.Before, observation.After) {
		t.Fatal("mutation-attempt fixture changed filesystem state")
	}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleNW004)
}
func TestOracleID001RejectsMutationBindingMismatch(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newBareObserver(t, request)
	attachVerifiedWorkflow(t, observer, request)
	foreignBinding := requestObservationBinding(request)
	foreignBinding.RunID = "stale-run"
	evidence, err := newVerifiedMutationEvidence(MutationEvidence{
		Status: MutationEvidenceVerified, Binding: foreignBinding,
	}, observer.paths)
	if err != nil {
		t.Fatal(err)
	}
	if err := observer.captureVerifiedMutation(evidence); err != nil {
		t.Fatal(err)
	}
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleID001)
}
func TestOracleNW003RejectsMutationEvidenceReplayTamper(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newStableObserver(t, request)
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	observation.Mutation.Attempts = []MutationAttempt{{
		Path: observer.paths.OutputPath, Operation: "write", Outcome: "rejected",
	}}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleNW003)
}
func TestReceiptReplayRejectsMissingMutationAttemptProof(t *testing.T) {
	receipt, request, _ := newCancelledReceipt(t)
	observer := newBareObserver(t, request)
	attachVerifiedWorkflow(t, observer, request)
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	assertReceiptOracleError(t, receipt.ValidateObservedNoWrite(request, &observation), OracleNW001)
}
