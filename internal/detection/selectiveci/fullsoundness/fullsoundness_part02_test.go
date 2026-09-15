package fullsoundness

import (
	"math"
	"testing"
)

func TestClosedReasons(t *testing.T) {
	cases := []struct {
		name     string
		mutate   func(*Input)
		decision Decision
		reason   Reason
	}{
		{"authorization", func(input *Input) { input.ExecutionAuthorized = true }, DecisionUnsound, ReasonAuthorizationPresent},
		{"CI authorization", func(input *Input) { input.CIAuthorized = true }, DecisionUnsound, ReasonAuthorizationPresent},
		{"selected receipt set", func(input *Input) {
			input.SelectionReceipt.CommandIDs = append(input.SelectionReceipt.CommandIDs, id("command/pass"))
		}, DecisionUnsound, ReasonSelectedSetMismatch},
		{"global guard", func(input *Input) { selectOnly(input, []string{id("command/impact")}) }, DecisionUnsound, ReasonGlobalGuardOmitted},
		{"selected extra failure", func(input *Input) {
			selected := findOutcome(input.SelectedOutcomes, id("command/impact"))
			selected.Status = OutcomeFail
		}, DecisionUnsound, ReasonSelectedExtraFailure},
		{"status mismatch", func(input *Input) {
			full := findOutcome(input.FullOutcomes, id("command/impact"))
			full.Status, full.FailureCode = OutcomeFail, "full"
		}, DecisionUnsound, ReasonSelectedFullStatusMismatch},
		{"failure code", func(input *Input) {
			full := findOutcome(input.FullOutcomes, id("command/impact"))
			selected := findOutcome(input.SelectedOutcomes, id("command/impact"))
			full.Status, selected.Status, full.FailureCode, selected.FailureCode = OutcomeFail, OutcomeFail, "full", "selected"
		}, DecisionUnsound, ReasonFailureCodeMismatch},
		{"output digest", func(input *Input) {
			findOutcome(input.SelectedOutcomes, id("command/impact")).OutputDigest = digest("9")
		}, DecisionUnsound, ReasonOutputDigestMismatch},
		{"omitted full failure", func(input *Input) {
			selectOnly(input, []string{id("command/guard")})
			full := findOutcome(input.FullOutcomes, id("command/impact"))
			full.Status, full.FailureCode = OutcomeFail, "failed"
		}, DecisionUnsound, ReasonOmittedFullFailure},
		{"impacted omitted", func(input *Input) { selectOnly(input, []string{id("command/guard")}) }, DecisionUnsound, ReasonImpactedCommandOmitted},
		{"missing selected receipt", func(input *Input) { input.SelectedResourceReceipts = nil }, DecisionUnknown, ReasonFullSuiteRequired},
		{"missing full receipt", func(input *Input) { input.FullResourceReceipts = nil }, DecisionUnknown, ReasonFullSuiteRequired},
		{"missing selection receipt", func(input *Input) { input.SelectionReceipt = nil }, DecisionUnknown, ReasonFullSuiteRequired},
		{"duplicate command obligation", func(input *Input) {
			input.Commands[1].ObligationIDs = append(input.Commands[1].ObligationIDs, input.Commands[1].ObligationIDs[0])
		}, DecisionUnknown, ReasonFullSuiteRequired},
		{"unregistered obligation", func(input *Input) {
			input.Commands[1].ObligationIDs = append(input.Commands[1].ObligationIDs, id("obligation/missing"))
		}, DecisionUnknown, ReasonUnregisteredObligation},
		{"unprovable obligation", func(input *Input) { input.Obligations[0].Authority = AuthorityCandidate }, DecisionUnknown, ReasonUnprovableObligation},
		{"zero commands", zeroCommandInput, DecisionUnknown, ReasonZeroCommandDenominator},
		{"invalid outcome", func(input *Input) { findOutcome(input.FullOutcomes, id("command/impact")).Status = "OTHER" }, DecisionUnknown, ReasonInvalidOutcome},
		{"digest binding", func(input *Input) { input.SelectionReceipt.SnapshotDigest = digest("0") }, DecisionUnknown, ReasonDigestBindingMismatch},
		{"resource overflow", func(input *Input) { input.FullResourceReceipts[0].CPUCoreNS = math.MaxInt64 }, DecisionUnknown, ReasonResourceOverflow},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			input := soundInput()
			test.mutate(&input)
			got := Evaluate(input)
			if got.Decision != test.decision || got.Reason != test.reason {
				t.Fatalf("got %s/%s, want %s/%s", got.Decision, got.Reason, test.decision, test.reason)
			}
			if got.ExecutionAuthorized || got.CIAuthorized {
				t.Fatalf("result authorized execution: %#v", got)
			}
		})
	}
}
