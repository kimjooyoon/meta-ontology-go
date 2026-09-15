package selectiveci

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"sort"
)

func normalizeManifestSources(inputs []Source, registered map[string]struct{}) ([]Source, error) {
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
		bindings := make([]Binding, len(input.Bindings))
		copy(bindings, input.Bindings)
		for index := range bindings {
			binding, err := normalizeManifestBinding(bindings[index], repoPath, registered)
			if err != nil {
				return nil, err
			}
			bindings[index] = binding
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
func normalizeBinding(binding semanticbinding.Binding, repoPath string, registered map[string]struct{}) (Binding, error) {
	spanPath, err := normalizeRepoPath(binding.Span.Filename)
	if err != nil {
		return Binding{}, fail(CodeMissingBinding, "binding %q has no valid source span: %v", binding.ID, err)
	}
	if spanPath != repoPath {
		return Binding{}, fail(CodeAmbiguousBinding, "binding %q is attached to %q, not %q", binding.ID, spanPath, repoPath)
	}
	if !validRole(binding.Role) {
		return Binding{}, fail(CodeInvalidBinding, "binding %q has invalid role %q", binding.ID, binding.Role)
	}
	id, err := normalizeID(binding.ID)
	if err != nil {
		return Binding{}, err
	}
	if _, ok := registered[id]; !ok {
		return Binding{}, fail(CodeUnregisteredID, "binding ID %q is not in the explicit registry", id)
	}
	expected := binding.StableHash()
	if !validRawDigest(binding.Digest) || binding.Digest != expected || binding.CanonicalDigest != expected {
		return Binding{}, fail(CodeStaleSnapshot, "binding %q digest does not match its explicit semantic record", id)
	}
	return Binding{ID: id, Role: binding.Role, Status: StatusBound, BindingDigest: expected}, nil
}
