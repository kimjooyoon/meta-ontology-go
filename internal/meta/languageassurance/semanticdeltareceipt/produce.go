package semanticdeltareceipt

func Produce(input Input) (Receipt, error) {
	beforeSource, beforeErr := projectSource(input.Before)
	afterSource, afterErr := projectSource(input.After)
	before := snapshot(input.Before, beforeSource, beforeErr)
	after := snapshot(input.After, afterSource, afterErr)
	text := textualDelta(input.Before, input.After)
	receipt := Receipt{Schema: ReceiptSchema, CaseID: input.CaseID, SubjectSHA: input.SubjectSHA,
		Producer: Producer, Consumer: Consumer, MetaOperation: MetaOperation, ProofChoice: "COHERENCE",
		Stage: "produce", Step: "separate-delta-layers", Before: before, After: after,
		TextualDelta: text, RepositoryWrites: 0}
	if beforeErr != nil || afterErr != nil {
		return unknownReceipt(receipt, beforeSource, afterSource, beforeErr, afterErr), nil
	}
	structural, err := structuralDelta(beforeSource, afterSource)
	if err != nil {
		return unknownReceipt(receipt, beforeSource, afterSource, err, nil), nil
	}
	claims := claimDelta(beforeSource, afterSource)
	receipt.StructuralDelta, receipt.SemanticClaimDelta = structural, claims
	if hasSemanticDelta(structural, claims) {
		receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = DecisionDelta, ResolutionExact, ClassChanged, ReasonMeaning
	} else {
		receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = DecisionFixedPoint, ResolutionExact, ClassPreserved, ReasonTextualOnly
	}
	receipt.Stage, receipt.Step = "produce", "classify"
	receipt.ClaimTransitions = transitions(beforeSource, afterSource, receipt.Classification, receipt.Reason)
	sealReceipt(&receipt)
	return receipt, nil
}

func unknownReceipt(receipt Receipt, before, after projectedSource, beforeErr, afterErr error) Receipt {
	receipt.StructuralDelta, receipt.SemanticClaimDelta = unknownStructuralDelta(), unknownClaimDelta()
	receipt.Decision, receipt.Resolution, receipt.Classification, receipt.Reason = DecisionFailClosed, ResolutionUnknown, ClassIndeterminate, ReasonUnavailable
	receipt.Stage, receipt.Step = "produce", "fail-closed"
	receipt.ClaimTransitions = transitions(before, after, ClassIndeterminate, ReasonUnavailable)
	if beforeErr == nil && afterErr != nil {
		receipt.After.ParseReason = "UNSUPPORTED_GOOO_SOURCE"
	}
	if beforeErr != nil && afterErr == nil {
		receipt.Before.ParseReason = "UNSUPPORTED_GOOO_SOURCE"
	}
	sealReceipt(&receipt)
	return receipt
}
