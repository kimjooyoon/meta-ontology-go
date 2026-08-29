package promotioncontinuity

func buildIndicators(report Report, guardOK, recoveryOK, effectsOK, authorityOK, mixed bool) []Indicator {
	resolution := report.Resolution
	producer, consumer, operation := report.Producer, report.Consumer, report.MetaOperation
	effectsApplicability, effectsSatisfied := "APPLICABLE", effectsOK
	if mixed {
		effectsApplicability, effectsSatisfied = "NOT_APPLICABLE", true
	}
	values := []struct {
		id, class, choice string
		applicability     string
		value, target     int
		satisfied         bool
	}{
		{"gooo.metric.language.promotion-continuity-readiness-bps.v1", "OUTCOME", "COHERENCE", "APPLICABLE", report.Summary.ReadinessBPS, 10000, report.Decision == "PASS"},
		{"gooo.metric.language.promotion-continuity-authorized-guards.v1", "DRIVER", "FOUNDATION", "APPLICABLE", report.Summary.AuthorizedGuardReceipts, 1, guardOK},
		{"gooo.metric.language.promotion-continuity-authorized-routes.v1", "DRIVER", "COHERENCE", "APPLICABLE", report.Summary.AuthorizedRecoveryRoutes, 1, recoveryOK},
		{"gooo.metric.language.promotion-continuity-unresolved.guardrail.v1", "GUARDRAIL", "FOUNDATION", "APPLICABLE", report.Summary.Unresolved, 0, report.Summary.Unresolved == 0},
		{"gooo.metric.language.promotion-continuity-effects.guardrail.v1", "GUARDRAIL", "REGRESSION", effectsApplicability, report.Source.Recovery.TransformationEffects, 0, effectsSatisfied},
		{"gooo.metric.language.promotion-continuity-writes.guardrail.v1", "GUARDRAIL", "REGRESSION", "APPLICABLE", report.RepositoryWrites, 0, authorityOK},
		{"gooo.metric.language.promotion-continuity-authority.guardrail.v1", "GUARDRAIL", "REGRESSION", "APPLICABLE", boolInt(report.RepositoryMutationAuthorized), 0, authorityOK},
		{"gooo.metric.language.promotion-continuity-source-mutations.guardrail.v1", "GUARDRAIL", "REGRESSION", "APPLICABLE", boolInt(!report.Source.Recovery.SourceWorkspaceUnchanged), 0, effectsOK},
		{"gooo.metric.language.promotion-continuity-terminal-preserved.v1", "OUTCOME", "REGRESSION", "APPLICABLE", boolInt(mixed), 1, mixed},
	}
	indicators := make([]Indicator, 0, len(values))
	for _, value := range values {
		indicators = append(indicators, Indicator{
			MetricID: value.id, Class: value.class, ProofChoice: value.choice,
			Producer: producer, Consumer: consumer, MetaOperation: operation,
			Applicability: value.applicability, Resolution: resolution,
			Value: value.value, Target: value.target,
			Satisfied: value.satisfied,
		})
	}
	return indicators
}

func buildProofs(report Report, foundation, coherence, regression, mixed bool) []Proof {
	foundationOperation := "bind-authorized-cycle-receipts"
	coherenceOperation := "cohere-successor-authorization"
	regressionOperation := "reject-effects-writes-or-authority"
	if mixed {
		foundationOperation = "bind-non-promoting-cycle-receipts"
		coherenceOperation = "cohere-non-promoting-terminal"
		regressionOperation = "preserve-non-promoting-boundary"
	}
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: foundationOperation,
			EvidenceDigest: report.Source.Guard.FileSHA256, Passed: foundation},
		{Choice: "COHERENCE", MetaOperation: coherenceOperation,
			EvidenceDigest: report.Source.Recovery.ReportDigest, Passed: coherence},
		{Choice: "REGRESSION", MetaOperation: regressionOperation,
			EvidenceDigest: report.Source.Recovery.FileSHA256, Passed: regression},
	}
}
