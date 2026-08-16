package pressurecoverage

import "sort"

type validation struct {
	records map[string]PressureRecord
	reason  Reason
}

// Observe computes a read-only, explicit pressure coverage observation.
func Observe(input Input) Output {
	output := Output{Schema: SchemaVersion, InputDigest: CanonicalInputDigest(input),
		RequiredPressureCount: uint64(len(input.RequiredPressureIDs))}
	output.DeterministicWorkUnits, output.CPUCoreNS, output.MemoryBytes, output.ProvRecords, output.ProvPaths = cost(input, 0)
	if reason := validateHeader(input); reason != "" {
		return finish(output, input, DecisionUnknown, reason)
	}
	check := validateRecords(input)
	if check.reason != "" {
		decision := DecisionUnknown
		if check.reason == ReasonDuplicateID || check.reason == ReasonDuplicatePressureID ||
			check.reason == ReasonConflictingGroupBinding || check.reason == ReasonMalformedIdentity ||
			check.reason == ReasonMalformedFinitePath || check.reason == ReasonCatalogMismatch {
			decision = DecisionFailClosed
		}
		return finish(output, input, decision, check.reason)
	}
	groups := make(map[string][]string)
	applicability := ""
	for _, id := range input.RequiredPressureIDs {
		record := check.records[id]
		if record.IndependenceGroupID == "" || record.ApplicabilityRuleID == "" {
			return finish(output, input, DecisionUnknown, ReasonApplicabilityUnproven)
		}
		if applicability == "" {
			applicability = record.ApplicabilityRuleID
		} else if applicability != record.ApplicabilityRuleID {
			return finish(output, input, DecisionUnknown, ReasonInputAmbiguous)
		}
		groups[record.IndependenceGroupID] = append(groups[record.IndependenceGroupID], id)
	}
	output.DistinctGroupCount = uint64(len(groups))
	if uint64(len(groups)) < effectiveMinimum(input.MinimumIndependent) || uint64(len(groups)) < effectiveK(input.RequestedK) {
		return finish(output, input, DecisionUnknown, ReasonIndependentGroupShortfall)
	}
	output.SelectedIDs = selectRepresentatives(groups, input.RequestedK)
	output.UnselectedIDs = subtract(input.RequiredPressureIDs, output.SelectedIDs)
	output.DeterministicWorkUnits, output.CPUCoreNS, output.MemoryBytes, output.ProvRecords, output.ProvPaths = cost(input, len(output.SelectedIDs))
	if reason := ceilingReason(output, input.ResourceCeilings); reason != "" {
		return finish(output, input, DecisionUnknown, reason)
	}
	output.Decision, output.Reason = DecisionPass, ReasonNone
	return seal(output)
}

func Evaluate(input Input) Output { return Observe(input) }

func validateHeader(input Input) Reason {
	if input.Schema == "" || input.AuthoritySnapshotDigest == "" || input.PolicyDigest == "" ||
		input.RegistryDigest == "" || input.ToolchainOptionsDigest == "" || input.RequestedK == 0 ||
		input.MinimumIndependent == 0 || input.PressureRecords == nil || input.RequiredPressureIDs == nil ||
		input.FinitePathIDs == nil || input.GuardIDs == nil {
		return ReasonRequiredInputMissing
	}
	if missingCeiling(input.ResourceCeilings) {
		return ReasonResourceCeilingMissing
	}
	if input.Schema != SchemaVersion || !validDigest(input.AuthoritySnapshotDigest) || !validDigest(input.PolicyDigest) ||
		!validDigest(input.RegistryDigest) || !validDigest(input.ToolchainOptionsDigest) {
		return ReasonStaleDigest
	}
	return ""
}

func validateRecords(input Input) validation {
	if hasDuplicate(input.RequiredPressureIDs) || hasDuplicate(input.GuardIDs) {
		return validation{reason: ReasonDuplicateID}
	}
	if reason := validateIDs(input.RequiredPressureIDs); reason != "" {
		return validation{reason: ReasonMalformedIdentity}
	}
	if reason := validateFinitePaths(input.FinitePathIDs); reason != "" {
		return validation{reason: reason}
	}
	if reason := validateIDs(input.GuardIDs); reason != "" {
		return validation{reason: ReasonMalformedIdentity}
	}
	records := make(map[string]PressureRecord, len(input.PressureRecords))
	for _, record := range input.PressureRecords {
		if !validID(record.PressureID) || !validID(record.CategoryID) ||
			!validID(record.IndependenceGroupID) && record.IndependenceGroupID != "" ||
			!validID(record.ApplicabilityRuleID) && record.ApplicabilityRuleID != "" {
			return validation{reason: ReasonMalformedIdentity}
		}
		if record.IndependenceGroupID == "" || record.ApplicabilityRuleID == "" {
			return validation{reason: ReasonApplicabilityUnproven}
		}
		if prior, exists := records[record.PressureID]; exists {
			if prior.PressureID == record.PressureID && prior.CategoryID == record.CategoryID &&
				prior.IndependenceGroupID == record.IndependenceGroupID && prior.ApplicabilityRuleID == record.ApplicabilityRuleID {
				return validation{reason: ReasonDuplicatePressureID}
			}
			return validation{reason: ReasonConflictingGroupBinding}
		}
		records[record.PressureID] = record
	}
	for _, id := range input.RequiredPressureIDs {
		if _, ok := records[id]; !ok {
			return validation{reason: ReasonCatalogMismatch}
		}
	}
	return validation{records: records}
}

func validateFinitePaths(values []string) Reason {
	if hasDuplicate(values) {
		return ReasonMalformedFinitePath
	}
	if reason := validateIDs(values); reason != "" {
		return ReasonMalformedFinitePath
	}
	return ""
}

func validateIDs(values []string) Reason {
	for _, value := range values {
		if !validID(value) {
			return ReasonMalformedIdentity
		}
	}
	return ""
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
	output.Decision, output.Reason, output.FullSuiteRequired = decision, reason, true
	output.SelectedIDs, output.UnselectedIDs, output.UnknownIDs = []string{}, []string{}, sortedUnique(input.RequiredPressureIDs)
	return seal(output)
}

func seal(output Output) Output {
	output.SelectedIDs, output.UnselectedIDs, output.UnknownIDs = sortedUnique(output.SelectedIDs), sortedUnique(output.UnselectedIDs), sortedUnique(output.UnknownIDs)
	output.OutputDigest = CanonicalOutputDigest(output)
	output.ReplayDigest = ReplayDigest(output.InputDigest, output.OutputDigest)
	return output
}

func effectiveK(value uint64) uint64 {
	if value < LanguageFloor {
		return LanguageFloor
	}
	return value
}
func effectiveMinimum(value uint64) uint64 {
	if value < LanguageFloor {
		return LanguageFloor
	}
	return value
}

func cost(input Input, selected int) (work, cpu, memory, records, paths uint64) {
	work = uint64(len(input.PressureRecords)) + 2*uint64(len(input.RequiredPressureIDs)) + uint64(len(input.GuardIDs)) + uint64(len(input.FinitePathIDs))
	cpu = uint64(selected)
	memory = 1024 * uint64(len(input.RequiredPressureIDs))
	records = uint64(len(input.PressureRecords)) + uint64(len(input.GuardIDs))
	paths = uint64(len(input.FinitePathIDs))
	return
}

func ceilingReason(output Output, ceilings ResourceCeilings) Reason {
	if output.CPUCoreNS > ceilings.CPUCoreNS {
		return ReasonCPUCeilingExceeded
	}
	if output.MemoryBytes > ceilings.MemoryBytes {
		return ReasonMemoryCeilingExceeded
	}
	if output.DeterministicWorkUnits > ceilings.WorkUnits {
		return ReasonWorkCeilingExceeded
	}
	if output.ProvRecords > ceilings.ProvRecords {
		return ReasonProvRecordCeilingExceeded
	}
	if output.ProvPaths > ceilings.ProvPaths {
		return ReasonProvPathCeilingExceeded
	}
	return ""
}

func missingCeiling(value ResourceCeilings) bool {
	return value.CPUCoreNS == 0 || value.MemoryBytes == 0 || value.WorkUnits == 0 || value.ProvRecords == 0 || value.ProvPaths == 0
}

func hasDuplicate(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}
