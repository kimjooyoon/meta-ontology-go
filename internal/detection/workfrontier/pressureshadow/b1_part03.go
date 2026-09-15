package pressureshadow

import (
	"encoding/json"
	"sort"
)

func fromUpstream(upstream Result) B1Result {
	decision, reason := DecisionUnknown, ReasonUpstreamUnknown
	if upstream.Decision == DecisionFailClosed {
		decision, reason = DecisionFailClosed, ReasonUpstreamFailClosed
	}
	return finishB1(newB1Result(upstream), decision, reason)
}
func newB1Result(upstream Result) B1Result {
	return B1Result{
		Schema:                     SchemaVersion,
		InputDigest:                upstream.InputDigest,
		UpstreamResultDigest:       upstream.ResultDigest,
		MissingRequiredPressureIDs: []RequiredPressureSetIssue{},
		ExtraRequiredPressureIDs:   []RequiredPressureSetIssue{},
		MissingKPathIDs:            []string{},
		RequestedKIssues:           []RequestedKIssue{},
		EnforcementEffect:          EnforcementNoEffect,
	}
}
func finishB1(result B1Result, decision Decision, reason Reason) B1Result {
	result.Decision, result.Reason = decision, reason
	result = normalizeB1Result(result)
	result.ResultDigest = CanonicalB1ResultDigest(result)
	result.ReplayDigest = digestBytes([]byte("b1-replay\x00" + result.InputDigest + "\x00" +
		result.UpstreamResultDigest + "\x00" + result.ResultDigest))
	return result
}

// CanonicalB1ResultDigest binds the upstream result and every mapping issue.
func CanonicalB1ResultDigest(result B1Result) string {
	result = normalizeB1Result(result)
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}
func normalizeB1Result(result B1Result) B1Result {
	result.MissingRequiredPressureIDs = normalizeSetIssues(result.MissingRequiredPressureIDs)
	result.ExtraRequiredPressureIDs = normalizeSetIssues(result.ExtraRequiredPressureIDs)
	result.MissingKPathIDs = sortedStrings(result.MissingKPathIDs)
	result.RequestedKIssues = normalizeKIssues(result.RequestedKIssues)
	return result
}
func normalizeSetIssues(issues []RequiredPressureSetIssue) []RequiredPressureSetIssue {
	result := append([]RequiredPressureSetIssue{}, issues...)
	for index := range result {
		result[index].PressureIDs = sortedStrings(result[index].PressureIDs)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].PathID < result[right].PathID
	})
	return result
}
func normalizeKIssues(issues []RequestedKIssue) []RequestedKIssue {
	result := append([]RequestedKIssue{}, issues...)
	sort.Slice(result, func(left, right int) bool {
		if result[left].PathID != result[right].PathID {
			return result[left].PathID < result[right].PathID
		}
		if result[left].SelectorK != result[right].SelectorK {
			return result[left].SelectorK < result[right].SelectorK
		}
		return result[left].CoverageK < result[right].CoverageK
	})
	return result
}
