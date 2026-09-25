package toolchainformatfix

func proofs(report Report, cycle cycleEvidence) []Proof {
	foundation := report.Source.ObservationKnown && report.Summary.BinaryBindings == 1 &&
		report.Summary.RegistryDrift == 0
	coherence := cycle.passed && report.Summary.PositivePaths == FixedPositive &&
		report.Summary.InMemoryApplications == 1 && report.Summary.FixedPoints == 2 &&
		report.Summary.DirectWrites == 0
	regression := report.Summary.GuardrailRejections == FixedGuardrails &&
		report.Summary.ExitMismatches == 0 && report.Summary.OutputMismatches == 0 &&
		report.Summary.ReplayMismatches == 0 && report.Summary.RepositoryWrites == 0
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-format-fix-authorities",
			EvidenceDigest: digestJSON(report.Source), Passed: foundation},
		{Choice: "COHERENCE", MetaOperation: "apply-plan-to-fixed-point-in-memory",
			EvidenceDigest: digestJSON(cycle), Passed: coherence},
		{Choice: "REGRESSION", MetaOperation: "reject-unknown-and-write-boundaries",
			EvidenceDigest: digestJSON(report.Cases), Passed: regression},
	}
}
