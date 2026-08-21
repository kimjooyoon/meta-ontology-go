package pressureshadow

import (
	"encoding/json"
)

// CanonicalB2ResultDigest binds upstream state and all three issue sets.
func CanonicalB2ResultDigest(result B2Result) string {
	result = normalizeB2Result(result)
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}
func normalizeB2Result(result B2Result) B2Result {
	result.MissingRequiredPressureRecordIDs = normalizeSetIssues(result.MissingRequiredPressureRecordIDs)
	result.MissingSelectorPressureIDs = normalizeSetIssues(result.MissingSelectorPressureIDs)
	result.UnregisteredPressureRecordIDs = normalizeSetIssues(result.UnregisteredPressureRecordIDs)
	return result
}
