package generation

func VerifyReceipts(plan Plan, receipts []OperationReceipt) ReceiptReport {
	return VerifyReceiptsWithFailures(plan, receipts, nil)
}

func VerifyReceiptsWithFailures(plan Plan, receipts []OperationReceipt, failures []ObservationFailure) ReceiptReport {
	normalized := normalizeOperationReceipts(receipts)
	report := ReceiptReport{
		SchemaVersion: ReceiptReportSchemaVersion,
		BaseSHA: plan.BaseSHA, HeadSHA: plan.HeadSHA,
		PlanDigest: plan.PlanDigest, Receipts: normalized,
		Failures: normalizeObservationFailures(failures),
	}
	if !receiptPlanKnown(plan) || validatePlanIndicatorDecisionLedger(plan) != nil {
		report.Decision, report.Reason = ReceiptDecisionUnknown, ReceiptReasonInvalidPlan
		return finishReceiptReport(report)
	}
	report.IndicatorDecisionLedgerDigest,
		report.IndicatorDecisionLedgerCount = planIndicatorDecisionLedgerProvenance(plan)
	if plan.Decision == DecisionFixedPoint {
		if len(normalized) != 0 || len(report.Failures) != 0 {
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
	if !validReceiptFailureList(report.Failures) || !receiptFailuresMatchPlan(plan, report.Failures) {
		report.Decision, report.Reason = ReceiptDecisionUnknown, ReceiptReasonUnknownIndicator
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
			for _, required := range action.RequiredIndicatorIDs {
				report.Unknowns = append(report.Unknowns, missingReceiptUnknown(action, required))
			}
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
	case hasRefutedFailure(report.Failures):
		report.Decision = ReceiptDecisionRefuted
		report.Reason = ReceiptReasonRefutedOperation
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

func receiptFailuresMatchPlan(plan Plan, failures []ObservationFailure) bool {
	actions, valid := selectedActionIndex(plan)
	if !valid {
		return false
	}
	seen := make(map[string]bool, len(failures))
	for _, failure := range failures {
		if seen[failure.ActionIndicatorID] {
			return false
		}
		if _, exists := actions[failure.ActionIndicatorID]; !exists {
			return false
		}
		seen[failure.ActionIndicatorID] = true
	}
	return true
}

func hasRefutedFailure(failures []ObservationFailure) bool {
	for _, failure := range failures {
		if failure.Decision == "REFUTED" {
			return true
		}
	}
	return false
}
