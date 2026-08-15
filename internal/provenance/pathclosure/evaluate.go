package pathclosure

import (
	"errors"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func indexEdges(edges []semantic.InferenceEdge) (map[semantic.ID]semantic.InferenceEdge, []semantic.ID) {
	indexed := make(map[semantic.ID]semantic.InferenceEdge, len(edges))
	duplicates := make([]semantic.ID, 0)
	for _, edge := range edges {
		if _, exists := indexed[edge.RecordID]; exists {
			duplicates = appendID(duplicates, edge.RecordID)
			continue
		}
		indexed[edge.RecordID] = edge
	}
	sortIDs(duplicates)
	return indexed, duplicates
}

func evaluateRequirement(requirement Requirement, edges map[semantic.ID]semantic.InferenceEdge) issueClass {
	var previous semantic.InferenceEdge
	for i, recordID := range requirement.RecordIDs {
		edge, exists := edges[recordID]
		if !exists {
			return issueMissingEvidence
		}
		if edge.Kind != requirement.ExpectedKinds[i] {
			return issueMalformed
		}
		if i == 0 {
			if edge.SubjectID != requirement.StartID {
				return issueMalformed
			}
		} else if previous.ObjectID != edge.SubjectID {
			return issueMalformed
		}
		previous = edge
		if i == len(requirement.RecordIDs)-1 && edge.ObjectID != requirement.EndID {
			return issueMalformed
		}
	}
	return 0
}

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
			current = issueDuplicate
		case "duplicate-evidence":
			current = issueDuplicate
		case "orphan-evidence":
			current = issueMissingEvidence
		case "edge", "evidence", "claim":
			lower := strings.ToLower(issue.Detail)
			if strings.Contains(lower, "duplicate") {
				current = issueDuplicate
			} else if strings.Contains(lower, "snapshot") && strings.Contains(lower, "required") {
				current = issueMissingSnapshot
			}
		}
		class = moreSevere(class, current)
	}
	if class == 0 {
		return issueMalformed
	}
	return class
}

func moreSevere(left, right issueClass) issueClass {
	if left == issueDuplicate || right == issueDuplicate {
		return issueDuplicate
	}
	if left == issueMalformed || right == issueMalformed {
		return issueMalformed
	}
	if left == issueMissingSnapshot || right == issueMissingSnapshot {
		return issueMissingSnapshot
	}
	return issueMissingEvidence
}

func decision(result Result) (Status, string) {
	if len(result.Duplicate) != 0 {
		return FAIL_CLOSED, CodeDuplicate
	}
	if len(result.Malformed) != 0 {
		return FAIL_CLOSED, CodeMalformed
	}
	if len(result.Missing) != 0 {
		return UNKNOWN, CodeMissingRecord
	}
	if result.Denominator == 0 {
		return UNKNOWN, CodeZeroDenominator
	}
	return PASS, CodePass
}

func decisionWithSemanticClass(result Result, class issueClass) (Status, string) {
	if len(result.Duplicate) != 0 {
		return FAIL_CLOSED, CodeDuplicate
	}
	if len(result.Malformed) != 0 {
		return FAIL_CLOSED, CodeMalformed
	}
	if class == issueDuplicate {
		return FAIL_CLOSED, CodeDuplicate
	}
	if class == issueMissingSnapshot {
		return UNKNOWN, CodeMissingSnapshot
	}
	if class == issueMissingEvidence {
		return UNKNOWN, CodeMissingEvidence
	}
	return FAIL_CLOSED, CodeInvalidSemantic
}

func appendID(ids []semantic.ID, id semantic.ID) []semantic.ID {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}

func appendIDs(ids, additions []semantic.ID) []semantic.ID {
	for _, id := range additions {
		ids = appendID(ids, id)
	}
	sortIDs(ids)
	return ids
}

func sortIDs(ids []semantic.ID) {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
}

func sortResultIDs(result *Result) {
	sortIDs(result.Required)
	sortIDs(result.Complete)
	sortIDs(result.Missing)
	sortIDs(result.Malformed)
	sortIDs(result.Duplicate)
}
