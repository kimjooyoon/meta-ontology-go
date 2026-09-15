package languagepackageruntime

import (
	"errors"

	"github.com/kimjooyoon/meta-ontology-go/internal/packageruntime"
)

func executeGuardrail(definition Definition) CaseResult {
	result := CaseResult{ID: definition.ID, Kind: definition.Kind, Expected: definition.ExpectedCode}
	_, err := packageruntime.Run(manifestFor(definition))
	if err == nil {
		result.Observed, result.Reason = "ACCEPTED", "INVALID_RUNTIME_ACCEPTED"
		return result
	}
	result.Observed, result.Reason = failureCode(err), err.Error()
	result.Satisfied = result.Observed == definition.ExpectedCode
	if result.Satisfied {
		result.Reason = "INVALID_RUNTIME_REJECTED"
	}
	return result
}

func failureCode(err error) string {
	failure := new(packageruntime.Failure)
	if errors.As(err, &failure) {
		return failure.Code
	}
	return "UNKNOWN_FAILURE"
}
