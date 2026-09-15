package languageartifactoracle

func summarize(input Input, cases []CaseResult) Summary {
	summary := Summary{CasesTotal: len(cases), ProducerDependencies: input.Independence.ProducerDependencies}
	if legacyAccepted(input.LegacyAcceptance) {
		summary.LegacyValidatorCounterexamples = 1
	}
	for _, result := range cases {
		if result.Status == "SATISFIED" {
			summary.CasesSatisfied++
		}
		for _, check := range result.Checks {
			if check.Status == "UNKNOWN" {
				summary.UnknownChecks++
			}
		}
		switch result.ID {
		case "genuine-source-bound":
			if result.ObservedDecision == "PASS" {
				summary.ExactSourceBindings = 1
			}
		case "resealed-output-forgery":
			if result.ObservedReason == "ARTIFACT_SOURCE_PROJECTION_MISMATCH" {
				summary.ResealedForgeriesRejected = 1
			}
		case "unknown-artifact-decision":
			if result.ObservedReason == "ARTIFACT_DECISION_UNKNOWN" {
				summary.UnknownDecisionsRejected = 1
			}
		case "unsupported-source-projection":
			if result.ObservedResolution == "LOWER_RESOLUTION" {
				summary.LowerResolutions = 1
			}
		}
	}
	return summary
}
