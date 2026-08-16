package pressureindependence

import (
	"encoding/hex"
	"sort"
	"strings"
)

// EvaluateBaseline is a deliberately separate typed-config/full-suite
// comparison. It observes every declared pressure and never reuses the
// pressure-independence evaluator's selection helpers.
func EvaluateBaseline(input Input) BaselineResult {
	result := BaselineResult{FullSuite: true, LocalizedIDs: []string{}}
	result.WorkUnits = baselineWork(input)
	result.CostReceipt = baselineReceipt(input, result.WorkUnits)
	if baselineMissing(input) {
		return baselineUnknown(result, input, ReasonRequiredInputMissing)
	}
	if baselineStale(input) {
		return baselineUnknown(result, input, ReasonStaleDigest)
	}
	records := make(map[string]PressureRecord, len(input.PressureRecords))
	for _, record := range input.PressureRecords {
		if prior, exists := records[record.PressureID]; exists {
			if prior == record {
				return baselineFail(result, input, ReasonDuplicatePressureID)
			}
			return baselineFail(result, input, ReasonConflictingGroupBinding)
		}
		records[record.PressureID] = record
	}
	groups := make(map[string]struct{})
	applicabilityRule := ""
	for _, id := range input.RequiredPressureIDs {
		record, exists := records[id]
		if !exists {
			return baselineFail(result, input, ReasonCatalogMismatch)
		}
		if record.IndependenceGroupID == "" || record.ApplicabilityRuleID == "" {
			return baselineUnknown(result, input, ReasonApplicabilityUnproven)
		}
		if applicabilityRule == "" {
			applicabilityRule = record.ApplicabilityRuleID
		} else if applicabilityRule != record.ApplicabilityRuleID {
			return baselineUnknown(result, input, ReasonInputAmbiguous)
		}
		groups[record.IndependenceGroupID] = struct{}{}
	}
	if uint64(len(groups)) < input.MinimumIndependent ||
		uint64(len(groups)) < baselineK(input.RequestedK) {
		return baselineUnknown(result, input, ReasonIndependentGroupShortfall)
	}
	result.Decision, result.Reason = DecisionPass, ReasonNone
	return result
}

func Compare(input Input) Comparison {
	oracle := Evaluate(input)
	baseline := EvaluateBaseline(input)
	localized := oracle.UnknownIDs
	if oracle.Decision == DecisionPass {
		localized = nil
	}
	comparison := Comparison{
		Oracle: oracle, Baseline: baseline, ResearchWorkUnits: oracle.CostReceipt.WorkUnits,
		BaselineWorkUnits: baseline.WorkUnits,
		OutcomeMatch:      oracle.Decision == baseline.Decision,
		ReasonMatch:       oracle.Reason == baseline.Reason,
		LocalizationMatch: equalStrings(localized, baseline.LocalizedIDs),
	}
	comparison.ResearchBudgetOK = withinResearchBudget(comparison.ResearchWorkUnits, comparison.BaselineWorkUnits)
	if comparison.OutcomeMatch && comparison.ReasonMatch && comparison.LocalizationMatch && comparison.ResearchBudgetOK {
		comparison.Finding = NoUniqueBenefit
	} else {
		comparison.Finding = UniqueBenefitNotEstablished
	}
	return comparison
}

func baselineMissing(input Input) bool {
	values := []string{input.Schema, input.FixtureID, input.AuthoritySnapshotDigest, input.PolicyDigest,
		input.RegistryDigest, input.OracleDigest, input.ToolchainOptionsDigest}
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			return true
		}
	}
	return input.Schema != SchemaV1 || input.RequestedK == 0 || input.MinimumIndependent == 0 ||
		len(input.PressureRecords) == 0 || len(input.RequiredPressureIDs) == 0 || len(input.GuardIDs) == 0 ||
		len(input.FinitePathIDs) == 0 || input.ResourceCeilings == (ResourceCeilings{})
}

func baselineStale(input Input) bool {
	for _, value := range []string{input.AuthoritySnapshotDigest, input.PolicyDigest, input.RegistryDigest,
		input.OracleDigest, input.ToolchainOptionsDigest} {
		if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") || zeroHex(value[7:]) {
			return true
		}
	}
	return false
}

func zeroHex(value string) bool {
	if _, err := hex.DecodeString(value); err != nil {
		return true
	}
	for _, character := range value {
		if character != '0' {
			return false
		}
	}
	return true
}

func baselineUnknown(result BaselineResult, input Input, reason Reason) BaselineResult {
	result.Decision, result.Reason = DecisionUnknown, reason
	result.LocalizedIDs = sortedUnique(input.RequiredPressureIDs)
	return result
}

func baselineFail(result BaselineResult, input Input, reason Reason) BaselineResult {
	result.Decision, result.Reason = DecisionFailClosed, reason
	result.LocalizedIDs = sortedUnique(input.RequiredPressureIDs)
	return result
}

func baselineK(requested uint64) uint64 {
	if requested < 2 {
		return 2
	}
	return requested
}

func baselineWork(input Input) uint64 {
	return uint64(len(input.PressureRecords) + 2*len(input.RequiredPressureIDs) +
		len(input.GuardIDs) + len(input.FinitePathIDs))
}

func baselineReceipt(input Input, work uint64) CostReceipt {
	return CostReceipt{
		CPUCoreNS: uint64(len(input.RequiredPressureIDs)), MemoryBytes: uint64(len(input.RequiredPressureIDs)) * 1024,
		WorkUnits:   work,
		ProvRecords: uint64(len(input.PressureRecords) + len(input.RequiredPressureIDs) + len(input.GuardIDs)),
		ProvPaths:   uint64(len(input.FinitePathIDs)),
	}
}

func withinResearchBudget(research, baseline uint64) bool {
	if baseline == 0 {
		return false
	}
	if research <= baseline {
		return true
	}
	return research <= baseline+(baseline+3)/4
}

func equalStrings(left, right []string) bool {
	left, right = append([]string(nil), left...), append([]string(nil), right...)
	sort.Strings(left)
	sort.Strings(right)
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
