package generation

import (
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestVerifyReceiptsFailsClosed(t *testing.T) {
	plan := actionableReceiptPlan()
	complete := passingReceipts(plan)
	missingEvidence := append(
		[]IndicatorReceipt{}, complete[0].Indicators[1:]...)
	missing := append([]OperationReceipt{}, complete...)
	missing[0] = SealReceipt(plan, plan.Selected[0], missingEvidence)
	missingReport := VerifyReceipts(plan, missing)
	if missingReport.Decision != ReceiptDecisionUnknown ||
		missingReport.Reason != ReceiptReasonMissingIndicator {
		t.Fatalf("missing evidence did not fail closed: %+v", missingReport)
	}

	failedEvidence := append(
		[]IndicatorReceipt{}, complete[0].Indicators...)
	failedEvidence[0].Verdict = IndicatorVerdictFail
	failed := append([]OperationReceipt{}, complete...)
	failed[0] = SealReceipt(plan, plan.Selected[0], failedEvidence)
	failedReport := VerifyReceipts(plan, failed)
	if failedReport.Decision != ReceiptDecisionRejected ||
		failedReport.Reason != ReceiptReasonRejectedIndicator {
		t.Fatalf("explicit failure was not rejected: %+v", failedReport)
	}

	tampered := append([]OperationReceipt{}, complete...)
	tampered[0].Evaluator = "unbound-evaluator"
	tamperedReport := VerifyReceipts(plan, tampered)
	if tamperedReport.Decision != ReceiptDecisionUnknown ||
		tamperedReport.Reason != ReceiptReasonUnknownIndicator {
		t.Fatalf("tampered receipt was not unknown: %+v", tamperedReport)
	}
}

func TestVerifyReceiptsCarriesMissingOperationBindings(t *testing.T) {
	plan := actionableReceiptPlan()
	report := VerifyReceipts(plan, nil)
	want := 0
	for _, action := range plan.Selected {
		want += len(action.RequiredIndicatorIDs)
	}
	if report.Decision != ReceiptDecisionUnknown ||
		report.Reason != ReceiptReasonMissingIndicator || len(report.Unknowns) != want {
		t.Fatalf("missing receipt evidence = %+v, want %d unknown bindings", report, want)
	}
	for _, unknown := range report.Unknowns {
		if unknown.ActionIndicatorID == "" || unknown.RequiredIndicatorID == "" ||
			unknown.Operation == "" || unknown.Activity == "" || unknown.Output == "" ||
			unknown.Executor == "" || unknown.Evaluator == "" ||
			unknown.Reason != ReceiptReasonMissingIndicator ||
			unknown.UnknownClass != ReceiptUnknownClassDirectMissing ||
			unknown.NextOperation != unknown.Executor || len(unknown.BlockedBy) != 0 {
			t.Fatalf("missing receipt binding is not typed: %+v", unknown)
		}
	}
}

func TestVerifyReceiptsRecognizesExactFixedPoint(t *testing.T) {
	report := sourcepolicy.Report{
		Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{
			metric("floor", sourcepolicy.OperationSplitGo, true, true),
		},
	}
	plan := Build(strings.Repeat("3", 40), strings.Repeat("4", 40), report)
	verified := VerifyReceipts(plan, nil)
	if verified.Decision != ReceiptDecisionFixedPoint ||
		verified.Reason != ReceiptReasonExactFixedPoint ||
		verified.PromotionAuthorized {
		t.Fatalf("unexpected fixed-point receipt report: %+v", verified)
	}
}
