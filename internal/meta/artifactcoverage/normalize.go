package artifactcoverage

import (
	"slices"
	"sort"
)

func normalizeObservations(document *ObservationDocument) {
	for index := range document.Artifacts {
		keys := document.Artifacts[index].EvidenceKeys
		sort.Strings(keys)
		document.Artifacts[index].EvidenceKeys = slices.Compact(keys)
	}
	sort.Slice(document.Artifacts, func(i, j int) bool {
		left, right := document.Artifacts[i], document.Artifacts[j]
		if left.Name != right.Name {
			return left.Name < right.Name
		}
		if left.HeadSHA != right.HeadSHA {
			return left.HeadSHA < right.HeadSHA
		}
		if left.Digest != right.Digest {
			return left.Digest < right.Digest
		}
		return left.ReplayDigest < right.ReplayDigest
	})
}
