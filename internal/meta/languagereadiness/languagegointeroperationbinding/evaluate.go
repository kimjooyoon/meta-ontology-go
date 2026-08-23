package languagegointeroperationbinding

func Evaluate(input Input) Artifact {
	boundCoordinates := coordinates(input)
	summary := summarize(input, boundCoordinates)
	artifact := Artifact{Schema: Schema, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION",
		ReasonCode: "GO_INTEROPERATION_READINESS_COORDINATE_UNKNOWN", ExpectedHeadSHA: input.ExpectedHeadSHA,
		ConceptDigest: input.Concept.ArtifactDigest, ReadinessDigest: input.Readiness.Digest,
		InteropDigest: input.Interoperation.ReportDigest, Coordinates: boundCoordinates,
		Summary: summary, RepositoryWrites: 0, MutationAuthorized: false, Indicators: indicators(summary)}
	artifact.Proofs = proofs(input, artifact)
	if summary.BoundCoordinates == FixedCoordinates && allProofsPassed(artifact.Proofs) {
		artifact.Decision, artifact.Resolution = "PASS", "EXACT"
		artifact.ReasonCode = "ALL_GO_INTEROPERATION_READINESS_COORDINATES_BOUND"
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
	return Summary{Coordinates: len(coordinates), BoundCoordinates: bound, Unresolved: len(coordinates) - bound,
		ReadinessCompleted: input.Readiness.Summary.Completed, ReadinessTotal: input.Readiness.Summary.Total,
		ReadinessBPS: input.Readiness.Summary.ReadinessBPS, InteropSatisfied: input.Interoperation.Summary.Satisfied,
		InteropTotal: input.Interoperation.Summary.Total, Concepts: input.Concept.Report.Summary.Concepts,
		MetricBindings:      input.Interoperation.Summary.MetricBindings,
		EffectfulStages:     input.Interoperation.Summary.EffectfulStages,
		RepositoryWrites:    input.Interoperation.RepositoryWrites,
		MutationAuthorities: boolCount(input.Interoperation.MutationAuthorized)}
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}
