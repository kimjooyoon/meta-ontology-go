package pressureindependence

import (
	"encoding/hex"
	"sort"
	"strings"
	"unicode"
)

type validation struct {
	decision Decision
	reason   Reason
	records  map[string]PressureRecord
}

func Evaluate(input Input) Output {
	output := Output{Schema: SchemaV1, FixtureID: input.FixtureID, InputDigest: CanonicalInputDigest(input)}
	check := validate(input)
	if check.decision != "" {
		return finishUnknown(output, input, check.decision, check.reason)
	}
	groups := make(map[string][]string)
	for _, id := range input.RequiredPressureIDs {
		groups[check.records[id].IndependenceGroupID] = append(groups[check.records[id].IndependenceGroupID], id)
	}
	output.DistinctGroupCount = uint64(len(groups))
	if uint64(len(groups)) < input.MinimumIndependent ||
		uint64(len(groups)) < effectiveK(input.RequestedK) {
		return finishUnknown(output, input, DecisionUnknown, ReasonIndependentGroupShortfall)
	}
	selected := selectRepresentatives(groups, input.RequestedK)
	output.SelectedIDs = selected
	output.UnselectedIDs = subtract(input.RequiredPressureIDs, selected)
	output.CostReceipt = receipt(input, len(selected))
	if !withinCeilings(output.CostReceipt, input.ResourceCeilings) {
		return finishUnknown(output, input, DecisionUnknown, ReasonInvalidResourceReceipt)
	}
	output.Decision = DecisionPass
	output.Reason = ReasonNone
	output.FullSuiteRequired = false
	output.ProofValid = finiteProof(input, output)
	return seal(output)
}

func validate(input Input) validation {
	if missingInput(input) {
		if len(input.FinitePathIDs) == 0 && len(input.GuardIDs) > 0 {
			return validation{decision: DecisionUnknown, reason: ReasonProvPathMissing}
		}
		return validation{decision: DecisionUnknown, reason: ReasonRequiredInputMissing}
	}
	if staleDigest(input) {
		return validation{decision: DecisionUnknown, reason: ReasonStaleDigest}
	}
	if hasDuplicate(input.RequiredPressureIDs) {
		return validation{decision: DecisionFailClosed, reason: ReasonDuplicatePressureID}
	}
	if malformedLists(input) {
		return validation{decision: DecisionFailClosed, reason: ReasonProvPathMalformed}
	}
	records := make(map[string]PressureRecord, len(input.PressureRecords))
	for _, record := range input.PressureRecords {
		if _, exists := records[record.PressureID]; exists {
			if records[record.PressureID] == record {
				return validation{decision: DecisionFailClosed, reason: ReasonDuplicatePressureID}
			}
			return validation{decision: DecisionFailClosed, reason: ReasonConflictingGroupBinding}
		}
		records[record.PressureID] = record
	}
	applicabilityRule := ""
	for _, id := range input.RequiredPressureIDs {
		if _, exists := records[id]; !exists {
			return validation{decision: DecisionFailClosed, reason: ReasonCatalogMismatch}
		}
		record := records[id]
		if record.IndependenceGroupID == "" || record.ApplicabilityRuleID == "" {
			return validation{decision: DecisionUnknown, reason: ReasonApplicabilityUnproven}
		}
		if applicabilityRule == "" {
			applicabilityRule = record.ApplicabilityRuleID
		} else if applicabilityRule != record.ApplicabilityRuleID {
			return validation{decision: DecisionUnknown, reason: ReasonInputAmbiguous}
		}
	}
	return validation{records: records}
}

func missingInput(input Input) bool {
	values := []string{input.Schema, input.FixtureID, input.AuthoritySnapshotDigest, input.PolicyDigest,
		input.RegistryDigest, input.OracleDigest, input.ToolchainOptionsDigest}
	for _, value := range values {
		if !validToken(value) {
			return true
		}
	}
	return input.Schema != SchemaV1 || input.RequestedK == 0 || input.MinimumIndependent == 0 ||
		len(input.PressureRecords) == 0 || len(input.RequiredPressureIDs) == 0 ||
		len(input.GuardIDs) == 0 || len(input.FinitePathIDs) == 0 || missingCeiling(input.ResourceCeilings)
}

func staleDigest(input Input) bool {
	digests := []string{input.AuthoritySnapshotDigest, input.PolicyDigest, input.RegistryDigest,
		input.OracleDigest, input.ToolchainOptionsDigest}
	for _, value := range digests {
		if !validDigest(value) || allZeroDigest(value) {
			return true
		}
	}
	return false
}

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

func seal(output Output) Output {
	output.SelectedIDs = sortedUnique(output.SelectedIDs)
	output.UnselectedIDs = sortedUnique(output.UnselectedIDs)
	output.UnknownIDs = sortedUnique(output.UnknownIDs)
	output.CanonicalOutputDigest = CanonicalOutputDigest(output)
	output.ReplayDigest = ReplayDigest(output.InputDigest, output.CanonicalOutputDigest)
	return output
}

func effectiveK(requested uint64) uint64 {
	if requested < 2 {
		return 2
	}
	return requested
}

func finiteProof(input Input, output Output) bool {
	return output.Decision == DecisionPass && len(input.FinitePathIDs) >= len(input.RequiredPressureIDs) &&
		len(input.GuardIDs) > 0 && len(input.FinitePathIDs) > 0
}

func receipt(input Input, selected int) CostReceipt {
	return CostReceipt{
		CPUCoreNS: uint64(selected), MemoryBytes: uint64(len(input.RequiredPressureIDs)) * 1024,
		WorkUnits: uint64(len(input.PressureRecords) + len(input.RequiredPressureIDs) +
			len(input.GuardIDs) + len(input.FinitePathIDs)),
		ProvRecords: uint64(len(input.PressureRecords) + len(input.GuardIDs)),
		ProvPaths:   uint64(len(input.FinitePathIDs)),
	}
}

func withinCeilings(receipt CostReceipt, ceilings ResourceCeilings) bool {
	return receipt.CPUCoreNS <= ceilings.CPUCoreNS && receipt.MemoryBytes <= ceilings.MemoryBytes &&
		receipt.WorkUnits <= ceilings.WorkUnits && receipt.ProvRecords <= ceilings.ProvRecords &&
		receipt.ProvPaths <= ceilings.ProvPaths
}

func missingCeiling(ceilings ResourceCeilings) bool {
	return ceilings.CPUCoreNS == 0 || ceilings.MemoryBytes == 0 || ceilings.WorkUnits == 0 ||
		ceilings.ProvRecords == 0 || ceilings.ProvPaths == 0
}

func validToken(value string) bool {
	if value == "" || value != strings.TrimSpace(value) || len(value) > 256 {
		return false
	}
	for _, r := range value {
		if unicode.IsSpace(r) || unicode.IsControl(r) {
			return false
		}
	}
	return true
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(value[len("sha256:"):])
	return err == nil
}

func allZeroDigest(value string) bool {
	for _, r := range value[len("sha256:"):] {
		if r != '0' {
			return false
		}
	}
	return true
}

func hasDuplicate(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}
