package pressurecoverage

import "sort"

// Observe computes a read-only, explicit pressure coverage observation.
func Observe(input Input) Output {
	output := Output{Schema: SchemaVersion, InputDigest: CanonicalInputDigest(input), RequiredPressureCount: uint64(len(input.RequiredPressureIDs))}
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
	if uint64(len(groups)) < effectiveK(input.MinimumIndependent) || uint64(len(groups)) < effectiveK(input.RequestedK) {
		return finish(output, input, DecisionUnknown, ReasonIndependentGroupShortfall)
	}
	output.SelectedIDs = selectRepresentatives(groups, input.RequestedK)
	output.UnselectedIDs = subtract(input.RequiredPressureIDs, output.SelectedIDs)
	setCost(&output, input, len(output.SelectedIDs))
	return finish(output, input, DecisionPass, ReasonNone)
}

func validateHeader(input Input) Reason {
	if input.Schema == "" || input.AuthoritySnapshotDigest == "" || input.PolicyDigest == "" || input.RegistryDigest == "" || input.ToolchainOptionsDigest == "" || input.RequestedK == 0 || input.MinimumIndependent == 0 || input.PressureRecords == nil || input.RequiredPressureIDs == nil || input.FinitePathIDs == nil || input.GuardIDs == nil {
		return ReasonRequiredInputMissing
	}
	if input.Schema != SchemaVersion || !validDigest(input.AuthoritySnapshotDigest) || !validDigest(input.PolicyDigest) || !validDigest(input.RegistryDigest) || !validDigest(input.ToolchainOptionsDigest) {
		return ReasonStaleDigest
	}
	return ""
}

func validateRecords(input Input) (map[string]PressureRecord, Decision, Reason) {
	lists := []struct {
		values    []string
		duplicate Reason
		malformed Reason
	}{
		{input.RequiredPressureIDs, ReasonDuplicateID, ReasonCatalogMismatch},
		{input.GuardIDs, ReasonDuplicateID, ReasonCatalogMismatch},
		{input.FinitePathIDs, ReasonMalformedFinitePath, ReasonMalformedFinitePath},
	}
	for _, list := range lists {
		if reason := listReason(list.values, list.duplicate, list.malformed); reason != "" {
			return nil, listDecision(reason), reason
		}
	}
	records := make(map[string]PressureRecord, len(input.PressureRecords))
	for _, record := range input.PressureRecords {
		if !validID(record.PressureID) || !validID(record.CategoryID) || (record.IndependenceGroupID != "" && !validID(record.IndependenceGroupID)) || (record.ApplicabilityRuleID != "" && !validID(record.ApplicabilityRuleID)) {
			return nil, DecisionFailClosed, ReasonCatalogMismatch
		}
		if record.IndependenceGroupID == "" || record.ApplicabilityRuleID == "" {
			return nil, DecisionUnknown, ReasonApplicabilityUnproven
		}
		if prior, exists := records[record.PressureID]; exists {
			if prior.CategoryID == record.CategoryID && prior.IndependenceGroupID == record.IndependenceGroupID && prior.ApplicabilityRuleID == record.ApplicabilityRuleID {
				return nil, DecisionFailClosed, ReasonDuplicateID
			}
			return nil, DecisionFailClosed, ReasonConflictingGroupBinding
		}
		records[record.PressureID] = record
	}
	for _, id := range input.RequiredPressureIDs {
		if _, ok := records[id]; !ok {
			return nil, DecisionUnknown, ReasonRequiredInputMissing
		}
	}
	return records, DecisionPass, ""
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
	if reason == ReasonDuplicateID || reason == ReasonMalformedFinitePath {
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
	output.DeterministicWorkUnits = uint64(len(input.PressureRecords)) + 2*uint64(len(input.RequiredPressureIDs)) + uint64(len(input.GuardIDs)) + uint64(len(input.FinitePathIDs))
	output.CPUCoreNS = uint64(selected)
	output.MemoryBytes = 1024 * uint64(len(input.RequiredPressureIDs))
	output.ProvRecords = uint64(len(input.PressureRecords)) + uint64(len(input.GuardIDs))
	output.ProvPaths = uint64(len(input.FinitePathIDs))
}
