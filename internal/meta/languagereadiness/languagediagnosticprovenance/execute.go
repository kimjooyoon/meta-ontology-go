package languagediagnosticprovenance

func executeCases(definitions []Definition) []CaseResult {
	results := make([]CaseResult, 0, len(definitions))
	for _, definition := range definitions {
		if definition.Kind == CaseGuardrail {
			results = append(results, executeGuardrail(definition))
			continue
		}
		results = append(results, executePositive(definition))
	}
	return results
}

func executePositive(definition Definition) CaseResult {
	observation, ordered, found := positiveObservation(definition)
	if !found {
		return finishCase(definition,
			Evidence{ActualOutcome: "REJECT", FailureCode: "FIXTURE_UNKNOWN"}, nil, false)
	}
	first, failure := Normalize(observation)
	if failure != nil {
		return finishCase(definition,
			Evidence{ActualOutcome: "REJECT", FailureCode: failure.Code}, nil, false)
	}
	replay, replayFailure := Normalize(observation)
	replayed := replayFailure == nil && first.TraceDigest == replay.TraceDigest
	semanticExpected := definition.Kind == CaseSourceMap
	satisfied := first.Stage == definition.ExpectedStage && replayed &&
		(first.Semantic != nil) == semanticExpected
	evidence := Evidence{
		ActualOutcome: "TRACE", Traced: true,
		PhysicalBound: true, LogicalBound: true,
		SemanticBound:   first.Semantic != nil,
		LSPProjected:    first.Diagnostic.Code != "",
		CanonicalReplay: replayed, OrderedDiagnostics: ordered,
		LineDirectiveRemap: !sameSpan(first.Physical, first.Logical),
		TypeClassified: definition.Kind == CaseType &&
			(first.Hardness == "HARD" || first.Hardness == "SOFT"),
		ProvenanceSteps: len(first.Steps),
	}
	return finishCase(definition, evidence, &first, satisfied)
}

func positiveObservation(definition Definition) (Observation, bool, bool) {
	switch definition.Kind {
	case CaseSyntax:
		observation, ordered := syntaxObservation(definition.Fixture)
		return observation, ordered, observation.Code != ""
	case CaseType:
		observation, ordered := typeObservation(definition.Fixture)
		return observation, ordered, observation.Code != ""
	case CaseSourceMap:
		observation, found := sourceMapObservation(definition.Fixture)
		return observation, false, found
	default:
		return Observation{}, false, false
	}
}

func finishCase(definition Definition, evidence Evidence, trace *Trace, satisfied bool) CaseResult {
	status := StatusNotSatisfied
	if satisfied {
		status = StatusSatisfied
	}
	result := CaseResult{Definition: definition, Evidence: evidence, Trace: trace, Status: status}
	result.Digest = digestJSON(result)
	return result
}
