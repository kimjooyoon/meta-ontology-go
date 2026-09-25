package coupling

import (
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func normalizeSurface(raw Surface) (Surface, *evaluationIssue) {
	surface := raw
	var issue *evaluationIssue
	if surface.SurfaceID, issue = normalizeID(raw.SurfaceID, "surface ID"); issue != nil {
		return Surface{}, issue
	}
	if surface.CodeSymbolID, issue = normalizeID(raw.CodeSymbolID, "code symbol ID"); issue != nil {
		return Surface{}, issue
	}
	if surface.SemanticOwnerID, issue = normalizeID(raw.SemanticOwnerID, "semantic owner ID"); issue != nil {
		return Surface{}, issue
	}
	if surface.Binding.SourceMapID, issue = normalizeID(raw.Binding.SourceMapID, "source map ID"); issue != nil {
		return Surface{}, issue
	}
	if issue = normalizeDigestValue(surface.Binding.BindingDigest, "source map binding digest"); issue != nil {
		return Surface{}, issue
	}
	if expected := bindingDigest(surface); expected != surface.Binding.BindingDigest {
		return Surface{}, failIssue(ReasonSourceMapMismatch, surface.SurfaceID.String())
	}
	return surface, nil
}
func normalizeManifest(
	manifest ChangeManifest, config Config, registry map[semantic.ID]Surface,
) ([]ManifestEntry, *evaluationIssue) {
	if issue := validateManifestHeader(manifest, config); issue != nil {
		return nil, issue
	}
	entries, issue := normalizeManifestEntries(manifest.Entries, registry)
	if issue != nil {
		return nil, issue
	}
	changed := changedManifestEntries(entries)
	if manifest.ZeroChange != (len(changed) == 0) {
		return nil, failIssue(ReasonContradictoryReceipt, "zero-change claim")
	}
	return changed, nil
}
