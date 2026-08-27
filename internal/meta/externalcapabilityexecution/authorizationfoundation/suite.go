package authorizationfoundation

import "errors"

func EvaluateSuite(input Input) Suite {
	result := Suite{Schema: SuiteSchema, SubjectSHA: input.ExpectedSubject, Decision: "PASS", Resolution: "EXACT"}
	for _, definition := range suiteDefinitions() {
		actualDecision, actualResolution := classify(suiteInput(input, definition.Mutation))
		passed := actualDecision == definition.Decision && actualResolution == definition.Resolution
		result.Cases = append(result.Cases, SuiteCase{CaseID: definition.ID,
			ExpectedDecision: definition.Decision, ExpectedResolution: definition.Resolution,
			ActualDecision: actualDecision, ActualResolution: actualResolution, Passed: passed})
		result.Total++
		if passed {
			result.Passed++
		} else {
			result.Decision, result.Resolution = "FAIL_CLOSED", "UNKNOWN"
		}
		switch actualDecision {
		case "AUTHORIZED_SHADOW":
			result.AuthorizedCases++
		case "FAIL_CLOSED":
			result.UnknownCases++
		case "DENIED":
			result.DeniedCases++
		}
	}
	if result.Total > 0 {
		result.CoverageBPS = result.Passed * 10000 / result.Total
	}
	return sealSuite(result)
}

func classify(input Input) (string, string) {
	if _, err := Evaluate(input); err == nil {
		return "AUTHORIZED_SHADOW", "EXACT"
	} else {
		if value, ok := errors.AsType[*resolutionError](err); ok {
			return value.Decision, value.Resolution
		}
	}
	return "FAIL_CLOSED", "UNKNOWN"
}
