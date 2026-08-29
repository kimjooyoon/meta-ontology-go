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
	root := extractionCounterexample(action.Subject)
	derived := root + "ExtractedSuffix08"
	evidence := generation.ObservationFailureEvidence{
		IndicatorID:    "indicator",
		Observed:       0,
		Expected:       1,
		Decision:       "UNKNOWN",
		Counterexample: derived}
	failure := generation.ObservationFailure{
		ActionIndicatorID: action.IndicatorID,
		Counterexample:    root,
		DerivedRelations: []generation.CounterexampleRelation{{
			Counterexample: derived, DerivedFrom: root, Relation: "DERIVED_FROM",
		}},
		FailureEvidence: []generation.ObservationFailureEvidence{evidence}}
	if !validRefutedIndicatorLinks(failure, action) {
		t.Fatal("stable extraction counterexample was rejected")
	}
	failure.FailureEvidence[0].Counterexample = root + "ExtractedSuffixSibling"
	if validRefutedIndicatorLinks(failure, action) {
		t.Fatal("arbitrary extraction counterexample was accepted")
	}

	failure.FailureEvidence[0].Counterexample = derived
	failure.Stage = "derive-recipe"
	failure.Step = "select-declaration"
	failure.Reason = "NO_SAFE_DECLARATION_CAPACITY"
	failure.NextOperation = "report-counterexample"
	unknown := generation.ReceiptUnknown{
		ActionIndicatorID:   action.IndicatorID,
		RequiredIndicatorID: "indicator",
		Stage:               failure.Stage,
		Step:                failure.Step,
		Reason:              generation.ReceiptReason(failure.Reason),
		UnknownClass:        generation.ReceiptUnknownClassDependencyBlocked,
		NextOperation:       failure.NextOperation,
		BlockedBy:           []string{"operation-failure:" + action.IndicatorID}}
	if !validDependencyUnknowns([]generation.ReceiptUnknown{unknown}, failure, action) {
		t.Fatal("stable dependency frontier was rejected")
	}
	unknown.BlockedBy = []string{"arbitrary-frontier"}
	if validDependencyUnknowns([]generation.ReceiptUnknown{unknown}, failure, action) {
		t.Fatal("arbitrary dependency frontier was accepted")
	}

	failure.DerivedRelations[0].Counterexample = root + "ExtractedSuffix"
	if validRefutedIndicatorLinks(failure, action) {
		t.Fatal("prefix-only derived counterexample was accepted")
	}
}
