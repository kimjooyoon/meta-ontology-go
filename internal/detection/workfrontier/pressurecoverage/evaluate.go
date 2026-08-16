package pressurecoverage

import "sort"

// Observe computes a read-only, explicit pressure coverage observation.
func Observe(input Input) Output {
	output := Output{
		Schema:                SchemaVersion,
		InputDigest:           CanonicalInputDigest(input),
		RequiredPressureCount: uint64(len(input.RequiredPressureIDs)),
	}
	setCost(&output, input, 0)
	if reason := validateHeader(input); reason != "" {
		return finish(output, input, DecisionUnknown, reason)
	}
	records, decision, reason := validateRecords(input)
	if reason != "" {
		return finish(output, input, decision, reason)
	}
	groups, applicability := map[string][]string{}, ""
	for _, id := range input.RequiredPressureIDs {
		record := records[id]
		if applicability == "" {
			applicability = record.ApplicabilityRuleID
		} else if applicability != record.ApplicabilityRuleID {
			return finish(output, input, DecisionUnknown, ReasonInputAmbiguous)
		}
		groups[record.IndependenceGroupID] = append(groups[record.IndependenceGroupID], id)
	}
	output.DistinctGroupCount = uint64(len(groups))
	if uint64(len(groups)) < effectiveK(input.MinimumIndependent) ||
		uint64(len(groups)) < effectiveK(input.RequestedK) {
		return finish(output, input, DecisionUnknown, ReasonIndependentGroupShortfall)
	}
	output.SelectedIDs = selectRepresentatives(groups, input.RequestedK)
	output.UnselectedIDs = subtract(input.RequiredPressureIDs, output.SelectedIDs)
	setCost(&output, input, len(output.SelectedIDs))
	return finish(output, input, DecisionPass, ReasonNone)
}

func validateHeader(input Input) Reason {
	if missingRequiredInput(input) {
		return ReasonRequiredInputMissing
	}
	if input.Schema != SchemaVersion || !validDigests(input) || !boundDigests(input) {
		return ReasonStaleDigest
	}
	return ""
}

func missingRequiredInput(input Input) bool {
	return input.Schema == "" || input.AuthoritySnapshotDigest == "" ||
		input.PolicyDigest == "" || input.RegistryDigest == "" ||
		input.ToolchainOptionsDigest == "" || input.RequestedK == 0 ||
		input.MinimumIndependent == 0 || len(input.PressureRecords) == 0 ||
		len(input.RequiredPressureIDs) == 0 || len(input.FinitePathIDs) == 0 ||
		len(input.GuardIDs) == 0
}

func validDigests(input Input) bool {
	return validDigest(input.AuthoritySnapshotDigest) &&
		validDigest(input.PolicyDigest) && validDigest(input.RegistryDigest) &&
		validDigest(input.ToolchainOptionsDigest)
}

func validateRecords(input Input) (map[string]PressureRecord, Decision, Reason) {
	if reason := validateLists(input); reason != "" {
		return nil, listDecision(reason), reason
	}
	records := make(map[string]PressureRecord, len(input.PressureRecords))
	for _, record := range input.PressureRecords {
		decision, reason := addRecord(records, record)
		if reason != "" {
			return nil, decision, reason
		}
	}
	for _, id := range input.RequiredPressureIDs {
		if _, ok := records[id]; !ok {
			return nil, DecisionUnknown, ReasonRequiredInputMissing
		}
	}
	return records, DecisionPass, ""
}

func validateLists(input Input) Reason {
	lists := []struct {
		values    []string
		duplicate Reason
		malformed Reason
	}{
		{input.RequiredPressureIDs, ReasonDuplicateID, ReasonInvalidStableID},
		{input.GuardIDs, ReasonDuplicateID, ReasonInvalidStableID},
		{input.FinitePathIDs, ReasonMalformedFinitePath, ReasonMalformedFinitePath},
	}
	for _, list := range lists {
		if reason := listReason(list.values, list.duplicate, list.malformed); reason != "" {
			return reason
		}
	}
	return ""
}

func addRecord(records map[string]PressureRecord, record PressureRecord) (Decision, Reason) {
	if !validRecord(record) {
		return DecisionFailClosed, ReasonInvalidStableID
	}
	if record.IndependenceGroupID == "" || record.ApplicabilityRuleID == "" {
		return DecisionUnknown, ReasonApplicabilityUnproven
	}
	if prior, exists := records[record.PressureID]; exists {
		if prior == record {
			return DecisionFailClosed, ReasonDuplicatePressureID
		}
		return DecisionFailClosed, ReasonConflictingGroupBinding
	}
	records[record.PressureID] = record
	return DecisionPass, ""
}

func validRecord(record PressureRecord) bool {
	return validID(record.PressureID) && validID(record.CategoryID) &&
		optionalID(record.IndependenceGroupID) && optionalID(record.ApplicabilityRuleID)
}

func optionalID(value string) bool {
	return value == "" || validID(value)
}

func listReason(values []string, duplicate, malformed Reason) Reason {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return duplicate
		}
		if !validID(value) {
			return malformed
		}
		seen[value] = struct{}{}
	}
	return ""
}

func listDecision(reason Reason) Decision {
	if reason == ReasonDuplicateID || reason == ReasonInvalidStableID ||
		reason == ReasonMalformedFinitePath {
		return DecisionFailClosed
	}
	return DecisionUnknown
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
	return sortedUnique(selected)
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

func finish(output Output, input Input, decision Decision, reason Reason) Output {
	output.Decision, output.Reason = decision, reason
	output.FullSuiteRequired = decision != DecisionPass
	if decision != DecisionPass {
		output.SelectedIDs, output.UnselectedIDs = []string{}, []string{}
		output.UnknownIDs = sortedUnique(input.RequiredPressureIDs)
	}
	return seal(output)
}

func seal(output Output) Output {
	output.SelectedIDs = sortedUnique(output.SelectedIDs)
	output.UnselectedIDs = sortedUnique(output.UnselectedIDs)
	output.UnknownIDs = sortedUnique(output.UnknownIDs)
	output.OutputDigest = CanonicalOutputDigest(output)
	output.ReplayDigest = digestBytes([]byte(output.InputDigest + "\x00" + output.OutputDigest))
	return output
}

func effectiveK(value uint64) uint64 {
	if value < LanguageFloor {
		return LanguageFloor
	}
	return value
}

func setCost(output *Output, input Input, selected int) {
	pressureCount := uint64(len(input.PressureRecords))
	requiredCount := uint64(len(input.RequiredPressureIDs))
	guardCount := uint64(len(input.GuardIDs))
	pathCount := uint64(len(input.FinitePathIDs))
	output.DeterministicWorkUnits = pressureCount + 2*requiredCount + guardCount + pathCount
	output.CPUCoreNS = uint64(selected)
	output.MemoryBytes = 1024 * requiredCount
	output.ProvRecords = pressureCount + guardCount
	output.ProvPaths = pathCount
}
