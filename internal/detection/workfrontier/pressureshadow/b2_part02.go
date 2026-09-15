package pressureshadow

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/workfrontier/pressurecoverage"
)

func b2Issues(input Input) ([]RequiredPressureSetIssue, []RequiredPressureSetIssue,
	[]RequiredPressureSetIssue) {
	selectorIDs := make([]string, 0, len(input.Selector.Pressures))
	for _, pressure := range input.Selector.Pressures {
		selectorIDs = append(selectorIDs, pressureID(pressure))
	}
	selectorIDs = sortedStrings(selectorIDs)
	paths := make(map[string][]string, len(input.Selector.Paths))
	for _, path := range input.Selector.Paths {
		paths[pathID(path)] = path.RequiredPressureIDs
	}
	rows := coverageRows(input)
	ids := make([]string, 0, len(paths))
	for id := range paths {
		ids = append(ids, id)
	}
	var missingRecords, missingSelector, unregistered []RequiredPressureSetIssue
	for _, id := range sortedStrings(ids) {
		recordIDs := pressureRecordIDs(rows[id].Coverage.PressureRecords)
		if values := pressureDifference(paths[id], recordIDs); len(values) > 0 {
			missingRecords = append(missingRecords, RequiredPressureSetIssue{id, values})
		}
		if values := pressureDifference(paths[id], selectorIDs); len(values) > 0 {
			missingSelector = append(missingSelector, RequiredPressureSetIssue{id, values})
		}
		if values := pressureDifference(recordIDs, selectorIDs); len(values) > 0 {
			unregistered = append(unregistered, RequiredPressureSetIssue{id, values})
		}
	}
	return missingRecords, missingSelector, unregistered
}
func pressureRecordIDs(records []pressurecoverage.PressureRecord) []string {
	ids := make([]string, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.PressureID)
	}
	return sortedStrings(ids)
}
func fromB2Upstream(upstream B1Result) B2Result {
	decision, reason := DecisionUnknown, ReasonUpstreamUnknown
	if upstream.Decision == DecisionFailClosed {
		decision, reason = DecisionFailClosed, ReasonUpstreamFailClosed
	}
	return finishB2(newB2Result(upstream), decision, reason)
}
func newB2Result(upstream B1Result) B2Result {
	return B2Result{
		Schema:               SchemaVersion,
		InputDigest:          upstream.InputDigest,
		UpstreamResultDigest: upstream.ResultDigest,
		EnforcementEffect:    EnforcementNoEffect,
	}
}
func finishB2(result B2Result, decision Decision, reason Reason) B2Result {
	result.Decision, result.Reason = decision, reason
	result = normalizeB2Result(result)
	result.ResultDigest = CanonicalB2ResultDigest(result)
	result.ReplayDigest = digestBytes([]byte("b2-replay\x00" + result.InputDigest + "\x00" +
		result.UpstreamResultDigest + "\x00" + result.ResultDigest))
	return result
}
