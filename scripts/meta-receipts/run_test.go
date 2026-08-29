package main

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/generation"
)

func TestMixedRefutationRejectsUnboundCounterexampleAndFrontier(t *testing.T) {
	action := generation.Action{
		IndicatorID:          "sha256:" + strings.Repeat("0", 64),
		Subject:              "fixture.go:10:Selected",
		RequiredIndicatorIDs: []string{"indicator"},
	}
	failure := generation.ObservationFailure{
		ActionIndicatorID: action.IndicatorID,
		FailureEvidence:   []generation.ObservationFailureEvidence{{
			IndicatorID: "indicator", Observed: 0, Expected: 1,
			Decision: "UNKNOWN", Counterexample: "fixture.go#func:Selected",
		}},
	}
	if !validRefutedIndicatorLinks(failure, action) {
		t.Fatal("stable extraction counterexample was rejected")
	}
	failure.FailureEvidence[0].Counterexample = "arbitrary-counterexample"
	if validRefutedIndicatorLinks(failure, action) {
		t.Fatal("arbitrary extraction counterexample was accepted")
	}

	failure.FailureEvidence[0].Counterexample = "fixture.go#func:Selected"
	failure.Stage, failure.Step, failure.Reason, failure.NextOperation =
		"derive-recipe", "select-declaration", "NO_SAFE_DECLARATION_CAPACITY", "report-counterexample"
	unknown := generation.ReceiptUnknown{
		ActionIndicatorID: action.IndicatorID, RequiredIndicatorID: "indicator",
		Stage: failure.Stage, Step: failure.Step, Reason: generation.ReceiptReason(failure.Reason),
		UnknownClass: generation.ReceiptUnknownClassDependencyBlocked,
		NextOperation: failure.NextOperation, BlockedBy: []string{"operation-failure:" + action.IndicatorID},
	}
	if !validDependencyUnknowns([]generation.ReceiptUnknown{unknown}, failure, action) {
		t.Fatal("stable dependency frontier was rejected")
	}
	unknown.BlockedBy = []string{"arbitrary-frontier"}
	if validDependencyUnknowns([]generation.ReceiptUnknown{unknown}, failure, action) {
		t.Fatal("arbitrary dependency frontier was accepted")
	}
}
