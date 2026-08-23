package languagesemanticbinding

func validateReadiness(value readinessArtifact, head, conceptDigest string) (readinessObligation, error) {
	if err := require(value.Schema == ReadinessSchema && value.HeadSHA == head, "readiness identity mismatch"); err != nil {
		return readinessObligation{}, err
	}
	if err := require(validDigest(value.ArtifactDigest), "readiness digest unknown"); err != nil {
		return readinessObligation{}, err
	}
	snapshot := value.Snapshot
	if err := require(snapshot.Schema == SnapshotSchema && snapshot.ContractSchema == ContractSchema, "contract mismatch"); err != nil {
		return readinessObligation{}, err
	}
	if err := require(snapshot.Decision == "PASS" && snapshot.RegistryDigest == ReadinessRegistryDigest, "registry not fixed"); err != nil {
		return readinessObligation{}, err
	}
	if err := require(snapshot.SourceArtifactDigest == conceptDigest && snapshot.RepositoryWrites == 0, "concept chain mismatch"); err != nil {
		return readinessObligation{}, err
	}
	summary := snapshot.Summary
	if err := require(summary.Completed == 15 && summary.Total == 24 && summary.NotSatisfied == 9, "readiness count mismatch"); err != nil {
		return readinessObligation{}, err
	}
	if err := require(summary.Unresolved == 0 && summary.ReadinessBPS == 6250, "readiness resolution mismatch"); err != nil {
		return readinessObligation{}, err
	}
	if err := require(summary.RatioNumerator == 15 && summary.RatioDenominator == 24, "ratio mismatch"); err != nil {
		return readinessObligation{}, err
	}
	transition := value.TransitionInput
	if err := require(transition.ContractSchema == SnapshotSchema && transition.RegistryDigest == snapshot.RegistryDigest, "transition contract mismatch"); err != nil {
		return readinessObligation{}, err
	}
	if err := require(transition.Completed == 15 && transition.Total == 24 && transition.BasisPoints == 6250, "transition count mismatch"); err != nil {
		return readinessObligation{}, err
	}
	return semanticObligation(snapshot.Obligations, transition.Evidence)
}

func semanticObligation(obligations []readinessObligation, evidence []transitionEvidence) (readinessObligation, error) {
	for _, obligation := range obligations {
		if obligation.ID != ObligationID {
			continue
		}
		valid := obligation.Area == "LANGUAGE" && obligation.ProofChoice == "COHERENCE"
		valid = valid && obligation.ConceptID == ConceptID && obligation.Status == "SATISFIED"
		valid = valid && obligation.Reason == "CONCEPT_CONFORMANCE_EXPLICIT" && validDigest(obligation.EvidenceDigest)
		if err := require(valid && transitionSatisfied(evidence), "semantic obligation not explicit"); err != nil {
			return readinessObligation{}, err
		}
		return obligation, nil
	}
	return readinessObligation{}, require(false, "semantic obligation missing")
}

func transitionSatisfied(evidence []transitionEvidence) bool {
	for _, item := range evidence {
		if item.ID == ObligationID {
			return item.Status == "SATISFIED"
		}
	}
	return false
}
