package selectiveci

import (
	"sort"
)

func normalizeSources(inputs []SourceInput, registered map[string]struct{}) ([]Source, error) {
	sources := make([]Source, 0, len(inputs))
	seenPaths := make(map[string]struct{}, len(inputs))
	for _, input := range inputs {
		repoPath, err := normalizeRepoPath(input.Path)
		if err != nil {
			return nil, err
		}
		if _, exists := seenPaths[repoPath]; exists {
			return nil, fail(CodeDuplicateBinding, "source path %q is duplicated", repoPath)
		}
		seenPaths[repoPath] = struct{}{}
		blobDigest, err := normalizeDigest(input.BlobDigest, "blob digest")
		if err != nil {
			return nil, err
		}
		if input.Bindings == nil {
			return nil, fail(CodeMissingBinding, "source %q has no explicit semantic binding", repoPath)
		}
		bindings := make([]Binding, 0, len(input.Bindings))
		for _, binding := range input.Bindings {
			record, err := normalizeBinding(binding, repoPath, registered)
			if err != nil {
				return nil, err
			}
			bindings = append(bindings, record)
		}
		sort.Slice(bindings, func(i, j int) bool { return compareBindings(bindings[i], bindings[j]) < 0 })
		sources = append(sources, Source{Path: repoPath, BlobDigest: blobDigest, Bindings: bindings})
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })
	if err := rejectDuplicateManifestIDs(sources); err != nil {
		return nil, err
	}
	return sources, nil
}
