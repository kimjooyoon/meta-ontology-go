package adapter

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type rejectedTransactionFixture struct {
	Name           string        `json:"name"`
	Reason         RejectionKind `json:"reason"`
	PreservesState bool          `json:"preserves_state"`
}

func TestRejectedTransactionFixturesPreserveFilesystemState(t *testing.T) {
	fixtures := loadRejectedTransactionFixtures(t)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			request := sampleRequest(StatusFail)
			request.Expected.FailureCode = "marker-overlap"
			observer := newStableObserver(t, request)
			before, err := captureState(observer.paths)
			if err != nil {
				t.Fatal(err)
			}
			if fixture.Reason == "invalid" {
				if _, err := observer.CaptureRejected(fixture.Reason); err == nil {
					t.Fatal("invalid rejection kind was accepted")
				}
				assertStateUnchanged(t, observer.paths, before)
				return
			}
			observation, err := observer.CaptureRejected(fixture.Reason)
			if err != nil {
				t.Fatal(err)
			}
			if observation.Reason != fixture.Reason {
				t.Fatalf("rejection reason was not bound: %q", observation.Reason)
			}
			evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
			if !evaluation.Matched || evaluation.PromotionEligible {
				t.Fatalf("rejected transaction was not accepted as non-promotable: %+v", evaluation)
			}
			if fixture.PreservesState {
				assertStateUnchanged(t, observer.paths, before)
			}
		})
	}
}

func TestRejectedTransactionCloseCannotBeReopened(t *testing.T) {
	request := sampleRequest(StatusFail)
	observer := newStableObserver(t, request)
	before, err := captureState(observer.paths)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := observer.CaptureRejected(RejectionClosed); err != nil {
		t.Fatal(err)
	}
	if _, err := observer.CaptureRejected(RejectionCancelled); err == nil {
		t.Fatal("closed observer accepted a second rejection")
	}
	assertStateUnchanged(t, observer.paths, before)
}

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

func TestOracleID001RejectsRejectedBindingRelabels(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func(*ObservationBinding)
	}{
		{name: "fixture", edit: func(binding *ObservationBinding) { binding.Fixture = "other" }},
		{name: "operation", edit: func(binding *ObservationBinding) { binding.Operation = OperationParseAST }},
		{name: "run", edit: func(binding *ObservationBinding) { binding.RunID = "other-run" }},
	} {
		t.Run(test.name, func(t *testing.T) {
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
			test.edit(&observation.Binding)
			evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
			assertOracleFailure(t, evaluation, OracleID001)
			assertStateUnchanged(t, observer.paths, before)
		})
	}
}

func loadRejectedTransactionFixtures(t *testing.T) []rejectedTransactionFixture {
	t.Helper()
	data, err := os.ReadFile("testdata/no-write/rejected-transactions.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures []rejectedTransactionFixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	return fixtures
}

func assertStateUnchanged(t *testing.T, paths ObserverPaths, before FilesystemState) {
	t.Helper()
	after, err := captureState(paths)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatalf("rejected operation changed filesystem state:\nbefore=%+v\nafter=%+v", before, after)
	}
}
