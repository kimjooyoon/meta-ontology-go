package shadow

import (
	"sort"
)

func normalizedSelection(planner plannerInput) ([]string, []string, []string, map[string][]string) {
	selected := sortedCopy(planner.SelectedCommandIDs)
	guards := sortedCopy(planner.SelectedGuardCommandIDs)
	work := sortedCopy(planner.SelectedWorkIDs)
	commands := map[string][]string{}
	for _, item := range append(append([]command{}, planner.Commands...), planner.GuardCommands...) {
		commands[item.ID] = append([]string(nil), item.Argv...)
	}
	argv := map[string][]string{}
	for _, id := range append(append([]string{}, selected...), guards...) {
		argv[id] = append([]string(nil), commands[id]...)
	}
	return selected, guards, work, argv
}
func changedRoots(base, head analyzerSnapshot) []string {
	left, right := fileIndex(base.Files), fileIndex(head.Files)
	paths := map[string]struct{}{}
	for path := range left {
		paths[path] = struct{}{}
	}
	for path := range right {
		paths[path] = struct{}{}
	}
	ids := map[string]struct{}{}
	for path := range paths {
		before, beforeOK := left[path]
		after, afterOK := right[path]
		if beforeOK && afterOK && before.BlobDigest == after.BlobDigest && equalStrings(before.SemanticIDs, after.SemanticIDs) {
			continue
		}
		if beforeOK {
			for _, id := range before.SemanticIDs {
				ids[id] = struct{}{}
			}
		}
		if afterOK {
			for _, id := range after.SemanticIDs {
				ids[id] = struct{}{}
			}
		}
	}
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}
func derivedManifest(snapshot analyzerSnapshot) plannerManifest {
	files := normalizeFiles(snapshot.Files)
	result := plannerManifest{Schema: ManifestSchema, Files: files}
	result.Digest = manifestDigest(result)
	return result
}
func manifestEqual(left, right plannerManifest) bool {
	return left.Schema == right.Schema && left.Digest == right.Digest && equalFiles(left.Files, right.Files)
}
func normalizeManifest(value plannerManifest) plannerManifest {
	value.Files = normalizeFiles(value.Files)
	return value
}
