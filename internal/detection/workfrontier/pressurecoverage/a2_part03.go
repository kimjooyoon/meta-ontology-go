package pressurecoverage

import (
	"encoding/json"
	"sort"
)

func resultIDs(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	if len(result) == 0 {
		return []string{}
	}
	return result
}
func CanonicalResultDigest(result Result) string {
	result.RequiredPressureIDs = resultIDs(result.RequiredPressureIDs)
	result.RequiredGroupIDs = resultIDs(result.RequiredGroupIDs)
	result.MissingPressureIDs = resultIDs(result.MissingPressureIDs)
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}
func finish(result Result, decision Decision, reason Reason) Result {
	result.Decision, result.Reason = decision, reason
	result.RequiredPressureIDs = resultIDs(result.RequiredPressureIDs)
	result.RequiredGroupIDs = resultIDs(result.RequiredGroupIDs)
	result.MissingPressureIDs = resultIDs(result.MissingPressureIDs)
	result.ResultDigest = CanonicalResultDigest(result)
	result.ReplayDigest = digestBytes([]byte(result.InputDigest + "\x00" + result.ResultDigest))
	return result
}
