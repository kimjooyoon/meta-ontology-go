package languagediagnosticprovenance

func Registry() CaseRegistry {
	cases := positiveDefinitions(CaseSyntax,
		[]string{"physical-error", "line-directive", "multiple-errors"}, "PARSE")
	cases = append(cases, positiveDefinitions(CaseType,
		[]string{"undefined-identifier", "assignment-mismatch", "constraint-violation"}, "TYPE")...)
	cases = append(cases, positiveDefinitions(CaseSourceMap,
		[]string{"entity", "field", "activity", "slot"}, "FORMAT")...)
	cases = append(cases, guardrailDefinitions()...)
	return CaseRegistry{Schema: RegistrySchema, Version: RegistryVersion, Cases: cases}
}

func positiveDefinitions(kind CaseKind, fixtures []string, stage string) []Definition {
	result := make([]Definition, 0, len(fixtures))
	for _, fixture := range fixtures {
		result = append(result, Definition{
			ID: string(kind) + ":" + fixture, Kind: kind, Fixture: fixture,
			ExpectedOutcome: "TRACE", ExpectedStage: stage,
			ProofChoice: "COHERENCE",
			MetaOperation: "normalize-diagnostic-provenance",
		})
	}
	return result
}

func guardrailDefinitions() []Definition {
	return []Definition{
		guard("unknown-origin", "PROVENANCE_ORIGIN_UNKNOWN", "UNKNOWN"),
		guard("unknown-stage", "PROVENANCE_STAGE_UNKNOWN", "UNKNOWN"),
		guard("unknown-severity", "PROVENANCE_SEVERITY_UNKNOWN", "UNKNOWN"),
		guard("empty-code", "PROVENANCE_CODE_UNKNOWN", "UNKNOWN"),
		guard("missing-physical", "PHYSICAL_POSITION_UNKNOWN", "INVALID"),
		guard("invalid-range", "PHYSICAL_RANGE_INVALID", "INVALID"),
		guard("missing-source-map", "SOURCE_MAP_MISSING", "MISSING_MAP"),
		guard("ambiguous-source-map", "SOURCE_MAP_AMBIGUOUS", "AMBIGUOUS_MAP"),
	}
}

func guard(fixture, reason, class string) Definition {
	return Definition{
		ID: "GUARDRAIL:" + fixture, Kind: CaseGuardrail, Fixture: fixture,
		ExpectedOutcome: "REJECT", ExpectedStage: "NORMALIZE",
		ExpectedReason: reason, ProofChoice: "REGRESSION",
		MetaOperation: "fail-closed-diagnostic-provenance",
		GuardrailClass: class,
	}
}
