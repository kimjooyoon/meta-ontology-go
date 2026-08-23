package languagediagnosticprovenancebinding

func proofs(input Input, artifact Artifact) []Proof {
	foundation := input.Concept.Decision == "PASS" &&
		input.Concept.Report.Summary.Concepts >= 18 &&
		artifact.Summary.BoundCoordinates == 12
	coherence := input.Provenance.Summary.Satisfied == 18 &&
		input.Readiness.Summary.Completed >= 18 &&
		input.Readiness.Summary.ReadinessBPS >= 7500
	regression := artifact.Summary.Unresolved == 0 &&
		artifact.Summary.EffectfulStages == 0 &&
		artifact.Summary.RepositoryWrites == 0 &&
		artifact.Summary.MutationAuthorities == 0
	return []Proof{
		{
			Choice:        "FOUNDATION",
			MetaOperation: "bind-versioned-diagnostic-and-readiness-contracts",
			EvidenceDigest: digestValue(struct{ Concept, Registry string }{
				input.Concept.ArtifactDigest, input.Provenance.Source.RegistryDigest}),
			Passed: foundation,
		},
		{
			Choice:         "COHERENCE",
			MetaOperation:  "bind-diagnostic-corpus-to-readiness-floor",
			EvidenceDigest: digestValue(input.Readiness.Summary),
			Passed:         coherence,
		},
		{
			Choice:         "REGRESSION",
			MetaOperation:  "reject-unresolved-effects-writes-and-mutation",
			EvidenceDigest: digestValue(artifact.Summary),
			Passed:         regression,
		},
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
