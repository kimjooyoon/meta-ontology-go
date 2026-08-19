package pressurecoverage

import (
	"sort"
)

func evaluateCoverage(result Result, input Input) Result {
	records := recordsByID(input.PressureRecords)
	groups := map[string]struct{}{}
	for _, id := range input.RequiredPressureIDs {
		record, ok := records[id]
		if !ok {
			result.MissingPressureIDs = append(result.MissingPressureIDs, id)
			continue
		}
		if record.IndependenceGroupID != "" {
			groups[record.IndependenceGroupID] = struct{}{}
		}
	}
	result.RequiredGroupIDs = groupIDs(groups)
	result.DistinctGroupCount = uint64(len(result.RequiredGroupIDs))
	if len(result.MissingPressureIDs) != 0 {
		return finish(result, DecisionUnknown, ReasonRequiredInputMissing)
	}
	if uint64(len(input.RequiredPressureIDs)) < input.RequestedK {
		return finish(result, DecisionUnknown, ReasonPressureCardinalityShortfall)
	}
	for _, id := range input.RequiredPressureIDs {
		record := records[id]
		if record.IndependenceGroupID == "" || record.ApplicabilityRuleID == "" {
			return finish(result, DecisionUnknown, ReasonApplicabilityOrGroupUnproven)
		}
	}
	if result.DistinctGroupCount < input.MinimumIndependent {
		return finish(result, DecisionUnknown, ReasonIndependentGroupShortfall)
	}
	return finish(result, DecisionPass, ReasonNone)
}
func newResult(input Input) Result {
	return Result{
		Schema:                SchemaVersion,
		InputDigest:           CanonicalInputDigest(input),
		RequiredPressureCount: uint64(len(input.RequiredPressureIDs)),
		RequiredPressureIDs:   resultIDs(input.RequiredPressureIDs),
	}
}
func blankBinding(input Input) bool {
	return input.AuthoritySnapshotDigest == "" || input.PolicyDigest == "" ||
		input.RegistryDigest == "" || input.ToolchainOptionsDigest == ""
}
func bindingMatches(input Input) bool {
	return input.AuthoritySnapshotDigest == authorityBindingDigest(input, "authority-snapshot") &&
		input.PolicyDigest == authorityBindingDigest(input, "policy") &&
		input.RegistryDigest == authorityBindingDigest(input, "registry") &&
		input.ToolchainOptionsDigest == authorityBindingDigest(input, "toolchain-options")
}
func recordsByID(records []PressureRecord) map[string]PressureRecord {
	result := make(map[string]PressureRecord, len(records))
	for _, record := range records {
		result[record.PressureID] = record
	}
	return result
}
func groupIDs(groups map[string]struct{}) []string {
	result := make([]string, 0, len(groups))
	for group := range groups {
		result = append(result, group)
	}
	sort.Strings(result)
	return result
}
