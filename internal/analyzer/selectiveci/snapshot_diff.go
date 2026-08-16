package selectiveci

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
)

// Diff computes the union of stable IDs for changed, deleted, added, or
// relocated exact source bindings. It returns no IDs on UNKNOWN.
func Diff(base, head Snapshot) (Delta, error) {
	left, err := normalizeSnapshot(base)
	if err != nil {
		return unknownDelta(err)
	}
	right, err := normalizeSnapshot(head)
	if err != nil {
		return unknownDelta(err)
	}
	leftBindings, rightBindings := bindingIndex(left), bindingIndex(right)
	if left.SourceMapDigest != right.SourceMapDigest || left.RegistryDigest != right.RegistryDigest {
		return Delta{Status: StatusBound, ChangedIDs: unionBindingIDs(leftBindings, rightBindings)}, nil
	}
	ids := unionBindingIDs(leftBindings, rightBindings)
	changed := make([]string, 0, len(ids))
	for _, id := range ids {
		before, beforeOK := leftBindings[id]
		after, afterOK := rightBindings[id]
		if !beforeOK || !afterOK || before != after {
			changed = append(changed, id)
		}
	}
	return Delta{Status: StatusBound, ChangedIDs: changed}, nil
}

// DiffSnapshots is an explicit alias for callers that distinguish manifests.
func DiffSnapshots(base, head Snapshot) (Delta, error) { return Diff(base, head) }

// CanonicalJSON returns the canonical JSON encoding of a delta.
func (d Delta) CanonicalJSON() ([]byte, error) {
	if d.Status != StatusBound && d.Status != StatusUnknown {
		return nil, fail(CodeInvalidStatus, "unknown delta status %q", d.Status)
	}
	ids := append([]string(nil), d.ChangedIDs...)
	sort.Strings(ids)
	if d.Status == StatusUnknown && len(ids) != 0 {
		return nil, fail(CodeInvalidStatus, "UNKNOWN delta cannot contain partial IDs")
	}
	type wireDelta struct {
		Status            Status   `json:"status"`
		ChangedIDs        []string `json:"changed_ids"`
		FullSuiteFallback bool     `json:"full_suite_fallback"`
	}
	return json.Marshal(wireDelta{Status: d.Status, ChangedIDs: ids, FullSuiteFallback: d.FullSuiteFallback})
}

func (d Delta) IsEmpty() bool { return d.Status == StatusBound && len(d.ChangedIDs) == 0 }

type indexedBinding struct {
	Path, BlobDigest, ID string
	Role                 semanticbinding.Role
	Status               Status
	BindingDigest        string
}

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
