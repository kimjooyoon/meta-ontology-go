package pressureshadow

import (
	"encoding/json"
)

func makeResult(inputDigest string, decision Decision, reason Reason, missing, orphan,
	missingBinding, mismatch []string) Result {
	result := Result{
		Schema:                 SchemaVersion,
		InputDigest:            inputDigest,
		Decision:               decision,
		Reason:                 reason,
		MissingPathIDs:         sortedStrings(missing),
		OrphanPathIDs:          sortedStrings(orphan),
		MissingBindingPathIDs:  sortedStrings(missingBinding),
		BindingMismatchPathIDs: sortedStrings(mismatch),
		EnforcementEffect:      EnforcementNoEffect,
	}
	result.ResultDigest = CanonicalResultDigest(result)
	result.ReplayDigest = digestBytes([]byte("replay\x00" + inputDigest + "\x00" + result.ResultDigest))
	return result
}

// CanonicalResultDigest binds the decision, reason, sets, and input digest.
func CanonicalResultDigest(result Result) string {
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}
