package toolchaincli

func proofs(report Report) []Proof {
	summary := report.Summary
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-versioned-cli-binary-and-corpus",
			EvidenceDigest: digestJSON(report.Source), Passed: report.Source.ObservationKnown &&
				summary.BinaryBindings == 1 && summary.RegistryDrift == 0},
		{Choice: "COHERENCE", MetaOperation: "replay-cli-command-contract",
			EvidenceDigest: digestJSON(report.Cases), Passed: summary.Satisfied == FixedTotal &&
				summary.Invocations == ExpectedRuns && summary.ReplayMatches == FixedTotal},
		{Choice: "REGRESSION", MetaOperation: "reject-cli-boundary-regressions",
			EvidenceDigest: digestJSON(summary), Passed: summary.GuardrailRejections == FixedGuardrails &&
				summary.ExitMismatches+summary.StdoutMismatches+summary.StderrMismatches+
					summary.ReplayMismatches+summary.RepositoryWrites == 0 && !report.MutationAuthorized},
	}
}

func allProofs(values []Proof) bool {
	if len(values) != 3 {
		return false
	}
	for _, value := range values {
		if !value.Passed {
			return false
		}
	}
	return true
}

func allIndicators(values []Indicator) bool {
	if len(values) != FixedIndicators {
		return false
	}
	for _, value := range values {
		if !value.Satisfied {
			return false
		}
	}
	return true
}
