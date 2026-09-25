package languageartifactoracle

func evaluateCases(input Input) []CaseResult {
	result := make([]CaseResult, len(input.Contract.Cases))
	for index, spec := range input.Contract.Cases {
		source, artifact, filename := selectCaseInput(input, spec.InputKind)
		observed := evaluateArtifact(source, artifact, filename, input.Entry)
		status := "NOT_SATISFIED"
		if observed.Decision == spec.ExpectedDecision && observed.Reason == spec.ExpectedReason {
			status = "SATISFIED"
		}
		result[index] = CaseResult{ID: spec.ID, Status: status,
			ExpectedDecision: spec.ExpectedDecision, ExpectedReason: spec.ExpectedReason,
			ObservedDecision: observed.Decision, ObservedResolution: observed.Resolution,
			ObservedReason: observed.Reason, ProofChoice: spec.ProofChoice,
			MetaOperation: spec.MetaOperation, Coordinate: observed.Coordinate, Checks: observed.Checks,
			SourceDigest: observed.SourceDigest, ArtifactDigest: observed.ArtifactDigest}
	}
	return result
}

func selectCaseInput(input Input, kind string) ([]byte, []byte, string) {
	switch kind {
	case "GENUINE":
		return input.Source, input.Genuine, input.Filename
	case "FORGED":
		return input.Source, input.Forged, input.Filename
	case "UNKNOWN_DECISION":
		return input.Source, input.UnknownDecision, input.Filename
	case "UNSUPPORTED_SOURCE":
		return input.UnsupportedSource, input.Genuine, input.UnsupportedFilename
	default:
		return nil, nil, ""
	}
}
