package generator

import (
	"bytes"
)

func markerManifestRegionLessV1(left, right markerManifestRegionV1) bool {
	if left.Start != right.Start {
		return left.Start < right.Start
	}
	if left.End != right.End {
		return left.End < right.End
	}
	if left.Kind != right.Kind {
		return left.Kind < right.Kind
	}
	return left.ID < right.ID
}
func markerManifestSlotLessV1(left, right markerManifestSlotV1) bool {
	if left.Start != right.Start {
		return left.Start < right.Start
	}
	if left.End != right.End {
		return left.End < right.End
	}
	return left.ID < right.ID
}
func sameParsedSlotV1(source []byte, observed parsedSlot, expected markerManifestSlotV1) bool {
	return observed.ID == expected.ID && observed.RegionID == expected.RegionID && observed.RegionKind == expected.RegionKind && observed.Start == expected.Start && observed.End == expected.End && observed.StartLine == expected.StartLine && observed.EndLine == expected.EndLine && observed.Start >= 0 && observed.End >= observed.Start && observed.End <= len(source) && bytes.Equal(observed.Body, source[observed.Start:observed.End])
}
