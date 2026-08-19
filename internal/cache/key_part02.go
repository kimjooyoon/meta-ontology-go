package cache

// Key is retained as the cache API name while exposing the ProjectionKey v2
// identity contract.
type Key = ProjectionKey

// NewProjectionKey creates a v2 key from already typed identity digests.
func NewProjectionKey(spec ProjectionKeySpec) (ProjectionKey, error) {
	normalized, err := normalizeProjectionSpec(spec)
	if err != nil {
		return ProjectionKey{}, err
	}
	buildTagsDigest, err := digestBuildTags(normalized.BuildTags)
	if err != nil {
		return ProjectionKey{}, err
	}
	optionsDigest, err := requireOptionsDigest(normalized.OptionsDigest, normalized.Options)
	if err != nil {
		return ProjectionKey{}, err
	}
	key := ProjectionKey{
		Domain: normalized.Domain, Namespace: normalized.Domain, Version: normalized.Version,
		ArtifactKind: normalized.ArtifactKind, Projection: normalized.Projection,
		HostStage: normalized.HostStage, SemanticClosureDigest: normalized.SemanticClosureDigest,
		DependencyRoot: normalized.DependencyRoot, PolicySchemaDigest: normalized.PolicySchemaDigest,
		Toolchain: normalized.Toolchain, ToolVersion: normalized.Toolchain, Target: normalized.Target,
		BuildTagsDigest: buildTagsDigest, OptionsDigest: optionsDigest,
		InputDigest: normalized.SemanticClosureDigest, DependencyDigest: normalized.DependencyRoot,
		ProvenanceDigest: normalized.PolicySchemaDigest,
	}
	key.Digest = digestForProjectionKey(key)
	return key, nil
}
