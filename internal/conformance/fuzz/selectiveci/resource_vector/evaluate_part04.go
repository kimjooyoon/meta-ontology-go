package resourcevector

import (
	"math"
	"sort"
)

func compareCeilings(vector Vector, ceiling CeilingSet, prefix string) []string {
	failures := make([]string, 0, 11)
	checks := []struct {
		name string
		got  uint64
		max  uint64
	}{
		{"cpu_core_ns", vector.CPUCoreNS, *ceiling.CPUCoreNS},
		{"memory_bytes", vector.MemoryBytes, *ceiling.MemoryBytes},
		{"peak_memory_bytes", vector.PeakMemoryBytes, *ceiling.PeakMemoryBytes},
		{"work_units", vector.WorkUnits, *ceiling.WorkUnits},
		{"affected_stable_ids", vector.AffectedStableIDs, *ceiling.AffectedStableIDs},
		{"applicable_pressures", vector.ApplicablePressures, *ceiling.ApplicablePressures},
		{"independent_groups", vector.IndependentGroups, *ceiling.IndependentGroups},
		{"unique_prov_records", vector.UniquePROVRecords, *ceiling.UniquePROVRecords},
		{"finite_prov_paths", vector.FinitePROVPaths, *ceiling.FinitePROVPaths},
		{"closure_numerator", vector.ClosureNumerator, *ceiling.ClosureNumerator},
		{"closure_denominator", vector.ClosureDenominator, *ceiling.ClosureDenominator},
	}
	for _, check := range checks {
		if check.got > check.max {
			failures = append(failures, prefix+":"+check.name)
		}
	}
	sort.Strings(failures)
	return failures
}
func add(left, right uint64) (uint64, bool) {
	if math.MaxUint64-left < right {
		return 0, false
	}
	return left + right, true
}
