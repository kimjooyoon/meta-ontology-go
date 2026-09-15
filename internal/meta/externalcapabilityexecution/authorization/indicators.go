package authorization

func makeIndicators(input Input) []Indicator {
	envelope := input.Envelope
	available := input.EnvelopeAvailable
	invocationKnown := available && input.Invocation.SubjectSHA != "" &&
		input.Invocation.RunID != "" && input.Invocation.RunAttempt > 0 &&
		envelope.RunID != "" && envelope.RunAttempt > 0
	defaultKnown := available && (envelope.DefaultDecision == "DENY" ||
		envelope.DefaultDecision == "ALLOW")
	states := []metricState{
		reportState(input),
		{available && envelope.Operation != "", envelope.Operation == ExpectedOperation, "OPERATION_UNKNOWN", envelope.EnvelopeDigest},
		{available && envelope.SubjectSHA != "" && input.ReportAvailable && input.Report.SubjectSHA != "" && input.Invocation.SubjectSHA != "", envelope.SubjectSHA == input.Invocation.SubjectSHA && input.Report.SubjectSHA == input.Invocation.SubjectSHA, "SUBJECT_UNKNOWN", envelope.EnvelopeDigest},
		{available && envelope.Issuer != "", envelope.Issuer == ExpectedIssuer, "ISSUER_UNKNOWN", envelope.EnvelopeDigest},
		{available && envelope.Scope != "", envelope.Scope == ExpectedScope, "SCOPE_UNKNOWN", envelope.EnvelopeDigest},
		policyState(input),
		{defaultKnown, envelope.DefaultDecision == ExpectedDefaultDecision, "DEFAULT_DECISION_UNKNOWN", envelope.EnvelopeDigest},
		{invocationKnown, envelope.RunID == input.Invocation.RunID && envelope.RunAttempt == input.Invocation.RunAttempt, "INVOCATION_UNKNOWN", envelope.EnvelopeDigest},
		{invocationKnown && validDigest(envelope.Nonce) && validDigest(envelope.EnvelopeDigest), envelopeSealExact(envelope), "NONCE_OR_ENVELOPE_DIGEST_UNKNOWN", envelope.EnvelopeDigest},
		{available && input.ReportAvailable, zeroEffect(envelope.EffectCeiling) && input.Report.RepositoryWrites == 0 && input.Report.ExternalRepositoryWrites == 0 && input.Report.OfficialMutationCount == 0 && input.Report.PromotionCount == 0, "EFFECT_EVIDENCE_UNKNOWN", envelope.EnvelopeDigest},
	}
	result := make([]Indicator, 0, len(metricSpecs))
	for index, spec := range metricSpecs {
		result = append(result, indicator(spec, states[index]))
	}
	return result
}

func indicator(spec metricSpec, state metricState) Indicator {
	metric := Indicator{MetricID: "gooo.metric.external-capability-authorization-" + spec.ID + ".v1",
		Class: spec.Class, ProofChoice: spec.Choice, Stage: spec.Stage,
		MetaOperation: spec.Operation, Target: 1, EvidenceDigest: state.evidence}
	switch {
	case !state.known:
		metric.Status, metric.UnknownReason = StatusUnknown, state.reason
	case state.satisfied:
		metric.Status, metric.Value = StatusSatisfied, 1
	default:
		metric.Status = StatusUnsatisfied
	}
	return metric
}

func zeroEffect(effect EffectCeiling) bool {
	return effect.RepositoryWrites == 0 && effect.ExternalRepositoryWrites == 0 &&
		effect.OfficialMutations == 0 && effect.Promotions == 0
}
