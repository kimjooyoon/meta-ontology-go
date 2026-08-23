package languagediagnosticprovenancebinding

func Evaluate(input Input) Artifact {
	boundCoordinates := coordinates(input)
	summary := summarize(input, boundCoordinates)
	artifact := Artifact{
		Schema: Schema, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION",
		ReasonCode:       "DIAGNOSTIC_PROVENANCE_READINESS_COORDINATE_UNKNOWN",
		ExpectedHeadSHA:  input.ExpectedHeadSHA,
		ConceptDigest:    input.Concept.ArtifactDigest,
		ReadinessDigest:  input.Readiness.Digest,
		ProvenanceDigest: input.Provenance.ReportDigest,
		Coordinates:      boundCoordinates, Summary: summary,
		RepositoryWrites: 0, MutationAuthorized: false,
		Indicators: indicators(summary),
	}
	artifact.Proofs = proofs(input, artifact)
	if summary.BoundCoordinates == FixedCoordinates &&
		allProofsPassed(artifact.Proofs) {
		artifact.Decision, artifact.Resolution = "PASS", "EXACT"
		artifact.ReasonCode = "ALL_DIAGNOSTIC_PROVENANCE_READINESS_COORDINATES_BOUND"
	}
	artifact.ArtifactDigest = artifactDigest(artifact)
	return artifact
}

func summarize(input Input, coordinates []Coordinate) Summary {
	bound := 0
	for _, coordinate := range coordinates {
		if coordinate.Bound {
			bound++
		}
	}
	return Summary{
		Coordinates: len(coordinates), BoundCoordinates: bound,
		Unresolved:          len(coordinates) - bound,
		ReadinessCompleted:  input.Readiness.Summary.Completed,
		ReadinessTotal:      input.Readiness.Summary.Total,
		ReadinessBPS:        input.Readiness.Summary.ReadinessBPS,
		ProvenanceSatisfied: input.Provenance.Summary.Satisfied,
		ProvenanceTotal:     input.Provenance.Summary.Total,
		Concepts:            input.Concept.Report.Summary.Concepts,
		MetricBindings:      input.Provenance.Summary.MetricBindings,
		EffectfulStages:     input.Provenance.Summary.EffectfulStages,
		RepositoryWrites:    input.Provenance.RepositoryWrites,
		MutationAuthorities: boolCount(input.Provenance.MutationAuthorized),
	}
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}
