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
	regions := skeletonRegions(report)
	if len(regions) == 0 {
		return append([]byte(nil), source...)
	}
	result := make([]byte, 0, len(source))
	cursor := 0
	for _, region := range regions {
		result = append(result, source[cursor:region.BodyStart]...)
		result = append(result, []byte(fmt.Sprintf("\x00gooo:protected-body:%s:%s\x00", region.Kind, region.ID))...)
		result = append(result, source[region.BodyEnd:region.End]...)
		cursor = region.End
	}
	return append(result, source[cursor:]...)
}
func skeletonRegions(report Report) []Region {
	generated := generatedRegions(report)
	regions := append([]Region(nil), generated...)
	for _, region := range report.Regions {
		if region.Kind != Slot && region.Kind != Handwritten {
			continue
		}
		insideGenerated := false
		for _, owner := range generated {
			if region.Start >= owner.Start && region.End <= owner.End {
				insideGenerated = true
				break
			}
		}
		if !insideGenerated {
			regions = append(regions, region)
		}
	}
	sort.SliceStable(regions, func(i, j int) bool { return regions[i].Start < regions[j].Start })
	return regions
}
