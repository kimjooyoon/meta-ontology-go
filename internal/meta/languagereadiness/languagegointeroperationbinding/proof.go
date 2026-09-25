package languagegointeroperationbinding

func proofs(input Input, artifact Artifact) []Proof {
	foundation := input.Concept.Decision == "PASS" && input.Concept.Report.Summary.Concepts >= 17 &&
		artifact.Summary.BoundCoordinates == 12
	coherence := input.Interoperation.Summary.Satisfied == 24 && input.Readiness.Summary.Completed >= 17 &&
		input.Readiness.Summary.ReadinessBPS >= 7083
	regression := artifact.Summary.Unresolved == 0 && artifact.Summary.EffectfulStages == 0 &&
		artifact.Summary.RepositoryWrites == 0 && artifact.Summary.MutationAuthorities == 0
	return []Proof{
		{Choice: "FOUNDATION", MetaOperation: "bind-versioned-go-interoperation-and-readiness-contracts",
			EvidenceDigest: digestValue(struct{ Concept, Registry string }{input.Concept.ArtifactDigest, input.Interoperation.Source.RegistryDigest}), Passed: foundation},
		{Choice: "COHERENCE", MetaOperation: "bind-go-interoperation-corpus-to-current-readiness",
			EvidenceDigest: digestValue(input.Readiness.Summary), Passed: coherence},
		{Choice: "REGRESSION", MetaOperation: "reject-unresolved-effects-writes-and-mutation",
			EvidenceDigest: digestValue(artifact.Summary), Passed: regression},
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
