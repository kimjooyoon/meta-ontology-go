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
		report.Unknowns = append(report.Unknowns,
			malformedReceiptUnknown(action, "receipt", ReceiptUnknownClassMalformedEvidence))
		return
	}
	indicators, valid := indicatorReceiptIndex(receipt.Indicators)
	if !valid {
		report.UnknownIndicatorIDs = append(
			report.UnknownIndicatorIDs,
			actionObligationID(action.IndicatorID, "indicator-set"),
		)
		report.Unknowns = append(report.Unknowns,
			malformedReceiptUnknown(action, "indicator-set", ReceiptUnknownClassMalformedEvidence))
		return
	}
	required := make(map[string]struct{}, len(action.RequiredIndicatorIDs))
	for _, identifier := range action.RequiredIndicatorIDs {
		required[identifier] = struct{}{}
		observation, exists := indicators[identifier]
		obligation := actionObligationID(action.IndicatorID, identifier)
		if !exists {
			report.MissingIndicatorIDs = append(report.MissingIndicatorIDs, obligation)
			report.Unknowns = append(report.Unknowns, missingReceiptUnknown(action, identifier))
			continue
		}
		if observation.ProofChoice != action.ProofChoice ||
			!validDigest(observation.EvidenceDigest) {
			report.UnknownIndicatorIDs = append(report.UnknownIndicatorIDs, obligation)
			report.Unknowns = append(report.Unknowns,
				malformedReceiptUnknown(action, identifier, ReceiptUnknownClassMalformedEvidence))
			continue
		}
		switch observation.Verdict {
		case IndicatorVerdictPass:
		case IndicatorVerdictFail:
			report.RejectedIndicatorIDs = append(report.RejectedIndicatorIDs, obligation)
		case IndicatorVerdictUnknown:
			report.UnknownIndicatorIDs = append(report.UnknownIndicatorIDs, obligation)
			report.Unknowns = append(report.Unknowns,
				malformedReceiptUnknown(action, identifier, ReceiptUnknownClassUnexpectedEvidence))
		default:
			report.UnknownIndicatorIDs = append(report.UnknownIndicatorIDs, obligation)
			report.Unknowns = append(report.Unknowns,
				malformedReceiptUnknown(action, identifier, ReceiptUnknownClassUnexpectedEvidence))
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
	return action.SubjectKind == action.InputSubjectKind &&
		validDigest(action.InputContractSourceDigest) && validDigest(action.InputContractSemanticDigest) &&
		receipt.SchemaVersion == OperationReceiptSchemaVersion &&
		receipt.BaseSHA == plan.BaseSHA && receipt.HeadSHA == plan.HeadSHA &&
		receipt.PlanDigest == plan.PlanDigest &&
		receipt.IndicatorDecisionLedgerDigest == ledgerDigest &&
		receipt.IndicatorDecisionLedgerCount == ledgerCount &&
		receipt.ActionIndicatorID == action.IndicatorID &&
		receipt.OperationInputDigest == action.SourceIndicator.OperationInputDigest &&
		receipt.SubjectKind == action.SubjectKind &&
		receipt.InputSubjectKind == action.InputSubjectKind &&
		receipt.InputContractSourceDigest == action.InputContractSourceDigest &&
		receipt.InputContractSemanticDigest == action.InputContractSemanticDigest &&
		receipt.Operation == action.Operation &&
		receipt.Activity == action.Activity &&
		receipt.Output == action.Output &&
		receipt.Executor == action.Executor &&
		receipt.Evaluator == action.Evaluator &&
		receipt.ProofChoice == action.ProofChoice &&
		validDigest(receipt.ReceiptDigest) &&
		receipt.ReceiptDigest == operationReceiptDigest(receipt)
}
