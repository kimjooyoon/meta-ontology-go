package adapter

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

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
