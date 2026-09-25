package cache

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func metadataSane(metadata Metadata) bool {
	if metadata.FormatVersion != metadataVersion || metadata.Key == "" || !isDigestName(metadata.Key) {
		return false
	}
	if !metadata.InputDigest.Known() || !metadata.SemanticClosureDigest.Known() ||
		!metadata.DependencyRoot.Known() || !metadata.PolicySchemaDigest.Known() ||
		!metadata.BuildTagsDigest.Known() || !metadata.OptionsDigest.Known() ||
		!metadata.DependencyDigest.Known() || !metadata.ProvenanceDigest.Known() ||
		!metadata.ContentDigest.Valid() {
		return false
	}
	if !metadata.Reconstructable || metadata.Size < 0 || metadata.CreatedAt.IsZero() {
		return false
	}
	if !metadata.HostStage.Valid() || validateKeyComponent("key version", metadata.KeyVersion, true) != nil ||
		validateKeyComponent("domain", metadata.Domain, true) != nil ||
		validateKeyComponent("namespace", metadata.Namespace, true) != nil ||
		validateKeyComponent("artifact kind", metadata.ArtifactKind, true) != nil ||
		validateKeyComponent("tool version", metadata.ToolVersion, false) != nil ||
		validateKeyComponent("toolchain", metadata.Toolchain, true) != nil ||
		validateKeyComponent("target", metadata.Target, true) != nil ||
		validateKeyComponent("artifact type", metadata.ArtifactType, false) != nil ||
		validateKeyComponent("projection", metadata.Projection, false) != nil {
		return false
	}
	if metadata.Domain != metadata.Namespace || metadata.Toolchain != metadata.ToolVersion ||
		metadata.InputDigest != metadata.SemanticClosureDigest ||
		metadata.DependencyDigest != metadata.DependencyRoot ||
		metadata.ProvenanceDigest != metadata.PolicySchemaDigest {
		return false
	}
	key := ProjectionKey{
		Digest: metadataDigestPlaceholder(metadata.Key), Domain: metadata.Domain, Namespace: metadata.Namespace,
		Version: metadata.KeyVersion, ArtifactKind: metadata.ArtifactKind, Projection: metadata.Projection,
		HostStage: metadata.HostStage, SemanticClosureDigest: metadata.SemanticClosureDigest,
		DependencyRoot: metadata.DependencyRoot, PolicySchemaDigest: metadata.PolicySchemaDigest,
		Toolchain: metadata.Toolchain, ToolVersion: metadata.ToolVersion, Target: metadata.Target,
		BuildTagsDigest: metadata.BuildTagsDigest, OptionsDigest: metadata.OptionsDigest,
		InputDigest: metadata.InputDigest, DependencyDigest: metadata.DependencyDigest,
		ProvenanceDigest: metadata.ProvenanceDigest,
	}
	if metadata.Key != digestForProjectionKey(key).String() {
		return false
	}
	return metadata.MetadataDigest.Valid() && digestMetadata(metadata) == metadata.MetadataDigest
}
func metadataDigestPlaceholder(serialized string) Digest { return Digest(serialized) }
func corruptError(key Key, reason string) error {
	return fmt.Errorf("%w: %s: %s", ErrCorrupt, key, reason)
}
func validateKeyComponent(label, value string, required bool) error {
	if required && value == "" {
		return fmt.Errorf("%s must not be empty", label)
	}
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s must be valid UTF-8", label)
	}
	if strings.IndexByte(value, 0) >= 0 {
		return fmt.Errorf("%s must not contain NUL", label)
	}
	return nil
}
