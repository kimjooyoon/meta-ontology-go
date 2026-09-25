package generation

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kimjooyoon/meta-ontology-go/internal/meta/sourcepolicy"
)

func TestVerifyReceiptsIsDeterministicAndNonAuthorizing(t *testing.T) {
	plan := actionableReceiptPlan()
	receipts := passingReceipts(plan)
	reversed := []OperationReceipt{receipts[1], receipts[0]}
	first := VerifyReceipts(plan, reversed)
	replay := VerifyReceipts(plan, receipts)
	if !reflect.DeepEqual(first, replay) {
		t.Fatal("receipt verification did not replay deterministically")
	}
	if first.Decision != ReceiptDecisionConformant ||
		first.Reason != ReceiptReasonVerified ||
		first.ReplayDigest == "" {
		t.Fatalf("unexpected receipt report: %+v", first)
	}
	if first.PromotionAuthorized ||
		first.PromotionAuthorizedByReceipts() {
		t.Fatal("receipt evidence acquired promotion authority")
	}
	for _, receipt := range first.Receipts {
		if !validDigest(receipt.ReceiptDigest) {
			t.Fatalf("receipt is not sealed: %+v", receipt)
		}
		if receipt.Operation == "" || receipt.Activity == "" || receipt.Output == "" || receipt.Executor == "" || receipt.Evaluator == "" {
			t.Fatalf("receipt lost operation binding: %+v", receipt)
		}
	}
}

func actionableReceiptPlan() Plan {
	report := sourcepolicy.Report{
		Schema: sourcepolicy.IndicatorSchema, Policy: sourcepolicy.Default(),
		Indicators: []sourcepolicy.Indicator{
			metric("expression", sourcepolicy.OperationCollapseAssign, false, false),
			metric("topology", sourcepolicy.OperationSplitGo, false, false),
		},
	}
	return Build(strings.Repeat("1", 40), strings.Repeat("2", 40), report)
}

func passingReceipts(plan Plan) []OperationReceipt {
	result := make([]OperationReceipt, 0, len(plan.Selected))
	for _, action := range plan.Selected {
		indicators := make([]IndicatorReceipt, 0, len(action.RequiredIndicatorIDs))
		for _, identifier := range action.RequiredIndicatorIDs {
			indicators = append(indicators, IndicatorReceipt{
				ID: identifier, Verdict: IndicatorVerdictPass,
				EvidenceDigest: digestJSON([]string{action.IndicatorID, identifier}),
				ProofChoice:    action.ProofChoice,
			})
		}
		result = append(result, SealReceipt(plan, action, indicators))
	}
	return result
}
