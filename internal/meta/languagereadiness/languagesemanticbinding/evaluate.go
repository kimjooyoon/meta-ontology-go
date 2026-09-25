package languagesemanticbinding

func Evaluate(input Input) (Report, error) {
	readiness, readinessBytes, err := loadJSON[readinessArtifact](input.ReadinessPath)
	if err != nil {
		return Report{}, err
	}
	concept, conceptBytes, err := loadJSON[conceptArtifact](input.ConceptPath)
	if err != nil {
		return Report{}, err
	}
	semantic, semanticBytes, err := loadJSON[semanticArtifact](input.SemanticPath)
	if err != nil {
		return Report{}, err
	}
	definition, err := validateConcept(concept)
	if err != nil {
		return Report{}, err
	}
	obligation, err := validateReadiness(readiness, input.ExpectedHeadSHA, concept.ArtifactDigest)
	if err != nil {
		return Report{}, err
	}
	if err := validateSemantic(semantic, input.ExpectedHeadSHA, definition.MetricBindings); err != nil {
		return Report{}, err
	}
	source := Source{
		ExpectedHeadSHA: input.ExpectedHeadSHA, ContractSchema: ContractSchema,
		RegistryDigest: ReadinessRegistryDigest, ReadinessFileDigest: fileDigest(readinessBytes),
		ReadinessArtifactDigest: readiness.ArtifactDigest, ConceptFileDigest: fileDigest(conceptBytes),
		ConceptArtifactDigest: concept.ArtifactDigest, ConceptEvidenceDigest: obligation.EvidenceDigest,
		SemanticFileDigest: fileDigest(semanticBytes), SemanticReportDigest: semantic.ReportDigest,
		SemanticRegistryDigest: semantic.Source.RegistryDigest, Producer: "languagesemanticbinding.Evaluate",
		Consumer: "self-improvement-cycle", MetaOperation: "bind-semantic-readiness-evidence",
	}
	summary := buildSummary(readiness, concept, semantic)
	report := Report{Schema: Schema, Decision: "PASS", ReasonCode: "SEMANTIC_READINESS_BOUND", Resolution: "EXACT"}
	report.Source, report.Summary = source, summary
	report.Indicators, report.Proofs = buildIndicators(summary), buildProofs(source)
	finalizeReport(&report)
	return report, nil
}

func buildSummary(readiness readinessArtifact, concept conceptArtifact, semantic semanticArtifact) Summary {
	return Summary{
		Coordinates: ExpectedCoordinates, BoundCoordinates: ExpectedCoordinates, Unresolved: 0,
		ReadinessCompleted: readiness.Snapshot.Summary.Completed, ReadinessTotal: readiness.Snapshot.Summary.Total,
		ReadinessBPS: readiness.Snapshot.Summary.ReadinessBPS, SemanticSatisfied: semantic.Summary.Satisfied,
		SemanticTotal: semantic.Summary.Total, Concepts: concept.Report.Summary.Concepts,
		MetricBindings: 19, Guardrails: 8, EffectfulStages: semantic.Summary.EffectfulStages,
		RepositoryWrites: 0, MutationAuthorities: 0,
	}
}
