package proposalcompat

func buildReceipt(source Source) Report {
	coordinates := []Coordinate{
		coordinate("exact-v2-source", "FOUNDATION", source.SourceSchema ==
			"gooo/autonomous-change-proposal-promotion/v2"),
		coordinate("source-promotion-contract", "FOUNDATION", source.SourceDecision ==
			DecisionPass && source.SourceSatisfied == 8 && source.SourceTotal == 8 &&
			source.SourceUnresolved == 0 && source.SourceRepositoryWrites == 0 &&
			!source.SourceMutationAuthorized),
		coordinate("exact-subject-link", "COHERENCE", source.ExpectedHeadSHA != ""),
		coordinate("exact-v1-target", "COHERENCE", source.TargetSchema == LegacySchema),
		coordinate("digest-bound-projection", "COHERENCE", validDigest(source.SourceReportDigest) &&
			validDigest(source.SourceFileSHA256) && validDigest(source.TargetReportDigest) &&
			validDigest(source.TargetFileSHA256)),
		coordinate("lossless-read-only-boundary", "REGRESSION", source.ProjectedFields ==
			projectedFields && source.FieldLosses == 0 && source.RepositoryWrites == 0),
	}
	summary := summarize(coordinates, source)
	report := Report{Schema: Schema, Decision: DecisionPass, Reason: ReasonReady,
		MetaOperation: "project-promotion-contract", Source: source, Summary: summary,
		Coordinates: coordinates, Indicators: indicators(source, summary),
		Proofs: proofs(coordinates), RepositoryWrites: source.RepositoryWrites}
	if summary.Satisfied != totalCoordinates || summary.Unresolved != 0 {
		report.Decision, report.Reason = DecisionFailClosed, ReasonRejected
	}
	return sealReport(report)
}

func coordinate(id, choice string, passed bool) Coordinate {
	status, reason := "NOT_SATISFIED", "COORDINATE_REJECTED"
	if passed {
		status, reason = "SATISFIED", "COORDINATE_EXACTLY_PROVEN"
	}
	return Coordinate{id, choice, status, reason}
}

func summarize(values []Coordinate, source Source) Summary {
	result := Summary{Total: len(values), ProjectedFields: source.ProjectedFields,
		FieldLosses: source.FieldLosses, RepositoryWrites: source.RepositoryWrites}
	for _, value := range values {
		if value.Status == "SATISFIED" {
			result.Satisfied++
		} else {
			result.NotSatisfied++
		}
	}
	result.ReadinessBPS = result.Satisfied * 10000 / result.Total
	return result
}
