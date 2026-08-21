package adapter

import (
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
