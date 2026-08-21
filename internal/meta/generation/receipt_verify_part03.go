package generation

func classifyOperationReceipt(
	plan Plan,
	action Action,
	receipt OperationReceipt,
	report *ReceiptReport,
) {
	if !receiptMatchesAction(plan, action, receipt) {
		report.UnknownIndicatorIDs = append(
			report.UnknownIndicatorIDs,
			actionObligationID(action.IndicatorID, "receipt"),
		)
		return
	}
	indicators, valid := indicatorReceiptIndex(receipt.Indicators)
	if !valid {
		report.UnknownIndicatorIDs = append(
			report.UnknownIndicatorIDs,
			actionObligationID(action.IndicatorID, "indicator-set"),
		)
		return
	}
	required := make(map[string]struct{}, len(action.RequiredIndicatorIDs))
	for _, identifier := range action.RequiredIndicatorIDs {
		required[identifier] = struct{}{}
		observation, exists := indicators[identifier]
		obligation := actionObligationID(action.IndicatorID, identifier)
		if !exists {
			report.MissingIndicatorIDs = append(report.MissingIndicatorIDs, obligation)
			continue
		}
		if observation.ProofChoice != action.ProofChoice ||
			!validDigest(observation.EvidenceDigest) {
			report.UnknownIndicatorIDs = append(report.UnknownIndicatorIDs, obligation)
			continue
		}
		switch observation.Verdict {
		case IndicatorVerdictPass:
		case IndicatorVerdictFail:
			report.RejectedIndicatorIDs = append(report.RejectedIndicatorIDs, obligation)
		case IndicatorVerdictUnknown:
			report.UnknownIndicatorIDs = append(report.UnknownIndicatorIDs, obligation)
		default:
			report.UnknownIndicatorIDs = append(report.UnknownIndicatorIDs, obligation)
		}
	}
	for identifier := range indicators {
		if _, exists := required[identifier]; !exists {
			report.UnknownIndicatorIDs = append(
				report.UnknownIndicatorIDs,
				actionObligationID(action.IndicatorID, identifier),
			)
		}
	}
}

func receiptMatchesAction(plan Plan, action Action, receipt OperationReceipt) bool {
	ledgerDigest, ledgerCount := planIndicatorDecisionLedgerProvenance(plan)
	return receipt.SchemaVersion == OperationReceiptSchemaVersion &&
		receipt.BaseSHA == plan.BaseSHA && receipt.HeadSHA == plan.HeadSHA &&
		receipt.PlanDigest == plan.PlanDigest &&
		receipt.IndicatorDecisionLedgerDigest == ledgerDigest &&
		receipt.IndicatorDecisionLedgerCount == ledgerCount &&
		receipt.ActionIndicatorID == action.IndicatorID &&
		receipt.Operation == action.Operation &&
		receipt.Evaluator == action.Evaluator &&
		receipt.ProofChoice == action.ProofChoice &&
		validDigest(receipt.ReceiptDigest) &&
		receipt.ReceiptDigest == operationReceiptDigest(receipt)
}
