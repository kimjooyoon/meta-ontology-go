package languagegointeroperation

func Registry() CaseRegistry {
	cases := generatorDefinitions()
	cases = append(cases, go127Definitions()...)
	cases = append(cases, guardrailDefinitions()...)
	return CaseRegistry{Schema: RegistrySchema, Version: RegistryVersion, Cases: cases}
}

func generatorDefinitions() []Definition {
	ids := []string{
		"single-entity", "two-entities", "builtin-output", "entity-flow",
		"two-inputs", "two-activities", "explicit-slot", "ordered-pipeline",
	}
	return positiveDefinitions(CaseGenerator, ids, "project-semantic-ir-through-go-generator")
}

func go127Definitions() []Definition {
	ids := []string{
		"generic-method", "generic-receiver-method", "generic-alias", "assignment-inference",
		"constrained-method", "generic-pair-method", "generic-codec-method", "alias-to-generic-type",
	}
	return positiveDefinitions(CaseGo127, ids, "reify-go-1.27-ast-and-types")
}

func positiveDefinitions(kind CaseKind, ids []string, operation string) []Definition {
	result := make([]Definition, 0, len(ids))
	for _, id := range ids {
		result = append(result, Definition{ID: string(kind) + ":" + id, Kind: kind, Fixture: id,
			ExpectedOutcome: "ACCEPT", ExpectedStage: "ACCEPTED", ProofChoice: "COHERENCE", MetaOperation: operation})
	}
	return result
}

func guardrailDefinitions() []Definition {
	return []Definition{
		guard("parse-error", "PARSE"),
		guard("type-mismatch", "TYPE"),
		guard("duplicate-declaration", "TYPE"),
		guard("undefined-identifier", "TYPE"),
		guard("import-authority", "AUTHORITY"),
		guard("unexported-api", "API"),
		guard("constraint-violation", "TYPE"),
		guard("unknown-payload", "REGISTRY"),
	}
}

func guard(id, stage string) Definition {
	return Definition{ID: "GUARDRAIL:" + id, Kind: CaseGuardrail, Fixture: id,
		ExpectedOutcome: "REJECT", ExpectedStage: stage, ProofChoice: "REGRESSION",
		MetaOperation: "fail-closed-go-interoperation-boundary"}
}
