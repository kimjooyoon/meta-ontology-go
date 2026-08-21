package pressureshadow

import (
	"sort"
)

func evaluateB1(input Input, upstream Result) B1Result {
	if upstream.Decision != DecisionPass {
		return fromUpstream(upstream)
	}
	missing, extra, missingK, mismatches := b1Issues(input)
	result := newB1Result(upstream)
	result.MissingRequiredPressureIDs = missing
	result.ExtraRequiredPressureIDs = extra
	result.MissingKPathIDs = missingK
	result.RequestedKIssues = mismatches
	switch {
	case len(extra) > 0:
		return finishB1(result, DecisionFailClosed, ReasonRequiredSetExtra)
	case len(mismatches) > 0:
		return finishB1(result, DecisionFailClosed, ReasonRequestedKMismatch)
	case len(missing) > 0:
		return finishB1(result, DecisionUnknown, ReasonRequiredSetMissing)
	case len(missingK) > 0:
		return finishB1(result, DecisionUnknown, ReasonRequestedKMissing)
	default:
		return finishB1(result, DecisionPass, ReasonNone)
	}
}
func b1Issues(input Input) ([]RequiredPressureSetIssue, []RequiredPressureSetIssue,
	[]string, []RequestedKIssue) {
	paths := make(map[string][]string, len(input.Selector.Paths))
	for _, path := range input.Selector.Paths {
		paths[pathID(path)] = path.RequiredPressureIDs
	}
	rows := coverageRows(input)
	missing, extra := []RequiredPressureSetIssue{}, []RequiredPressureSetIssue{}
	missingK, mismatches := []string{}, []RequestedKIssue{}
	ids := make([]string, 0, len(paths))
	for id := range paths {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		row := rows[id]
		if ids := pressureDifference(paths[id], row.Coverage.RequiredPressureIDs); len(ids) > 0 {
			missing = append(missing, RequiredPressureSetIssue{PathID: id, PressureIDs: ids})
		}
		if ids := pressureDifference(row.Coverage.RequiredPressureIDs, paths[id]); len(ids) > 0 {
			extra = append(extra, RequiredPressureSetIssue{PathID: id, PressureIDs: ids})
		}
		selectorK, coverageK := uint64(input.Selector.MinimumSelectedPressures), row.Coverage.RequestedK
		if selectorK == 0 || coverageK == 0 {
			missingK = append(missingK, id)
		} else if selectorK != coverageK {
			mismatches = append(mismatches, RequestedKIssue{id, selectorK, coverageK})
		}
	}
	return missing, extra, missingK, mismatches
}
func pressureDifference(left, right []string) []string {
	known := make(map[string]struct{}, len(right))
	for _, id := range right {
		known[id] = struct{}{}
	}
	result := []string{}
	for _, id := range left {
		if _, exists := known[id]; !exists {
			result = append(result, id)
		}
	}
	return sortedStrings(result)
}
