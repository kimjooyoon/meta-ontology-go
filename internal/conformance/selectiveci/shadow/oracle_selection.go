package shadow

import (
	"encoding/json"
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

func normalizeFiles(values []manifestFile) []manifestFile {
	result := append([]manifestFile(nil), values...)
	for i := range result {
		result[i].SemanticIDs = sortedCopy(result[i].SemanticIDs)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}

func normalizeCommands(values []command) []command {
	result := append([]command(nil), values...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	for i := range result {
		result[i].Argv = append([]string(nil), result[i].Argv...)
	}
	return result
}

func fileIndex(values []manifestFile) map[string]manifestFile {
	result := map[string]manifestFile{}
	for _, value := range values {
		result[value.Path] = value
	}
	return result
}

func equalFiles(left, right []manifestFile) bool {
	return stringJSON(normalizeFiles(left)) == stringJSON(normalizeFiles(right))
}

func stringJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func equalStrings(left, right []string) bool {
	return stringJSON(sortedCopy(left)) == stringJSON(sortedCopy(right))
}

func sortedCopy(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}
