package generation

func VerifyReceipts(plan Plan, receipts []OperationReceipt) ReceiptReport {
	normalized := normalizeOperationReceipts(receipts)
	report := ReceiptReport{
		SchemaVersion: ReceiptReportSchemaVersion,
		BaseSHA: plan.BaseSHA, HeadSHA: plan.HeadSHA,
		PlanDigest: plan.PlanDigest, Receipts: normalized,
	}
	if !receiptPlanKnown(plan) {
		report.Decision, report.Reason = ReceiptDecisionUnknown, ReceiptReasonInvalidPlan
		return finishReceiptReport(report)
	}
	if plan.Decision == DecisionFixedPoint {
		if len(normalized) != 0 {
			report.Decision, report.Reason = ReceiptDecisionUnknown, ReceiptReasonSetMismatch
		} else {
			report.Decision = ReceiptDecisionFixedPoint
			report.Reason = ReceiptReasonExactFixedPoint
		}
		return finishReceiptReport(report)
	}
	if plan.Decision != DecisionPlan {
		report.Decision = ReceiptDecisionUnknown
		report.Reason = ReceiptReasonPlanNotExecutable
		return finishReceiptReport(report)
	}
	actions, _ := selectedActionIndex(plan)
	receiptIndex, valid := operationReceiptIndex(normalized)
	if !valid {
		report.UnknownIndicatorIDs = []string{"receipt-set"}
		report.Decision, report.Reason = ReceiptDecisionUnknown, ReceiptReasonSetMismatch
		return finishReceiptReport(report)
	}
	for _, action := range sortedSelectedActions(plan.Selected) {
		receipt, exists := receiptIndex[action.IndicatorID]
		if !exists {
			report.MissingIndicatorIDs = append(
				report.MissingIndicatorIDs, actionObligationID(action.IndicatorID, "receipt"))
			continue
		}
		classifyOperationReceipt(plan, action, receipt, &report)
	}
	for actionID := range receiptIndex {
		if _, exists := actions[actionID]; !exists {
			report.UnknownIndicatorIDs = append(
				report.UnknownIndicatorIDs, actionObligationID(actionID, "receipt"))
		}
	}
	switch {
	case len(report.RejectedIndicatorIDs) != 0:
		report.Decision = ReceiptDecisionRejected
		report.Reason = ReceiptReasonRejectedIndicator
	case len(report.UnknownIndicatorIDs) != 0:
		report.Decision = ReceiptDecisionUnknown
		report.Reason = ReceiptReasonUnknownIndicator
	case len(report.MissingIndicatorIDs) != 0:
		report.Decision = ReceiptDecisionUnknown
		report.Reason = ReceiptReasonMissingIndicator
	default:
		report.Decision = ReceiptDecisionConformant
		report.Reason = ReceiptReasonVerified
	}
	return finishReceiptReport(report)
}
