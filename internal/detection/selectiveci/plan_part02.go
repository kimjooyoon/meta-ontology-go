package selectiveci

import (
	"sort"
)

func changedSemanticIDs(base, head SnapshotManifest) ([]string, error) {
	baseFiles, headFiles := manifestFiles(base), manifestFiles(head)
	ids := map[string]struct{}{}
	paths := map[string]struct{}{}
	for path := range baseFiles {
		paths[path] = struct{}{}
	}
	for path := range headFiles {
		paths[path] = struct{}{}
	}
	for path := range paths {
		before, beforeOK := baseFiles[path]
		after, afterOK := headFiles[path]
		if beforeOK && afterOK && before.BlobDigest == after.BlobDigest && equalStrings(before.SemanticIDs, after.SemanticIDs) {
			continue
		}
		for _, id := range append(append([]string{}, before.SemanticIDs...), after.SemanticIDs...) {
			ids[id] = struct{}{}
		}
		if len(before.SemanticIDs) == 0 && len(after.SemanticIDs) == 0 {
			return nil, failure(ReasonUnknownPath, "changed path has no stable semantic ID")
		}
	}
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	sort.Strings(result)
	return result, nil
}
func manifestFiles(manifest SnapshotManifest) map[string]SnapshotFile {
	result := make(map[string]SnapshotFile, len(manifest.Files))
	for _, file := range manifest.Files {
		result[file.Path] = file
	}
	return result
}
func equalStrings(left, right []string) bool {
	left, right = sortedCopy(left), sortedCopy(right)
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}
