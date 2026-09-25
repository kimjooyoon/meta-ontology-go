package adapter

import (
	"encoding/json"
	"testing"
)

func TestOracleNW003RejectsObserverRelabelAndSnapshotTamper(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	request.RunID = "run-002"
	response.RunID = request.RunID
	observation.Binding.RunID = request.RunID
	evaluation := EvaluateObserved(request, response, &observation)
	assertOracleFailure(t, evaluation, OracleNW003)

	request = sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response = sampleResponse(StatusFail, false)
	observer = newStableObserver(t, request)
	observation, err = observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	observation.After.Output.ByteDigest = digestBytes([]byte("forged"))
	evaluation = EvaluateObserved(request, response, &observation)
	assertOracleFailure(t, evaluation, OracleNW003)
}
func TestOracleNW003RejectsWireRoundTripWithoutPrivateStamp(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(observation)
	if err != nil {
		t.Fatal(err)
	}
	var decoded NoWriteObservation
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, sampleResponse(StatusFail, false), &decoded)
	assertOracleFailure(t, evaluation, OracleNW003)
}
