package selectiveci

import (
	"sort"
	"strings"
)

func bindingIndex(snapshot Snapshot) map[string]indexedBinding {
	result := make(map[string]indexedBinding)
	for _, source := range snapshot.Sources {
		for _, binding := range source.Bindings {
			result[binding.ID] = indexedBinding{
				Path: source.Path, BlobDigest: source.BlobDigest, ID: binding.ID,
				Role: binding.Role, Status: binding.Status, BindingDigest: binding.BindingDigest,
			}
		}
	}
	return result
}
func unionBindingIDs(left, right map[string]indexedBinding) []string {
	set := make(map[string]struct{}, len(left)+len(right))
	for id := range left {
		set[id] = struct{}{}
	}
	for id := range right {
		set[id] = struct{}{}
	}
	ids := make([]string, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
func compareBindings(left, right Binding) int {
	if value := strings.Compare(left.ID, right.ID); value != 0 {
		return value
	}
	if value := strings.Compare(string(left.Role), string(right.Role)); value != 0 {
		return value
	}
	return strings.Compare(left.BindingDigest, right.BindingDigest)
}
func unknownDelta(err error) (Delta, error) {
	return Delta{Status: StatusUnknown, FullSuiteFallback: true}, err
}
