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

func TestVerifyReceiptsRecognizesExactFixedPoint(t *testing.T) {
	report := sourcepolicy.Report{
		Policy: sourcepolicy.Default(),
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
