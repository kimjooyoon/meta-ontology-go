package protectedregions

import (
	"bytes"
	"fmt"
	"sort"
)

// Check is the fail-fast structural validation entry point.
func Check(source []byte) error { return Validate(source).Err() }

// ValidateLocality checks that a refreshed projection only changed generated
// region bodies. It rejects changes to slot and handwritten bodies, changes to
// unmarked text, and changes to the set or boundaries of generated regions.
func ValidateLocality(before, after []byte) LocalityReport {
	result := LocalityReport{Before: Validate(before), After: Validate(after)}
	if !result.Before.Valid() || !result.After.Valid() {
		return result
	}
	if !bytes.Equal(generatedSkeleton(before, result.Before), generatedSkeleton(after, result.After)) {
		result.Violations = append(result.Violations, LocalityIssue{
			Kind:   LocalityUnownedChange,
			Detail: "text outside generated region bodies changed",
		})
	}
	result.Violations = append(result.Violations, compareProtectedBodies(before, after, result.Before, result.After)...)
	return result
}

// CheckLocality is the fail-fast locality validation entry point.
func CheckLocality(before, after []byte) error {
	return ValidateLocality(before, after).Err()
}

func generatedSkeleton(source []byte, report Report) []byte {
	regions := generatedRegions(report)
	if len(regions) == 0 {
		return append([]byte(nil), source...)
	}
	result := make([]byte, 0, len(source))
	cursor := 0
	for _, region := range regions {
		result = append(result, source[cursor:region.Start]...)
		result = append(result, []byte(fmt.Sprintf("\x00gooo:generated:%s:%s\x00", region.Kind, region.ID))...)
		cursor = region.End
	}
	return append(result, source[cursor:]...)
}

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
