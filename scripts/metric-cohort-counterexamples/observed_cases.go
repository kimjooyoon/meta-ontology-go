package main

import (
	co "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricprogram/conceptoperation"
)

const outsideObservedOperation1889 = "urn:cohort:unregistered-observation-1889"

type observedCasePlan struct {
	ID       string
	Expected co.ObservationAssessment
}

type observedCaseResult struct {
	ID                         string                   `json:"id"`
	Expected                   co.ObservationAssessment `json:"expected"`
	Receipt                    *co.ObservedReceipt      `json:"receipt,omitempty"`
	FirstConsumerReason        string                   `json:"first_consumer_reason"`
	ReplayConsumerReason       string                   `json:"replay_consumer_reason"`
	SourceReplayed             bool                     `json:"source_replayed"`
	ReadinessCredit            bool                     `json:"readiness_credit"`
	UnrelatedObligationsIntact *bool                    `json:"unrelated_obligations_intact,omitempty"`
	Passed                     bool                     `json:"passed"`
}

type observedCohortReport struct {
	Decision      string               `json:"decision"`
	CasesDeclared int                  `json:"cases_declared"`
	CasesPassed   int                  `json:"cases_passed"`
	NativeRuns    int                  `json:"native_runs"`
	NativePassed  int                  `json:"native_passed"`
	ModelCalls    int                  `json:"model_calls"`
	SetupError    string               `json:"setup_error,omitempty"`
	Cases         []observedCaseResult `json:"cases"`
}

func observedPlans1889(expected int) []observedCasePlan {
	missing := (expected - 1) * 10000 / expected
	rows := []observedCasePlan{
		{ID: "complete", Expected: co.ObservationAssessment{BoundCount: expected, CoverageBPS: 10000}},
		{ID: "extra-unregistered", Expected: co.ObservationAssessment{BoundCount: expected, UnknownCount: 1, CoverageBPS: 10000}},
		{ID: "missing", Expected: co.ObservationAssessment{BoundCount: expected - 1, CoverageBPS: missing}},
		{ID: "replaced", Expected: co.ObservationAssessment{BoundCount: expected - 1, UnknownCount: 1, CoverageBPS: missing}},
		{ID: "duplicate", Expected: co.ObservationAssessment{BoundCount: expected - 1, UnknownCount: 1, CoverageBPS: missing}},
		{ID: "empty", Expected: co.ObservationAssessment{}},
	}
	for index := range rows {
		r := &rows[index].Expected
		r.ExpectedCount, r.Status = expected, "FAIL_CLOSED"
		r.UnknownClass, r.Reason = "DIRECT_MISSING", "CONCEPT_OPERATION_DIRECT_EVIDENCE_MISSING"
		r.NextOperation = "CAPTURE_EXACT_PRODUCER_INPUTS"
		if r.UnknownCount > 0 {
			r.UnknownClass, r.Reason = "AMBIGUOUS", "CONCEPT_OPERATION_BINDING_AMBIGUOUS"
			r.NextOperation = "DISAMBIGUATE_CONCEPT_OPERATION_BINDINGS"
		}
		if rows[index].ID == "complete" {
			r.Status, r.UnknownClass, r.Reason, r.NextOperation = "VERIFIED", "", "", ""
		}
	}
	return rows
}

func reviseObserved1889(document *co.ObservationDocument, kind string) {
	switch kind {
	case "extra-unregistered":
		document.Observations = append(document.Observations, co.OperationObservation{Operation: outsideObservedOperation1889})
	case "missing":
		document.Observations = document.Observations[:len(document.Observations)-1]
	case "replaced":
		document.Observations[len(document.Observations)-1] = co.OperationObservation{Operation: outsideObservedOperation1889}
	case "duplicate":
		document.Observations = append(document.Observations, document.Observations[0])
	case "empty":
		document.Observations = []co.OperationObservation{}
	}
}
