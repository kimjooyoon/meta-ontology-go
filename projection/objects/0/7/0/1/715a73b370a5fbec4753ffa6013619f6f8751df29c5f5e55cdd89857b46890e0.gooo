package pressureshadow

import (
	"encoding/json"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier"
)

func s1b2Decision(result S1B2Result, upstream S1B1Result) (Decision, Reason) {
	if len(result.UnsafeSelectedFailPathIDs) > 0 {
		return DecisionFailClosed, ReasonSelectorSelectedFailedPressureCoverage
	}
	if len(upstream.PressureCoverageFailPathIDs) > 0 {
		return DecisionFailClosed, ReasonPressureCoverageFailClosed
	}
	if len(result.UnsafeSelectedUnknownPathIDs) > 0 {
		return DecisionUnknown, ReasonSelectorSelectedUnknownPressureCoverage
	}
	if len(upstream.PressureCoverageUnknownPathIDs) > 0 {
		return DecisionUnknown, ReasonPressureCoverageUnknown
	}
	if result.SelectorResult != nil && result.SelectorResult.Status == workfrontier.DecisionUnknown {
		return DecisionUnknown, ReasonSelectorUnknown
	}
	return DecisionPass, ReasonNone
}
func newS1B2Result(upstream S1B1Result) S1B2Result {
	return S1B2Result{
		Schema: SchemaVersion, InputDigest: upstream.InputDigest,
		UpstreamResultDigest: upstream.ResultDigest, UnsafeSelectedUnknownPathIDs: []string{},
		UnsafeSelectedFailPathIDs: []string{}, EnforcementEffect: EnforcementNoEffect,
	}
}
func finishS1B2(result S1B2Result, decision Decision, reason Reason) S1B2Result {
	result.Decision, result.Reason = decision, reason
	result = normalizeS1B2Result(result)
	result.ResultDigest = CanonicalS1B2ResultDigest(result)
	result.ReplayDigest = digestBytes([]byte("s1b2-replay\x00" + result.InputDigest + "\x00" +
		result.UpstreamResultDigest + "\x00" + result.ResultDigest))
	return result
}
func canonicalSelectorDigest(result workfrontier.Result) string {
	data, _ := json.Marshal(result)
	return digestBytes(append([]byte("s1b2-selector-result\x00"), data...))
}

// CanonicalS1B2ResultDigest binds selector evidence and both unsafe sets.
func CanonicalS1B2ResultDigest(result S1B2Result) string {
	result = normalizeS1B2Result(result)
	result.ResultDigest, result.ReplayDigest = "", ""
	data, _ := json.Marshal(result)
	return digestBytes(data)
}
func normalizeS1B2Result(result S1B2Result) S1B2Result {
	result.UnsafeSelectedUnknownPathIDs = sortedStrings(result.UnsafeSelectedUnknownPathIDs)
	result.UnsafeSelectedFailPathIDs = sortedStrings(result.UnsafeSelectedFailPathIDs)
	return result
}
func intersectSorted(selected, unsafe []string) []string {
	known := make(map[string]struct{}, len(unsafe))
	for _, id := range unsafe {
		known[id] = struct{}{}
	}
	result := make([]string, 0, len(selected))
	for _, id := range selected {
		if _, ok := known[id]; ok {
			result = append(result, id)
		}
	}
	result = sortedStrings(result)
	if len(result) == 0 {
		return []string{}
	}
	return result
}
