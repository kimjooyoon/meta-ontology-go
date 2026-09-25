package languageexampleexperiment

func Evaluate(input Input) Report {
	if reason, resolution := rejectInput(input); reason != "" {
		return closedReport(input, reason, resolution)
	}
	summary := summarize(input)
	values := indicators(summary, input.Contract.Fixed)
	satisfied := 0
	for _, value := range values {
		if value.Satisfied {
			satisfied++
		}
	}
	summary.Coordinates = Coordinates{Satisfied: satisfied, Total: len(values)}
	if len(values) > 0 {
		summary.Coordinates.BasisPoints = satisfied * 10000 / len(values)
	}
	report := Report{
		Schema: ReportSchema, Decision: "FAIL_CLOSED", Resolution: "EXACT",
		Reason: mismatchReason(values), Interpretation: "NO_LANGUAGE_QUALITY_CLAIM",
		SubjectSHA: input.Profile.SubjectSHA, ContractID: input.Contract.ID,
		Summary: summary, Indicators: values, Views: views(values), Proofs: proofs(values),
		NotClaimed: append([]string{}, input.Contract.NotClaimed...),
	}
	if satisfied == len(values) && len(values) == input.Contract.Fixed.Indicators {
		report.Decision, report.Reason = "PASS", "EXPERIMENT_CONTRACT_OBSERVED"
		report.Interpretation = "MINIMAL_VALUE_OBSERVED"
	}
	return finishReport(report)
}

func mismatchReason(values []Indicator) string {
	for _, value := range values {
		if !value.Satisfied {
			switch value.ID {
			case "value.artifact-digest-integrity":
				return "ARTIFACT_DIGEST_INVALID"
			case "value.golden-match":
				return "ARTIFACT_GOLDEN_MISMATCH"
			case "value.deterministic-replay":
				return "ARTIFACT_REPLAY_MISMATCH"
			case "resource.valid-samples":
				return "PROFILE_SAMPLE_INVALID"
			case "counterexample.unknown-emitter":
				return "UNKNOWN_EMITTER_NOT_REJECTED"
			}
		}
	}
	return "EXPERIMENT_CONTRACT_MISMATCH"
}
