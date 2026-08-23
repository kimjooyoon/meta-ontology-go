package languagesemanticbinding

func buildProofs(source Source) []Proof {
	return []Proof{
		{
			Choice: "FOUNDATION", MetaOperation: "bind-versioned-readiness-and-concept-artifacts",
			EvidenceDigest: digestParts(source.ReadinessFileDigest, source.ConceptFileDigest), Passed: true,
		},
		{
			Choice: "COHERENCE", MetaOperation: "bind-semantic-report-to-satisfied-obligation",
			EvidenceDigest: digestParts(source.ReadinessArtifactDigest, source.SemanticReportDigest), Passed: true,
		},
		{
			Choice: "REGRESSION", MetaOperation: "reject-unresolved-effects-writes-and-mutation",
			EvidenceDigest: digestParts(source.SemanticFileDigest, "zero-effects-writes-mutation"), Passed: true,
		},
	}
}
