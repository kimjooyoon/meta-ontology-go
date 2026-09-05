package main

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func TestExtractorDiagnosticTransportCohort(t *testing.T) {
	for _, kind := range []string{"types-check", "type-string"} {
		t.Run(kind, func(t *testing.T) {
			details := []string{"evidence=" + kind, "detail=unavailable type evidence"}
			record := extractorFailureRecord{
				Logical: "sample.go", Decision: "UNKNOWN", Stage: "derive-recipe",
				Step: "type-check-suffix", Reason: "TYPE_EVIDENCE_MISSING",
				UnknownClass: "DIRECT_MISSING", NextOperation: "restore-type-evidence",
				BlockedBy: []string{}, Diagnostics: details,
			}
			cause, ok := reportFailureOperationError([]extractorFailureRecord{record})
			if !ok || cause == nil {
				t.Fatal("extractor failure was not selected")
			}
			action := generation.Action{IndicatorID: "diagnostic-action"}
			failure := observationFailureFromError(action, cause, generation.ProcessObservation{})
			cause.diagnostics[0] = "mutated-caller"
			if !reflect.DeepEqual(failure.Diagnostics, details) {
				t.Fatal("diagnostics were lost or aliased in transport")
			}
			if failure.Decision != record.Decision || failure.Stage != record.Stage ||
				failure.Step != record.Step || failure.Reason != record.Reason ||
				failure.UnknownClass != record.UnknownClass || failure.NextOperation != record.NextOperation ||
				!reflect.DeepEqual(failure.BlockedBy, record.BlockedBy) {
				t.Fatal("diagnostic transport changed the causal fields")
			}
			bundle := generation.SealObservationBundle(generation.OperationObservationBundle{
				Failures: []generation.ObservationFailure{failure}, ObservationTotal: 1,
			})
			payload, err := generation.EncodeObservationBundle(bundle)
			if err != nil {
				t.Fatal(err)
			}
			var decoded generation.OperationObservationBundle
			if err := json.Unmarshal(payload, &decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded.Failures[0].Diagnostics, details) {
				t.Fatal("encoded observation did not preserve diagnostics")
			}
		})
	}
}
