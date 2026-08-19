package shadow

import (
	"strings"
)

func validFiles(values []manifestFile) bool {
	seenPaths := map[string]struct{}{}
	seenIDs := map[string]struct{}{}
	for _, value := range values {
		if strings.TrimSpace(value.Path) == "" || strings.TrimSpace(value.BlobDigest) == "" || value.SemanticIDs == nil || len(value.SemanticIDs) == 0 {
			return false
		}
		if _, exists := seenPaths[value.Path]; exists {
			return false
		}
		seenPaths[value.Path] = struct{}{}
		for _, id := range value.SemanticIDs {
			if strings.TrimSpace(id) == "" {
				return false
			}
			if _, exists := seenIDs[id]; exists {
				return false
			}
			seenIDs[id] = struct{}{}
		}
	}
	return true
}
func validLaneFacts(lane laneInput) bool {
	return validNonEmpty(lane.RegistryDigest, lane.BaseSHA, lane.LaneHeadSHA, lane.LaneID, lane.RegisteredBranch) &&
		lane.OwnedPathPrefixes != nil && lane.ChangedPaths != nil && lane.AheadCount >= 0 && lane.BehindCount >= 0 && lane.OpenPRCount >= 0 && lane.ActiveLeaseCount >= 0
}
