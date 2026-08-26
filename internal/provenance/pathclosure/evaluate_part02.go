package pathclosure

import (
	"errors"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
)

func resultForSemanticError(result Result, states []requirementState, err error) Result {
	class := classifySemanticError(err)
	for _, state := range states {
		id := state.normalized.PathID
		switch {
		case state.duplicate:
			result.Duplicate = appendID(result.Duplicate, id)
		case state.malformed:
			result.Malformed = appendID(result.Malformed, id)
		case class == issueDuplicate:
			result.Duplicate = appendID(result.Duplicate, id)
		case class == issueMissingEvidence, class == issueMissingSnapshot:
			result.Missing = appendID(result.Missing, id)
		default:
			result.Malformed = appendID(result.Malformed, id)
		}
	}
	sortResultIDs(&result)
	result.Numerator = 0
	result.Status, result.Code = decisionWithSemanticClass(result, class)
	return result
}
func classifySemanticError(err error) issueClass {
	var pathErrors *semantic.InferencePathErrors
	if !errors.As(err, &pathErrors) {
		return issueMalformed
	}
	var class issueClass
	for _, issue := range pathErrors.Issues {
		current := issueMalformed
		switch issue.Code {
		case "stable-id-collision":
			if strings.Contains(strings.ToLower(issue.Detail), "evidence") {
				current = issueMissingEvidence
			} else {
				current = issueDuplicate
			}
		case "duplicate-evidence", "orphan-evidence", "independent-evidence",
			"missing-acceptance-receipt", "orphan-acceptance-receipt", "unbacked-acceptance-receipt":
			current = issueMissingEvidence
		case "edge", "evidence", "claim":
			lower := strings.ToLower(issue.Detail)
			if strings.Contains(lower, "snapshot") && strings.Contains(lower, "required") {
				current = issueMissingSnapshot
			} else if strings.Contains(lower, "evidence") &&
				(strings.Contains(lower, "required") || strings.Contains(lower, "duplicate")) {
				current = issueMissingEvidence
			}
		}
		class = moreSevere(class, current)
	}
	if class == 0 {
		return issueMalformed
	}
	return class
}
