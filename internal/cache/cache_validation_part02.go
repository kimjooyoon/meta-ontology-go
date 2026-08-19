package cache

import (
	"fmt"
)

func validateMetadataForKey(metadata Metadata, key Key) error {
	if !metadataSane(metadata) {
		return fmt.Errorf("metadata envelope is invalid")
	}
	if metadata.Key != key.String() {
		return fmt.Errorf("metadata key mismatch")
	}
	if key.Version != "" && metadata.KeyVersion != key.Version {
		return fmt.Errorf("key version mismatch")
	}
	if key.Namespace != "" && metadata.Namespace != key.Namespace {
		return fmt.Errorf("namespace mismatch")
	}
	if key.Domain != "" && metadata.Domain != key.Domain {
		return fmt.Errorf("domain mismatch")
	}
	if key.ArtifactKind != "" && metadata.ArtifactKind != key.ArtifactKind {
		return fmt.Errorf("artifact kind mismatch")
	}
	if key.Projection != "" && metadata.Projection != key.Projection {
		return fmt.Errorf("projection mismatch")
	}
	if key.ToolVersion != "" && metadata.ToolVersion != key.ToolVersion {
		return fmt.Errorf("tool version mismatch")
	}
	if key.Toolchain != "" && metadata.Toolchain != key.Toolchain {
		return fmt.Errorf("toolchain mismatch")
	}
	if key.Target != "" && metadata.Target != key.Target {
		return fmt.Errorf("target mismatch")
	}
	if key.HostStage != "" && metadata.HostStage != key.HostStage {
		return fmt.Errorf("host stage mismatch")
	}
	if key.InputDigest.Valid() && metadata.InputDigest != key.InputDigest {
		return fmt.Errorf("input digest mismatch")
	}
	if key.SemanticClosureDigest.Known() && metadata.SemanticClosureDigest != key.SemanticClosureDigest {
		return fmt.Errorf("semantic closure digest mismatch")
	}
	if key.DependencyRoot.Known() && metadata.DependencyRoot != key.DependencyRoot {
		return fmt.Errorf("dependency root mismatch")
	}
	if key.PolicySchemaDigest.Known() && metadata.PolicySchemaDigest != key.PolicySchemaDigest {
		return fmt.Errorf("policy schema digest mismatch")
	}
	if key.BuildTagsDigest.Known() && metadata.BuildTagsDigest != key.BuildTagsDigest {
		return fmt.Errorf("build tags digest mismatch")
	}
	if key.OptionsDigest.Valid() && metadata.OptionsDigest != key.OptionsDigest {
		return fmt.Errorf("options digest mismatch")
	}
	if key.DependencyDigest.Valid() && metadata.DependencyDigest != key.DependencyDigest {
		return fmt.Errorf("dependency digest mismatch")
	}
	if key.ProvenanceDigest.Valid() && metadata.ProvenanceDigest != key.ProvenanceDigest {
		return fmt.Errorf("provenance digest mismatch")
	}
	return nil
}
