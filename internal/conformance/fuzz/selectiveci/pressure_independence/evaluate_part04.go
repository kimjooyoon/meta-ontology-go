package pressureindependence

import (
	"encoding/hex"
	"strings"
	"unicode"
)

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
