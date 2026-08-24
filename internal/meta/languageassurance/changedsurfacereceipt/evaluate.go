package changedsurfacereceipt

func Evaluate(input Input) Report {
	observed := observe(input)
	summary := summarize(input, observed)
	decision, resolution, effect, reason := DecisionFixedPoint, ResolutionExact, EffectObserve, ReasonTotal
	if !validSHA(input.SubjectSHA) {
		decision, resolution, effect, reason = DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonUnavailable
	} else if input.Schema != InputSchema {
		decision, resolution, effect, reason = DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonSchema
	} else if summary.MalformedPaths > 0 {
		decision, resolution, effect, reason = DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonMalformed
	} else if summary.ChangedDuplicates+summary.ReceiptDuplicates > 0 {
		decision, resolution, effect, reason = DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonDuplicate
	} else if summary.UnknownReceipts > 0 {
		decision, resolution, effect, reason = DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonUnknown
	} else if summary.MissingReceipts > 0 {
		decision, resolution, effect, reason = DecisionFailClosed, ResolutionUnknown, EffectBlock, ReasonMissing
	} else if summary.OrphanReceipts > 0 {
		decision, resolution, effect, reason = DecisionFailClosed, ResolutionInvariant, EffectBlock, ReasonOrphan
	}
	if resolution == ResolutionUnknown {
		summary.UnknownPaths = 1
	}
	if decision != DecisionFixedPoint {
		summary.BlockedPaths = 1
	}
	report := Report{Schema: ReportSchema, SubjectSHA: input.SubjectSHA, Decision: decision,
		Resolution: resolution, EnforcementEffect: effect, Reason: reason,
		DenominatorID: DenominatorID, DenominatorDigest: digestValue(Denominator()),
		Summary: summary, Indicators: buildIndicators(summary, resolution),
		MetaOperations: MetaOperations(), RepositoryWrites: 0}
	seal(&report)
	return report
}

func summarize(input Input, observed observation) Summary {
	bound, missing, orphan := 0, 0, 0
	for surface := range observed.changedSet {
		if receipt, ok := observed.receiptSet[surface]; ok && receipt.Decision == "PASS" && receipt.Resolution == ResolutionExact {
			bound++
		} else {
			missing++
		}
	}
	for surface := range observed.receiptSet {
		if _, ok := observed.changedSet[surface]; !ok {
			orphan++
		}
	}
	return Summary{ChangedSurfaces: len(input.ChangedSurfaces), ReceiptsObserved: len(input.Receipts),
		BoundReceipts: bound, MissingReceipts: missing, OrphanReceipts: orphan,
		ChangedDuplicates: observed.changedDuplicates, ReceiptDuplicates: observed.receiptDuplicates,
		UnknownReceipts: observed.unknownReceipts, MalformedPaths: observed.malformedPaths,
		TotalityBPS: ratio(bound, len(input.ChangedSurfaces)), ChangedSetBPS: ratio(len(observed.changedSet), len(input.ChangedSurfaces)),
		UniqueBindingBPS: ratio(len(observed.receiptSet), len(input.Receipts))}
}

func ratio(numerator, denominator int) int {
	if denominator == 0 {
		return 10000
	}
	return numerator * 10000 / denominator
}
