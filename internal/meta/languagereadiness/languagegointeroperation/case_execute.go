package languagegointeroperation

func executeCases(definitions []Definition) []CaseResult {
	results := make([]CaseResult, 0, len(definitions))
	for _, definition := range definitions {
		switch definition.Kind {
		case CaseGenerator:
			results = append(results, executeGeneratorCase(definition))
		case CaseGo127:
			results = append(results, executeGo127Case(definition))
		case CaseGuardrail:
			results = append(results, executeGuardrailCase(definition))
		default:
			results = append(results, finishCase(definition, rejectedEvidence("REGISTRY", "CASE_KIND_UNKNOWN"), false))
		}
	}
	return results
}

func finishCase(definition Definition, evidence Evidence, satisfied bool) CaseResult {
	status := StatusNotSatisfied
	if satisfied {
		status = StatusSatisfied
	}
	result := CaseResult{Definition: definition, Evidence: evidence, Status: status}
	result.Digest = digestJSON(result)
	return result
}

func rejectedEvidence(stage, code string) Evidence {
	return Evidence{ActualOutcome: "REJECT", FailureStage: stage, ErrorCode: code, Rejected: true}
}
