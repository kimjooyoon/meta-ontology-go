package languagedeterministicquerybinding

func proofs(input Input, artifact Artifact) []Proof {
	foundation := input.Concept.Decision == "PASS" && input.Concept.Report.Summary.Concepts >= 16 &&
		artifact.Summary.BoundCoordinates == 12
	coherence := input.Query.Summary.Satisfied == 32 && input.Readiness.Summary.Completed >= 16 &&
		input.Readiness.Summary.ReadinessBPS >= 6666
	regression := artifact.Summary.Unresolved == 0 && artifact.Summary.EffectfulStages == 0 &&
		artifact.Summary.RepositoryWrites == 0 && artifact.Summary.MutationAuthorities == 0
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-versioned-query-and-readiness-contracts", EvidenceDigest: digestValue(struct{ Concept, Query string }{input.Concept.ArtifactDigest, input.Query.Source.RegistryDigest}), Passed: foundation},
		{Choice: "COHERENCE", MetaOperation: "bind-query-corpus-to-current-readiness", EvidenceDigest: digestValue(input.Readiness.Summary), Passed: coherence},
		{Choice: "REGRESSION", MetaOperation: "reject-unresolved-effects-writes-and-mutation", EvidenceDigest: digestValue(artifact.Summary), Passed: regression},
	}
}

func allProofsPassed(proofs []Proof) bool {
	if len(proofs) != 3 {
		return false
	}
	for _, proof := range proofs {
		if !proof.Passed {
			return false
		}
	}
	return true
}
