package adapter

import (
	"os"
	"testing"
)

func TestOracleNW003RejectsRejectedReasonTamperWithoutWriting(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newStableObserver(t, request)
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	before, err := captureState(observer.paths)
	if err != nil {
		t.Fatal(err)
	}
	observation.Reason = RejectionClosed
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleNW003)
	assertStateUnchanged(t, observer.paths, before)
}
func TestRejectedEvidenceRejectionDoesNotWrite(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newStableObserver(t, request)
	observation, err := observer.CaptureRejected(RejectionClosed)
	if err != nil {
		t.Fatal(err)
	}
	before, err := captureState(observer.paths)
	if err != nil {
		t.Fatal(err)
	}
	observation.Binding.Fixture = "relabeled"
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleID001)
	assertStateUnchanged(t, observer.paths, before)
}
func TestOracleNW002StaleRejectionDoesNotWrite(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newStableObserver(t, request)
	observation, err := observer.CaptureRejected(RejectionCancelled)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(observer.paths.OutputPath, []byte("stale"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, err := captureState(observer.paths)
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleNW002)
	assertStateUnchanged(t, observer.paths, before)
}
