package pressureindependence

import (
	"encoding/hex"
	"strings"
)

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
