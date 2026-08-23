package languagedeterministicquerybinding

func Evaluate(input Input) Artifact {
	boundCoordinates := coordinates(input)
	summary := summarize(input, boundCoordinates)
	artifact := Artifact{
		Schema: Schema, Decision: "FAIL_CLOSED", Resolution: "LOWER_RESOLUTION",
		ReasonCode: "QUERY_READINESS_COORDINATE_UNKNOWN", ExpectedHeadSHA: input.ExpectedHeadSHA,
		ConceptDigest: input.Concept.ArtifactDigest, ReadinessDigest: input.Readiness.Digest,
		QueryDigest: input.Query.ReportDigest, Coordinates: boundCoordinates,
		Summary: summary, RepositoryWrites: 0, MutationAuthorized: false,

		Indicators: indicators(summary)}
	artifact.Proofs = proofs(input, artifact)
	if summary.BoundCoordinates == FixedCoordinates && allProofsPassed(artifact.Proofs) {
		artifact.Decision = "PASS"
		artifact.Resolution = "EXACT"
		artifact.ReasonCode = "ALL_QUERY_READINESS_COORDINATES_BOUND"
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
		Coordinates: len(coordinates), BoundCoordinates: bound, Unresolved: len(coordinates) - bound,
		ReadinessCompleted: input.Readiness.Summary.Completed,
		ReadinessTotal:     input.Readiness.Summary.Total, ReadinessBPS: input.Readiness.Summary.ReadinessBPS,
		QuerySatisfied: input.Query.Summary.Satisfied, QueryTotal: input.Query.Summary.Total,
		Concepts: input.Concept.Report.Summary.Concepts, MetricBindings: input.Query.Summary.MetricBindings,
		EffectfulStages:     input.Query.Summary.EffectfulStages,
		RepositoryWrites:    input.Query.RepositoryWrites,
		MutationAuthorities: boolCount(input.Query.MutationAuthorized),
	}
}

func boolCount(value bool) int {
	if value {
		return 1
	}
	return 0
}
