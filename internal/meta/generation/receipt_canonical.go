package generation

import "sort"

func SealReceipt(plan Plan, action Action, indicators []IndicatorReceipt) OperationReceipt {
	ledgerDigest, ledgerCount := planIndicatorDecisionLedgerProvenance(plan)
	receipt := OperationReceipt{
		SchemaVersion: OperationReceiptSchemaVersion,
		BaseSHA:       plan.BaseSHA, HeadSHA: plan.HeadSHA,
		PlanDigest: plan.PlanDigest, ActionIndicatorID: action.IndicatorID,
		IndicatorDecisionLedgerDigest: ledgerDigest,
		IndicatorDecisionLedgerCount:  ledgerCount,
		Operation:                     action.Operation, Activity: action.Activity, Output: action.Output,
		Executor:                      action.Executor, Evaluator: action.Evaluator,
		ProofChoice: action.ProofChoice,
		Indicators:  normalizeIndicatorReceipts(indicators),
	}
	receipt.ReceiptDigest = operationReceiptDigest(receipt)
	return receipt
}

func normalizeIndicatorReceipts(receipts []IndicatorReceipt) []IndicatorReceipt {
	result := append([]IndicatorReceipt{}, receipts...)
	sort.Slice(result, func(i, j int) bool {
		return indicatorReceiptKey(result[i]) < indicatorReceiptKey(result[j])
	})
	return result
}

func normalizeOperationReceipts(receipts []OperationReceipt) []OperationReceipt {
	result := append([]OperationReceipt{}, receipts...)
	for index := range result {
		result[index].Indicators = normalizeIndicatorReceipts(result[index].Indicators)
	}
	sort.Slice(result, func(i, j int) bool {
		return operationReceiptKey(result[i]) < operationReceiptKey(result[j])
	})
	return result
}

func indicatorReceiptKey(receipt IndicatorReceipt) string {
	return receipt.ID + "\x00" + string(receipt.Verdict) + "\x00" + receipt.EvidenceDigest
}

func operationReceiptKey(receipt OperationReceipt) string {
	return receipt.ActionIndicatorID + "\x00" + string(receipt.Operation) +
		"\x00" + receipt.Evaluator + "\x00" + receipt.ReceiptDigest
}

func operationReceiptDigest(receipt OperationReceipt) string {
	unsigned := receipt
	unsigned.ReceiptDigest = ""
	return digestJSON(unsigned)
}

func validDigest(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !('0' <= character && character <= '9') &&
			!('a' <= character && character <= 'f') {
			return false
		}
	}
	return true
}
