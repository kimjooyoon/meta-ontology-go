package coupling

import (
	"sort"
)

func SortReasons(result *Result) {
	if result == nil {
		return
	}
	result.Reasons = sortedReasons(result.Reasons)
	sort.Slice(result.AcceptedSurfaceIDs, func(i, j int) bool { return result.AcceptedSurfaceIDs[i] < result.AcceptedSurfaceIDs[j] })
	result.Digest = stableDigest(resultCanonical(*result))
}
