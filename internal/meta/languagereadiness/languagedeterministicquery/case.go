package languagedeterministicquery

func finishCase(definition Definition, evidence Evidence, satisfied bool) CaseResult {
	status := StatusNotSatisfied
	if satisfied {
		status = StatusSatisfied
	}
	result := CaseResult{Definition: definition, Evidence: evidence, Status: status}
	result.Digest = digestJSON(result)
	return result
}

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
