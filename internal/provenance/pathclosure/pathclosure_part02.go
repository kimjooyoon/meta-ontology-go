package pathclosure

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

// Evaluate validates path semantics before evaluating any requirement. It
// indexes each normalized edge once and checks only the exact named sequence.
func Evaluate(path semantic.InferencePathV1, requirements []Requirement) Result {
	normalized, pathErr := path.Normalized()
	result := Result{Denominator: len(requirements)}
	states, required := prepareRequirements(requirements)
	result.Required = required
	if len(requirements) == 0 {
		result.Status, result.Code = UNKNOWN, CodeZeroDenominator
		return result
	}
	if pathErr != nil {
		return resultForSemanticError(result, states, pathErr)
	}

	edges, duplicateEdges := indexEdges(normalized.Edges)
	for _, state := range states {
		if state.duplicate {
			result.Duplicate = appendID(result.Duplicate, state.normalized.PathID)
			continue
		}
		if state.malformed {
			result.Malformed = appendID(result.Malformed, state.normalized.PathID)
			continue
		}
		outcome := evaluateRequirement(state.normalized, edges)
		switch outcome {
		case issueDuplicate:
			result.Duplicate = appendID(result.Duplicate, state.normalized.PathID)
		case issueMalformed:
			result.Malformed = appendID(result.Malformed, state.normalized.PathID)
		case issueMissingEvidence, issueMissingSnapshot:
			result.Missing = appendID(result.Missing, state.normalized.PathID)
		default:
			result.Complete = appendID(result.Complete, state.normalized.PathID)
		}
	}
	result.Duplicate = appendIDs(result.Duplicate, duplicateEdges)
	sortResultIDs(&result)
	result.Numerator = len(result.Complete)
	result.Status, result.Code = decision(result)
	return result
}
