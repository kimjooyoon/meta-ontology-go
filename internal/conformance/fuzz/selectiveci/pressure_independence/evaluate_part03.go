package pressureindependence

import (
	"sort"
)

func malformedLists(input Input) bool {
	if hasDuplicate(input.GuardIDs) || hasDuplicate(input.FinitePathIDs) {
		return true
	}
	for _, id := range input.RequiredPressureIDs {
		if !validToken(id) {
			return true
		}
	}
	for _, id := range append(append([]string{}, input.GuardIDs...), input.FinitePathIDs...) {
		if !validToken(id) {
			return true
		}
	}
	for _, record := range input.PressureRecords {
		if !validToken(record.PressureID) || !validToken(record.CategoryID) ||
			(record.IndependenceGroupID != "" && !validToken(record.IndependenceGroupID)) ||
			(record.ApplicabilityRuleID != "" && !validToken(record.ApplicabilityRuleID)) {
			return true
		}
	}
	return false
}
func selectRepresentatives(groups map[string][]string, requested uint64) []string {
	keys := make([]string, 0, len(groups))
	for group, ids := range groups {
		sort.Strings(ids)
		keys = append(keys, group)
	}
	sort.Strings(keys)
	limit := effectiveK(requested)
	if uint64(len(keys)) < limit {
		limit = uint64(len(keys))
	}
	selected := make([]string, 0, limit)
	for _, group := range keys[:limit] {
		selected = append(selected, groups[group][0])
	}
	sort.Strings(selected)
	return selected
}
func subtract(all, selected []string) []string {
	chosen := make(map[string]struct{}, len(selected))
	for _, id := range selected {
		chosen[id] = struct{}{}
	}
	result := make([]string, 0, len(all))
	for _, id := range all {
		if _, ok := chosen[id]; !ok {
			result = append(result, id)
		}
	}
	return sortedUnique(result)
}
func finishUnknown(output Output, input Input, decision Decision, reason Reason) Output {
	output.Decision, output.Reason = decision, reason
	output.FullSuiteRequired = true
	output.ProofValid = false
	output.UnknownIDs = sortedUnique(input.RequiredPressureIDs)
	output.SelectedIDs = []string{}
	output.UnselectedIDs = []string{}
	output.CostReceipt = receipt(input, 0)
	return seal(output)
}
