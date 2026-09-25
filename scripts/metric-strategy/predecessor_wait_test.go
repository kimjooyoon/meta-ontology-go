package main

import (
	"context"
	"errors"
	"testing"

	predecessor "github.com/kimjooyoon/meta-ontology-go/internal/meta/metricstrategy/proposalpredecessor"
)

func TestPredecessorObservationPolicyCases(t *testing.T) {
	pending := pendingObservationFixture(t)
	ready := predecessor.Report{Schema: predecessor.Schema, Decision: "SELECTED", Reason: predecessor.ReasonSelected}
	missing := predecessor.Report{Schema: predecessor.Schema, Reason: predecessor.ReasonNotFound}
	refuted := predecessor.Report{Schema: predecessor.Schema, ObservationDecision: predecessor.DecisionRefuted}
	tests := []struct {
		name       string
		states     []predecessor.Report
		attempts   int
		cancelWait bool
		nilError   bool
		records    int
		waits      int
		selected   bool
	}{
		{"ready", []predecessor.Report{ready}, 3, false, false, 1, 0, true},
		{"pending-to-ready", []predecessor.Report{pending, ready}, 3, false, false, 2, 1, true},
		{"exhausted", []predecessor.Report{pending}, 3, false, false, 3, 2, false},
		{"missing", []predecessor.Report{missing}, 3, false, false, 1, 0, false},
		{"refuted", []predecessor.Report{refuted}, 3, false, false, 1, 0, false},
		{"canceled", []predecessor.Report{pending}, 3, true, false, 1, 1, false},
		{"unknown-without-error", []predecessor.Report{pending}, 3, false, true, 1, 0, false},
	}
	if len(tests) != 7 {
		t.Fatal("observation policy denominator must remain seven")
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			observed, recorded, waits := 0, 0, 0
			observe := func(context.Context) (predecessor.Report, []byte, error) {
				index := min(observed, len(test.states)-1)
				report := test.states[index]
				observed++
				if report.Ready() || test.nilError {
					return report, []byte("proposal"), nil
				}
				return report, nil, errors.New("not selected")
			}
			record := func(int, predecessor.Report) error { recorded++; return nil }
			wait := func(context.Context) error {
				waits++
				if test.cancelWait {
					return context.Canceled
				}
				return nil
			}
			report, payload, err := awaitProposalPredecessor(context.Background(), test.attempts, observe, record, wait)
			if recorded != test.records || waits != test.waits || (err == nil) != test.selected ||
				(err == nil && !report.Ready()) || (!test.selected && len(payload) != 0) {
				t.Fatalf("records=%d waits=%d report=%+v payload=%q err=%v", recorded, waits, report, payload, err)
			}
		})
	}
}
