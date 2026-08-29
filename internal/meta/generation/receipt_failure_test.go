package generation

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestVerifyReceiptsPreservesRefutedPrecedenceAndUnknownFields(t *testing.T) {
	plan := actionableReceiptPlan()
	process := ProcessObservation{
		Command: []string{"<not-executed>", "extract-function"}, ExitCode: -1,
		StdoutDigest: "sha256:" + strings.Repeat("0", 64),
		StderrDigest: "sha256:" + strings.Repeat("0", 64),
	}
	failures := []ObservationFailure{
		{
			ActionIndicatorID: plan.Selected[0].IndicatorID, Decision: "UNKNOWN",
			Stage: "execute-operation", Step: "run-selected-operation",
			Reason: "INSTANCE_EVIDENCE_UNAVAILABLE",
			UnknownClass: ReceiptUnknownClassDirectMissing,
			NextOperation: "restore-operation-evidence", BlockedBy: []string{}, Executor: process,
		},
		{
			ActionIndicatorID: plan.Selected[1].IndicatorID, Decision: "REFUTED",
			Stage: "evaluate-operation", Step: "validate-selected-operation",
			Reason: "KNOWN_CONTRADICTION", NextOperation: "report-counterexample",
			BlockedBy: []string{}, Executor: process,
		},
	}
	report := VerifyReceiptsWithFailures(plan, nil, failures)
	if report.Decision != ReceiptDecisionRefuted || report.Reason != ReceiptReasonRefutedOperation {
		t.Fatalf("refuted failure was not dominant: %+v", report)
	}
	if len(report.Failures) != 2 || report.Failures[0].Stage == "" || report.Failures[0].Step == "" ||
		report.Failures[0].Reason == "" || report.Failures[0].UnknownClass == "" ||
		report.Failures[0].NextOperation == "" || report.Failures[0].BlockedBy == nil {
		t.Fatalf("unknown failure lost its six-field evidence: %+v", report.Failures)
	}
}

func TestVerifyReceiptsRejectsUnknownFailureWithoutSixFields(t *testing.T) {
	plan := sourcepolicyPlanForFailureTest()
	process := ProcessObservation{Command: []string{"<not-executed>"}, ExitCode: -1,
		StdoutDigest: "sha256:" + strings.Repeat("0", 64), StderrDigest: "sha256:" + strings.Repeat("0", 64)}
	failure := ObservationFailure{ActionIndicatorID: plan.Selected[0].IndicatorID, Decision: "UNKNOWN",
		UnknownClass: ReceiptUnknownClassDirectMissing, NextOperation: "restore"}
	failure.Executor = process
	report := VerifyReceiptsWithFailures(plan, nil, []ObservationFailure{failure})
	if report.Decision != ReceiptDecisionUnknown || report.Reason != ReceiptReasonUnknownIndicator || len(report.Failures) != 1 {
		t.Fatalf("incomplete unknown failure was accepted: %+v", report)
	}
}

func sourcepolicyPlanForFailureTest() Plan {
	report := sourcepolicy.Report{Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{metric("failure", sourcepolicy.OperationSplitGo, false, false)}}
	return Build(strings.Repeat("7", 40), strings.Repeat("8", 40), report)
}
