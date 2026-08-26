package symbolicinvocationusecase

const (
	reasonSatisfied       = "SYMBOLIC_INVOCATION_USECASE_OBSERVED"
	reasonDecisionUnknown = "SYMBOLIC_INVOCATION_USECASE_DECISION_UNKNOWN"
	reasonSubjectMismatch = "SYMBOLIC_INVOCATION_USECASE_SUBJECT_MISMATCH"
	reasonEvidenceInvalid = "SYMBOLIC_INVOCATION_USECASE_EVIDENCE_INVALID"
	reasonLinkMismatch    = "SYMBOLIC_INVOCATION_USECASE_LINK_MISMATCH"
	reasonEffectsObserved = "SYMBOLIC_INVOCATION_USECASE_EFFECTS_OBSERVED"
)

func Evaluate(input Input) (Report, error) {
	if err := input.Contract.Validate(); err != nil {
		return Report{}, err
	}
	value, reason, resolution := collectFacts(input)
	indicators := buildIndicators(input.Contract, value)
	if reason != "" {
		for index := range indicators {
			indicators[index].Satisfied = false
		}
	}
	coordinates := countIndicators(indicators)
	report := Report{
		Schema: "gooo/symbolic-invocation-usecase-report/v1", SubjectSHA: input.SubjectSHA,
		MetricID: input.Contract.MetricID, Decision: "PASS", Resolution: "EXACT", Reason: reasonSatisfied,
		Summary: Summary{
			Coordinates: coordinates, UserDecisions: value.UserDecisions,
			AcceptedInstances: value.AcceptedInstances, RejectedInstances: value.RejectedInstances,
			GeneratedInstances: value.GeneratedInstances, GeneratedGoldenMatches: value.GeneratedGoldenMatches,
			DeterministicReplays: value.DeterministicReplays, Unknowns: value.Unknowns,
			Source: value.Source, Producer: value.Producer, Resources: value.Resources, Effects: value.Effects,
		},
		Indicators: indicators, Views: buildViews(indicators), PromotionCreditBPS: 0,
		RepositoryWrites: value.Effects.RepositoryWrites, MutationAuthority: value.Effects.MutationAuthority,
		NotClaimed: CanonicalNonClaims(),
	}
	if reason != "" {
		report.Decision, report.Resolution, report.Reason = "FAIL_CLOSED", resolution, reason
	} else if coordinates.Satisfied != coordinates.Total {
		report.Decision, report.Resolution, report.Reason = "FAIL_CLOSED", "INVARIANT_ONLY", reasonEvidenceInvalid
	}
	return sealReport(report), nil
}

func collectFacts(input Input) (facts, string, string) {
	if value, reason, resolution, failed := resolutionFailure(input); failed {
		return value, reason, resolution
	}
	if reason := identityFailure(input); reason != "" {
		return facts{}, reason, "INVARIANT_ONLY"
	}
	if reason := contractFailure(input); reason != "" {
		return facts{}, reason, "INVARIANT_ONLY"
	}
	value := invocationFacts(input)
	mutationAuthorities := 0
	if value.Effects.MutationAuthority {
		mutationAuthorities = 1
	}
	if value.Effects.RepositoryWrites != input.Contract.ExpectedRepositoryWrites ||
		mutationAuthorities != input.Contract.ExpectedMutationAuthorities {
		return value, reasonEffectsObserved, "INVARIANT_ONLY"
	}
	return value, "", "EXACT"
}
