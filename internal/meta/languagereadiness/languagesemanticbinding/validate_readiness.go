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
	counts := summary.Completed >= SemanticReadinessFloor && summary.Total == 24
	counts = counts && summary.NotSatisfied == summary.Total-summary.Completed
	if err := require(counts, "readiness count mismatch"); err != nil {
		return readinessObligation{}, err
	}
	expectedBPS := summary.Completed * 10000 / summary.Total
	resolved := summary.Unresolved == 0 && summary.ReadinessBPS == expectedBPS
	resolved = resolved && summary.ReadinessBPS >= SemanticReadinessBPS
	if err := require(resolved, "readiness resolution mismatch"); err != nil {
		return readinessObligation{}, err
	}
	if err := require(summary.RatioNumerator == summary.Completed && summary.RatioDenominator == summary.Total, "ratio mismatch"); err != nil {
		return readinessObligation{}, err
	}
	transition := value.TransitionInput
	if err := require(transition.ContractSchema == SnapshotSchema && transition.RegistryDigest == snapshot.RegistryDigest, "transition contract mismatch"); err != nil {
		return readinessObligation{}, err
	}
	transitionCurrent := transition.Completed == summary.Completed && transition.Total == summary.Total
	transitionCurrent = transitionCurrent && transition.BasisPoints == summary.ReadinessBPS
	if err := require(transitionCurrent, "transition count mismatch"); err != nil {
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
