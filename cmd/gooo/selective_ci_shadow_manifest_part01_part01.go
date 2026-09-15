package main

import (
	"errors"
	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"sort"
	"strings"
)

func plannerManifestFromAnalyzerSnapshot(snapshot analyzersci.Snapshot) (plannersci.SnapshotManifest, error) {
	files := make([]plannersci.SnapshotFile, 0, len(snapshot.Sources))
	for _, source := range snapshot.Sources {
		ids := make([]string, 0, len(source.Bindings))
		seen := map[string]struct{}{}
		for _, binding := range source.Bindings {
			if binding.Status != analyzersci.StatusBound || binding.ID == "" {
				return plannersci.SnapshotManifest{}, errors.New("source binding is not BOUND")
			}
			if _, exists := seen[binding.ID]; exists {
				return plannersci.SnapshotManifest{}, errors.New("duplicate source binding")
			}
			seen[binding.ID] = struct{}{}
			ids = append(ids, binding.ID)
		}
		sort.Strings(ids)
		files = append(files, plannersci.SnapshotFile{Path: source.Path, BlobDigest: rawDigest(source.BlobDigest), SemanticIDs: ids})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	manifest := plannersci.SnapshotManifest{SchemaVersion: plannersci.ManifestSchemaVersion, Files: files}
	manifest.Digest = manifest.ComputedDigest()
	if err := manifest.Validate(); err != nil {
		return plannersci.SnapshotManifest{}, err
	}
	return manifest, nil
}

// rawDigest bridges the analyzer's labelled SHA-256 spelling to the raw-hex
// spelling used by the planner, proof, and lane contracts. It does not hash or
// otherwise alter the digest identity.
func rawDigest(value string) string {
	return strings.TrimPrefix(value, "sha256:")
}
