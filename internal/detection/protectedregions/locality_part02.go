package protectedregions

import (
	"bytes"
	"sort"
)

func generatedRegions(report Report) []Region {
	regions := make([]Region, 0)
	for _, region := range report.Regions {
		if region.Kind == Generated {
			regions = append(regions, region)
		}
	}
	sort.SliceStable(regions, func(i, j int) bool { return regions[i].Start < regions[j].Start })
	return regions
}
func compareProtectedBodies(before, after []byte, beforeReport, afterReport Report) []LocalityIssue {
	beforeRegions := protectedRegions(beforeReport)
	afterRegions := protectedRegions(afterReport)
	keys := make([]string, 0, len(beforeRegions)+len(afterRegions))
	for key := range beforeRegions {
		keys = append(keys, key)
	}
	for key := range afterRegions {
		if _, exists := beforeRegions[key]; !exists {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	violations := make([]LocalityIssue, 0)
	for _, key := range keys {
		beforeRegion, beforeExists := beforeRegions[key]
		afterRegion, afterExists := afterRegions[key]
		if !beforeExists {
			violations = append(violations, LocalityIssue{
				Kind: LocalityProtectedChange, Marker: afterRegion.Kind, ID: afterRegion.ID,
				Detail: "protected region was added",
			})
			continue
		}
		if !afterExists {
			violations = append(violations, LocalityIssue{
				Kind: LocalityProtectedChange, Marker: beforeRegion.Kind, ID: beforeRegion.ID,
				Detail: "protected region was removed",
			})
			continue
		}
		if !bytes.Equal(beforeRegion.Body(before), afterRegion.Body(after)) {
			violations = append(violations, LocalityIssue{
				Kind: LocalityProtectedChange, Marker: beforeRegion.Kind, ID: beforeRegion.ID,
				Detail: "protected body changed",
			})
		}
	}
	return violations
}
func protectedRegions(report Report) map[string]Region {
	result := make(map[string]Region)
	for _, region := range report.Regions {
		if region.Kind != Slot && region.Kind != Handwritten {
			continue
		}
		result[string(region.Kind)+"\x00"+region.ID] = region
	}
	return result
}
