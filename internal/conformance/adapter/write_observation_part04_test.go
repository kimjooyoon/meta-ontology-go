package adapter

import (
	"testing"
)

func TestOracleFAIL002RejectsWrongFailureCode(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	response.Failure.Code = "different-failure"
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	evaluation := EvaluateObserved(request, response, &observation)
	assertOracleFailure(t, evaluation, OracleFAIL002)
}
func TestOraclePASS001RejectsProducerOnlyAndInvalidNoWriteClaims(t *testing.T) {
	request := sampleRequest(StatusPass)
	response := sampleResponse(StatusPass, false)
	claim := true
	response.ProducerClaims.NoWrite = &claim
	assertOracleFailure(t, Evaluate(request, response), OraclePASS001)
	invalid := NoWriteObservation{}
	assertOracleFailure(t, EvaluateObserved(request, response, &invalid), OraclePASS001)
}
func TestOracleID001RejectsRequestResponseAndObserverIdentityMismatch(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	response.RunID = "stale-run"
	assertOracleFailure(t, Evaluate(request, response), OracleID001)
	request.RunID = ""
	response.RunID = ""
	assertOracleFailure(t, Evaluate(request, response), OracleID001)
	request.RunID = "run-001"
	response.RunID = request.RunID
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	observation.Binding.Fixture = "other-fixture"
	assertOracleFailure(t, EvaluateObserved(request, response, &observation), OracleID001)
}
func newStableObserver(t *testing.T, request Request) *NoWriteObserver {
	return newObserverWithTempSetup(t, request, nil)
}
