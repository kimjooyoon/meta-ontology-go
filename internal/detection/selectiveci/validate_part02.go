package selectiveci

import (
	"encoding/json"
)

func (manifest SnapshotManifest) Validate() error {
	if manifest.SchemaVersion != ManifestSchemaVersion {
		return failure(ReasonUnsupportedSchema, "unsupported manifest schema_version")
	}
	if manifest.Files == nil {
		return failure(ReasonInvalidInput, "manifest files are required")
	}
	seen := map[string]struct{}{}
	for _, file := range manifest.Files {
		if !validRepoPath(file.Path) {
			return failure(ReasonUnknownPath, "manifest path is not normalized")
		}
		if _, exists := seen[file.Path]; exists {
			return failure(ReasonDuplicateID, "duplicate manifest path")
		}
		seen[file.Path] = struct{}{}
		if !validDigest(file.BlobDigest) {
			return failure(ReasonMismatchedDigest, "blob_digest is not SHA-256")
		}
		if err := validateIDs(file.SemanticIDs); err != nil {
			return err
		}
		if len(file.SemanticIDs) != len(sortedUnique(file.SemanticIDs)) {
			return failure(ReasonDuplicateID, "duplicate semantic ID in manifest file")
		}
	}
	if manifest.Digest != manifest.ComputedDigest() {
		return failure(ReasonMismatchedDigest, "snapshot_digest does not match manifest")
	}
	return nil
}
func (manifest SnapshotManifest) ComputedDigest() string {
	canonical, err := json.Marshal(normalizeManifest(manifest))
	if err != nil {
		return ""
	}
	var value struct {
		SchemaVersion string         `json:"schema_version"`
		Files         []SnapshotFile `json:"files"`
	}
	value.SchemaVersion, value.Files = manifest.SchemaVersion, normalizeManifest(manifest).Files
	canonical, err = json.Marshal(value)
	if err != nil {
		return ""
	}
	return digestBytes(canonical)
}
