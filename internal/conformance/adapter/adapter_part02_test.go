package adapter

import (
	"strings"
	"testing"
)

func TestCanonicalValidatorRejectsProtectedBoundaryCollisions(t *testing.T) {
	response := sampleResponse(StatusPass, false)
	response.Observed.Regions = append(response.Observed.Regions, Region{
		Kind:       "generated",
		SemanticID: "billing.total",
		Start:      40,
		End:        41,
	})
	if _, err := response.CanonicalJSON(); err == nil || !strings.Contains(err.Error(), "duplicate region") {
		t.Fatalf("duplicate protected boundary was accepted: %v", err)
	}
	response = sampleResponse(StatusPass, false)
	response.Observed.Regions[0].Start = 10
	response.Observed.Regions[0].End = 9
	if _, err := response.CanonicalJSON(); err == nil || !strings.Contains(err.Error(), "invalid range") {
		t.Fatalf("inverted range was accepted: %v", err)
	}
	response = sampleResponse(StatusPass, false)
	response.Observed.Delta = &Delta{Added: []Fact{{
		SubjectID: "billing.total", Predicate: "prov:wasDerivedFrom", ObjectID: "source",
		Class: "prov:Entity", SourceURI: "../outside.gooo", Start: 1, End: 2,
	}}}
	if _, err := response.CanonicalJSON(); err == nil || !strings.Contains(err.Error(), "escape") {
		t.Fatalf("escaping source URI was accepted: %v", err)
	}
}
func TestOracleNW001NegativeOracleRequiresIndependentObserver(t *testing.T) {
	request := sampleRequest(StatusFail)
	request.Expected.FailureCode = "marker-overlap"
	response := sampleResponse(StatusFail, false)
	claim := true
	response.ProducerClaims.NoWrite = &claim
	evaluation := Evaluate(request, response)
	if evaluation.Matched || evaluation.OracleCode != OracleNW001 || evaluation.PromotionEligible {
		t.Fatalf("producer-only negative result was accepted: %+v", evaluation)
	}
	evaluation = EvaluateObserved(request, response, nil)
	if evaluation.Matched || evaluation.OracleCode != OracleNW001 || evaluation.PromotionEligible {
		t.Fatalf("forged producer claim reached observed evaluator: %+v", evaluation)
	}
	observer := newStableObserver(t, request)
	observation, err := observer.Finish()
	if err != nil {
		t.Fatal(err)
	}
	evaluation = EvaluateObserved(request, response, &observation)
	if !evaluation.Matched || evaluation.ExitCode != ExitOK || evaluation.FailureCode != "marker-overlap" || evaluation.PromotionEligible {
		t.Fatalf("valid independent negative result was rejected: %+v", evaluation)
	}
}
func TestRequestValidationRequiresRunBinding(t *testing.T) {
	request := sampleRequest(StatusPass)
	request.RunID = ""
	if err := request.Validate(); err == nil || !strings.Contains(err.Error(), "run_id") {
		t.Fatalf("blank run binding was accepted: %v", err)
	}
}
