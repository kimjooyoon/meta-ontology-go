package authorization

func Evaluate(input Input) Receipt {
	receipt := Receipt{Schema: ReceiptSchema, SubjectSHA: input.Invocation.SubjectSHA,
		Total: MetricDenominator, RepositoryWrites: 0, OfficialMutationCount: 0,
		PromotionCount: 0, ExecutionAuthority: false,
		RepositoryMutationAuthority: false, PromotionAuthority: false,
		EnvelopeDigest:        input.Envelope.EnvelopeDigest,
		SourceReportDigest:    input.Envelope.SourceReportDigest,
		PolicySourceDigest:    input.Policy.SourceDigest,
		PolicyGeneratedDigest: input.Policy.GeneratedDigest,
		NonClaims:             append([]string(nil), nonClaims...),
		Indicators:            makeIndicators(input)}
	for _, indicator := range receipt.Indicators {
		if indicator.Status == StatusSatisfied {
			receipt.Completed++
		}
		if indicator.Status == StatusUnknown {
			receipt.UnknownIndicators++
		}
	}
	receipt.BasisPoints = receipt.Completed * 10000 / receipt.Total
	receipt.Claims, receipt.Unknowns, receipt.OpenClaims, receipt.DischargedClaims,
		receipt.RejectedClaims = makeClaims(receipt.Indicators)
	receipt.Proofs = makeProofs(receipt.Indicators)
	receipt.ReaderViews = makeReaderViews(receipt.Indicators)
	switch {
	case receipt.UnknownIndicators > 0:
		receipt.Decision, receipt.Resolution = DecisionFailClosed, ResolutionUnknown
		receipt.EnforcementEffect = EffectBlock
		receipt.Reason = "CAPABILITY_AUTHORIZATION_EVIDENCE_UNKNOWN"
	case receipt.Completed != receipt.Total:
		receipt.Decision, receipt.Resolution = DecisionDenied, ResolutionExact
		receipt.EnforcementEffect = EffectBlock
		receipt.Reason = "CAPABILITY_AUTHORIZATION_DENIED"
	default:
		receipt.Decision, receipt.Resolution = DecisionAuthorized, ResolutionExact
		receipt.EnforcementEffect = EffectNoEffect
		receipt.Reason = "CAPABILITY_AUTHORIZATION_SHADOW_EXACT"
	}
	receipt.ReceiptDigest = ""
	receipt.ReceiptDigest = digestValue(receipt)
	return receipt
}
