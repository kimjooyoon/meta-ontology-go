package coupling

import (
	"slices"
)

func SortReasons(result *Result) {
	if result == nil {
		return
	}
	result.Reasons = sortedReasons(result.Reasons)
	slices.Sort(result.AcceptedSurfaceIDs)
	result.Digest = stableDigest(resultCanonical(*result))
}
