package toolchainrelease

func guardrailIndicators(s Summary) []Indicator {
	values := []int{
		s.MissingReceipts, s.DuplicateReceipts, s.UnexpectedReceipts,
		s.CaseFailures, s.PlatformMismatches, s.ToolchainMismatches,
		s.HeadMismatches, s.DirtyBuilds, s.VCSRevisionMismatches,
		s.BinaryReplayMismatches, s.ArchiveReplayMismatches, s.SmokeFailures,
		s.ChecksumDrift, s.ReceiptDigestFailures, s.CorpusDrift,
		s.ConceptDrift, s.ProofFailures, s.Unresolved,
		s.RepositoryWrites, s.MutationAuthorities,
	}
	proofs := []string{
		"FOUNDATION", "FOUNDATION", "FOUNDATION", "COHERENCE",
		"COHERENCE", "FOUNDATION", "FOUNDATION", "REGRESSION",
		"FOUNDATION", "REGRESSION", "REGRESSION", "COHERENCE",
		"COHERENCE", "FOUNDATION", "FOUNDATION", "COHERENCE",
		"REGRESSION", "COHERENCE", "REGRESSION", "REGRESSION",
	}
	result := make([]Indicator, 0, GuardrailCount)
	for index, id := range guardrailMetricIDs {
		result = append(result, indicator(id, "GUARDRAIL", proofs[index],
			values[index], 0, "less_or_equal"))
	}
	return result
}
