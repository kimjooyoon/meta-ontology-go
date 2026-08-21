package adapter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOracleNW002RejectsStaleObserverTrace(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(observation.Paths.OutputPath, []byte("changed after capture"), 0o600); err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, response, &observation)
	assertOracleFailure(t, evaluation, OracleNW002)
}
func TestOracleNW004RejectsSourceAndOutputByteChanges(t *testing.T) {
	for _, test := range []struct {
		name string
		path func(NoWriteObservation) string
	}{
		{name: "ORACLE-NW-004-source", path: func(observation NoWriteObservation) string { return observation.Paths.SourcePath }},
		{name: "ORACLE-NW-004-output", path: func(observation NoWriteObservation) string { return observation.Paths.OutputPath }},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := sampleRequest(StatusFail)
			request.Expected.FailureCode = "marker-overlap"
			observer := newStableObserver(t, request)
			paths := observer.paths
			if err := os.WriteFile(test.path(NoWriteObservation{Paths: paths}), []byte("different bytes"), 0o600); err != nil {
				t.Fatal(err)
			}
			observation, err := observer.Finish()
			if err != nil {
				t.Fatal(err)
			}
			evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
			assertOracleFailure(t, evaluation, OracleNW004)
		})
	}
}
func TestOracleNW005RejectsSameByteReplacementByLstat(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newStableObserver(t, request)
	original, err := os.ReadFile(observer.paths.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	replacement := filepath.Join(filepath.Dir(observer.paths.OutputPath), "replacement.go")
	if err := os.WriteFile(replacement, original, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(replacement, observer.paths.OutputPath); err != nil {
		t.Fatal(err)
	}
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &observation)
	assertOracleFailure(t, evaluation, OracleNW005)
}
