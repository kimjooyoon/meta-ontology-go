package pathclosure

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

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
