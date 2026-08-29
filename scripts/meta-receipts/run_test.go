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
	evidence := generation.ObservationFailureEvidence{}
	evidence.IndicatorID = "indicator"
	evidence.Observed = 0
	evidence.Expected = 1
	evidence.Decision = "UNKNOWN"
	evidence.Counterexample = "fixture.go#func:Selected"
	failure := generation.ObservationFailure{}
	failure.ActionIndicatorID = action.IndicatorID
	failure.FailureEvidence = []generation.ObservationFailureEvidence{evidence}
	if !validRefutedIndicatorLinks(failure, action) {
		t.Fatal("stable extraction counterexample was rejected")
	}
	failure.FailureEvidence[0].Counterexample = "arbitrary-counterexample"
	if validRefutedIndicatorLinks(failure, action) {
		t.Fatal("arbitrary extraction counterexample was accepted")
	}

	failure.FailureEvidence[0].Counterexample = "fixture.go#func:Selected"
	failure.Stage = "derive-recipe"
	failure.Step = "select-declaration"
	failure.Reason = "NO_SAFE_DECLARATION_CAPACITY"
	failure.NextOperation = "report-counterexample"
	unknown := generation.ReceiptUnknown{}
	unknown.ActionIndicatorID = action.IndicatorID
	unknown.RequiredIndicatorID = "indicator"
	unknown.Stage = failure.Stage
	unknown.Step = failure.Step
	unknown.Reason = generation.ReceiptReason(failure.Reason)
	unknown.UnknownClass = generation.ReceiptUnknownClassDependencyBlocked
	unknown.NextOperation = failure.NextOperation
	unknown.BlockedBy = []string{"operation-failure:" + action.IndicatorID}
	if !validDependencyUnknowns([]generation.ReceiptUnknown{unknown}, failure, action) {
		t.Fatal("stable dependency frontier was rejected")
	}
	unknown.BlockedBy = []string{"arbitrary-frontier"}
	if validDependencyUnknowns([]generation.ReceiptUnknown{unknown}, failure, action) {
		t.Fatal("arbitrary dependency frontier was accepted")
	}
}
