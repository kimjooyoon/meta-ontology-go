package selectiveci

import (
	"sort"
)

func normalizeManifestBinding(binding Binding, repoPath string, registered map[string]struct{}) (Binding, error) {
	if binding.Status != StatusBound {
		return Binding{}, fail(CodeInvalidStatus, "binding %q is not BOUND", binding.ID)
	}
	id, err := normalizeID(binding.ID)
	if err != nil {
		return Binding{}, err
	}
	if _, ok := registered[id]; !ok {
		return Binding{}, fail(CodeUnregisteredID, "binding ID %q is not in the explicit registry", id)
	}
	if !validRole(binding.Role) {
		return Binding{}, fail(CodeInvalidBinding, "binding %q has invalid role %q", id, binding.Role)
	}
	if !validRawDigest(binding.BindingDigest) {
		return Binding{}, fail(CodeMalformedDigest, "binding %q digest is malformed", id)
	}
	return Binding{ID: id, Role: binding.Role, Status: StatusBound, BindingDigest: binding.BindingDigest}, nil
}
func rejectDuplicateManifestIDs(sources []Source) error {
	seen := make(map[string]string)
	for _, source := range sources {
		for _, binding := range source.Bindings {
			if previous, exists := seen[binding.ID]; exists {
				if previous == source.Path {
					return fail(CodeDuplicateBinding, "binding ID %q is duplicated in source %q", binding.ID, source.Path)
				}
				return fail(CodeAmbiguousBinding, "binding ID %q is bound by %q and %q", binding.ID, previous, source.Path)
			}
			seen[binding.ID] = source.Path
		}
	}
	return nil
}
func normalizeRegisteredIDs(values []string) (map[string]struct{}, error) {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		id, err := normalizeID(value)
		if err != nil {
			return nil, err
		}
		if _, exists := result[id]; exists {
			return nil, fail(CodeDuplicateBinding, "registered ID %q is duplicated", id)
		}
		result[id] = struct{}{}
	}
	return result, nil
}
func sortedIDs(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
