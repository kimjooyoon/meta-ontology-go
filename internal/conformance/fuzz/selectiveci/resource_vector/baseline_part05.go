package resourcevector

import (
	"sort"
)

func strictlyBetter(oracle Vector, baseline *PartialVector) bool {
	values := []struct {
		left  uint64
		right *uint64
	}{
		{oracle.CPUCoreNS, baseline.CPUCoreNS}, {oracle.MemoryBytes, baseline.MemoryBytes},
		{oracle.PeakMemoryBytes, baseline.PeakMemoryBytes}, {oracle.WorkUnits, baseline.WorkUnits},
		{oracle.AffectedStableIDs, baseline.AffectedStableIDs}, {oracle.ApplicablePressures, baseline.ApplicablePressures},
		{oracle.IndependentGroups, baseline.IndependentGroups}, {oracle.UniquePROVRecords, baseline.UniquePROVRecords},
		{oracle.FinitePROVPaths, baseline.FinitePROVPaths}, {oracle.ClosureNumerator, baseline.ClosureNumerator},
		{oracle.ClosureDenominator, baseline.ClosureDenominator},
	}
	strict := false
	for _, value := range values {
		if value.right == nil {
			return false
		}
		if value.left > *value.right {
			return false
		}
		if value.left < *value.right {
			strict = true
		}
	}
	return strict
}
func PartialVectorValues(vector *PartialVector) []uint64 {
	if vector == nil {
		return nil
	}
	values := make([]uint64, 0, 11)
	for _, value := range []*uint64{vector.CPUCoreNS, vector.MemoryBytes, vector.PeakMemoryBytes, vector.WorkUnits, vector.AffectedStableIDs, vector.ApplicablePressures, vector.IndependentGroups, vector.UniquePROVRecords, vector.FinitePROVPaths, vector.ClosureNumerator, vector.ClosureDenominator} {
		if value != nil {
			values = append(values, *value)
		}
	}
	sort.Slice(values, func(left, right int) bool { return values[left] < values[right] })
	return values
}
