package languagediagnosticprovenance

func executeGuardrail(definition Definition) CaseResult {
	observation, found := guardrailObservation(definition.Fixture)
	if !found {
		return finishCase(definition,
			Evidence{ActualOutcome: "REJECT", FailureCode: "FIXTURE_UNKNOWN"}, nil, false)
	}
	_, failure := Normalize(observation)
	rejected := failure != nil
	evidence := Evidence{ActualOutcome: "TRACE", Rejected: rejected}
	if rejected {
		evidence.ActualOutcome = "REJECT"
		evidence.FailureCode = failure.Code
	}
	if !rejected {
		switch definition.GuardrailClass {
		case "UNKNOWN":
			evidence.UnknownAccepted = true
		case "MISSING_MAP":
			evidence.MissingMapAccepted = true
		case "AMBIGUOUS_MAP":
			evidence.AmbiguousAccepted = true
		case "INVALID":
			evidence.InvalidAccepted = true
		}
	}
	satisfied := rejected && failure.Code == definition.ExpectedReason
	return finishCase(definition, evidence, nil, satisfied)
}

func guardrailObservation(fixture string) (Observation, bool) {
	observation, found := sourceMapObservation("entity")
	if !found {
		return Observation{}, false
	}
	switch fixture {
	case "unknown-origin":
		observation.Origin = "UNKNOWN"
	case "unknown-stage":
		observation.Stage = "UNKNOWN"
	case "unknown-severity":
		observation.Severity = 255
	case "empty-code":
		observation.Code = ""
	case "missing-physical":
		observation.Physical = Span{}
	case "invalid-range":
		observation.Physical.End.Offset = observation.Physical.Start.Offset - 1
	case "missing-source-map":
		observation.SourceMap.Mappings = nil
	case "ambiguous-source-map":
		observation.SourceMap.Mappings = append(
			observation.SourceMap.Mappings,
			observation.SourceMap.Mappings[0],
		)
	default:
		return Observation{}, false
	}
	return observation, true
}
