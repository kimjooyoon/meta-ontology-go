package languagereadiness

func Evaluate(raw []byte) (Snapshot, error) {
	return evaluate(raw, evidenceDigests{})
}

func evaluate(raw []byte, evidence evidenceDigests) (Snapshot, error) {
	artifact, err := decodeArtifact(raw)
	if err != nil {
		return Snapshot{}, err
	}
	snapshot := Snapshot{
		Schema: SnapshotSchema, ContractSchema: ContractSchema,
		RegistryDigest: registryDigest(), SourceArtifactDigest: artifact.ArtifactDigest,
		RepositoryWrites: artifact.RepositoryWrites,
	}
	if reason := artifactResolutionReason(artifact); reason != "" {
		for _, obligation := range obligations {
			snapshot.Obligations = append(snapshot.Obligations, ObligationResult{
				Obligation: obligation, Status: "UNRESOLVED", Reason: reason,
			})
		}
		return summarize(snapshot), nil
	}
	byID := make(map[string][]conceptEvidence, len(artifact.Report.Concepts))
	for _, concept := range artifact.Report.Concepts {
		byID[concept.ID] = append(byID[concept.ID], concept)
	}
	for _, obligation := range obligations {
		snapshot.Obligations = append(snapshot.Obligations,
			evaluateObligation(obligation, byID[obligation.ConceptID], evidence))
	}
	return summarize(snapshot), nil
}

func evaluateObligation(obligation Obligation, concepts []conceptEvidence, evidence evidenceDigests) ObligationResult {
	result := ObligationResult{Obligation: obligation}
	if len(concepts) == 0 {
		return ObligationResult{Obligation: obligation, Status: "NOT_SATISFIED", Reason: "CONCEPT_NOT_REGISTERED"}
	}
	if len(concepts) != 1 {
		result.Status, result.Reason = "UNRESOLVED", "CONCEPT_EVIDENCE_NOT_UNIQUE"
		return result
	}
	concept := concepts[0]
	result.EvidenceDigest = digestJSON(concept)
	if !completeConcept(concept) {
		result.Status, result.Reason = "NOT_SATISFIED", "CONCEPT_CONFORMANCE_INCOMPLETE"
		return result
	}
	if evidenceDigest, reason, required := requiredEvidence(obligation.ConceptID, evidence); required {
		if evidenceDigest == "" {
			result.Status, result.Reason = "NOT_SATISFIED", reason
			return result
		}
		result.EvidenceDigest = digestJSON(externalEvidence{concept, evidenceDigest})
	}
	result.Status, result.Reason = "SATISFIED", "CONCEPT_CONFORMANCE_EXPLICIT"
	return result
}

func completeConcept(concept conceptEvidence) bool {
	if concept.Stage != "OPERATING" && concept.Stage != "CONFORMED" {
		return false
	}
	if len(concept.CodeBindings) == 0 || len(concept.MetricBindings) == 0 || len(concept.UseCases) == 0 {
		return false
	}
	for _, useCase := range concept.UseCases {
		if useCase.ID == "" || useCase.Trigger == "" || useCase.ExpectedOutcome == "" {
			return false
		}
	}
	return true
}
