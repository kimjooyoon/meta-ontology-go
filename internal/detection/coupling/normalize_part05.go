package coupling

import (
	"slices"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"sort"
)

func normalizeManifestEntries(raw []ManifestEntry, registry map[semantic.ID]Surface) ([]ManifestEntry, *evaluationIssue) {
	if len(raw) != len(registry) {
		return nil, failIssue(ReasonMalformedBinding, "manifest entry cardinality")
	}
	entries := append([]ManifestEntry(nil), raw...)
	sort.Slice(entries, func(i, j int) bool { return entries[i].SurfaceID < entries[j].SurfaceID })
	seen := make(map[semantic.ID]struct{}, len(entries))
	for index := range entries {
		entry := &entries[index]
		if _, duplicate := seen[entry.SurfaceID]; duplicate {
			return nil, failIssue(ReasonDuplicateSurface, entry.SurfaceID.String())
		}
		seen[entry.SurfaceID] = struct{}{}
		surface, exists := registry[entry.SurfaceID]
		if !exists {
			return nil, failIssue(ReasonSurfaceUnregistered, entry.SurfaceID.String())
		}
		if entry.CodeSymbolID != surface.CodeSymbolID || entry.SemanticOwnerID != surface.SemanticOwnerID {
			return nil, failIssue(ReasonSourceMapMismatch, entry.SurfaceID.String())
		}
		if entry.AfterBindingDigest != surface.Binding.BindingDigest {
			return nil, failIssue(ReasonSourceMapMismatch, entry.SurfaceID.String())
		}
		if entry.BeforeSourcePath == "" || entry.AfterSourcePath == "" {
			return nil, required("manifest source path")
		}
		for _, value := range []struct {
			value string
			name  string
		}{
			{entry.BeforeBindingDigest, "before binding digest"},
			{entry.AfterBindingDigest, "after binding digest"},
			{entry.BeforeBlobDigest, "before blob digest"},
			{entry.AfterBlobDigest, "after blob digest"},
		} {
			if issue := normalizeDigestValue(value.value, value.name); issue != nil {
				return nil, issue
			}
		}
	}
	return entries, nil
}
func changedManifestEntries(entries []ManifestEntry) []ManifestEntry {
	changed := make([]ManifestEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.BeforeBindingDigest != entry.AfterBindingDigest || entry.BeforeBlobDigest != entry.AfterBlobDigest {
			changed = append(changed, entry)
		}
	}
	return changed
}
func sortedIDs(values []semantic.ID) []semantic.ID {
	result := append([]semantic.ID(nil), values...)
	slices.Sort(result)
	return result
}
func sortedReasons(values []Reason) []Reason {
	result := append([]Reason(nil), values...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Code != result[j].Code {
			return result[i].Code < result[j].Code
		}
		return result[i].Detail < result[j].Detail
	})
	return result
}
