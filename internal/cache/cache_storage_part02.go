package cache

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"
)

func makeMetadata(key Key, data []byte, info EntryInfo) Metadata {
	metadata := Metadata{
		FormatVersion:         metadataVersion,
		Key:                   key.String(),
		KeyVersion:            key.Version,
		Domain:                key.Domain,
		Namespace:             key.Namespace,
		ArtifactKind:          key.ArtifactKind,
		ToolVersion:           key.ToolVersion,
		Toolchain:             key.Toolchain,
		Target:                key.Target,
		HostStage:             key.HostStage,
		InputDigest:           key.InputDigest,
		SemanticClosureDigest: key.SemanticClosureDigest,
		DependencyRoot:        key.DependencyRoot,
		PolicySchemaDigest:    key.PolicySchemaDigest,
		BuildTagsDigest:       key.BuildTagsDigest,
		OptionsDigest:         key.OptionsDigest,
		DependencyDigest:      key.DependencyDigest,
		ProvenanceDigest:      key.ProvenanceDigest,
		ArtifactType:          info.ArtifactType,
		Projection:            key.Projection,
		Reconstructable:       true,
		Size:                  int64(len(data)),
		ContentDigest:         HashBytes(data),
		CreatedAt:             time.Now().UTC(),
	}
	metadata.MetadataDigest = digestMetadata(metadata)
	return metadata
}
func digestMetadata(metadata Metadata) Digest {
	metadata.MetadataDigest = ""
	data, err := CanonicalBytes(metadata)
	if err != nil {
		return ""
	}
	return HashBytes(data)
}
func writeObjectFiles(directory string, data []byte, metadata Metadata) error {
	if err := writeDurableFile(filepath.Join(directory, dataFileName), data); err != nil {
		return fmt.Errorf("cache: write projection: %w", err)
	}
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("cache: encode metadata: %w", err)
	}
	if err := writeDurableFile(filepath.Join(directory, metaFileName), metadataBytes); err != nil {
		return fmt.Errorf("cache: write metadata: %w", err)
	}
	return nil
}
