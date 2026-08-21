package resourcevector

import (
	"strings"
	"unicode"
)

func validateCeilings(ceilings ResourceCeilings) validationFailure {
	if !ceilingComplete(ceilings.Selected) || !ceilingComplete(ceilings.Full) {
		return validationFailure{DecisionUnknown, ReasonMissingInput}
	}
	return validationFailure{}
}
func ceilingComplete(ceiling CeilingSet) bool {
	return ceiling.CPUCoreNS != nil && ceiling.MemoryBytes != nil && ceiling.PeakMemoryBytes != nil &&
		ceiling.WorkUnits != nil && ceiling.AffectedStableIDs != nil && ceiling.ApplicablePressures != nil &&
		ceiling.IndependentGroups != nil && ceiling.UniquePROVRecords != nil && ceiling.FinitePROVPaths != nil &&
		ceiling.ClosureNumerator != nil && ceiling.ClosureDenominator != nil
}
func validRoot(root string) bool {
	return root != "" && root == strings.TrimSpace(root) && !strings.ContainsAny(root, "\x00")
}
func validToken(value string) bool {
	if value == "" || value != strings.TrimSpace(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}
