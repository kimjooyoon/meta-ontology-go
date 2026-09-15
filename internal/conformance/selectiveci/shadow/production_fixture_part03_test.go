package shadow

import (
	"crypto/sha256"
	"encoding/hex"
	analyzersci "github.com/kimjooyoon/meta-ontology-go/internal/analyzer/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/analyzer/semanticbinding"
	"github.com/kimjooyoon/meta-ontology-go/internal/detection/resourceenvelope"
	plannersci "github.com/kimjooyoon/meta-ontology-go/internal/detection/selectiveci"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
	"strings"
	"testing"
)

func buildProductionSnapshot(t *testing.T, source, id, registryDigest string) analyzersci.Snapshot {
	t.Helper()
	result, err := semanticbinding.Extract(semanticbinding.Input{Sources: []semanticbinding.SourceFile{{Filename: "cmd/gooo/fixture.gooo", PackagePath: "fixture", Source: []byte(source)}}})
	if err != nil || result.Status != semanticbinding.StatusBound || len(result.Bindings) != 1 {
		t.Fatalf("semantic binding fixture = %#v, err = %v", result, err)
	}
	snapshot, err := analyzersci.Build(analyzersci.SnapshotInput{
		Sources:         []analyzersci.SourceInput{{Path: "cmd/gooo/fixture.gooo", BlobDigest: productionPrefixedDigest(source), Bindings: result.Bindings}},
		SourceMapDigest: productionPrefixedDigest("source-map"), RegistryDigest: registryDigest, RegisteredIDs: []string{id},
	})
	if err != nil {
		t.Fatalf("analyzer snapshot fixture = %v", err)
	}
	return snapshot
}
func productionManifest(t *testing.T, snapshot analyzersci.Snapshot) plannersci.SnapshotManifest {
	t.Helper()
	files := make([]plannersci.SnapshotFile, 0, len(snapshot.Sources))
	for _, source := range snapshot.Sources {
		ids := make([]string, 0, len(source.Bindings))
		for _, binding := range source.Bindings {
			ids = append(ids, binding.ID)
		}
		files = append(files, plannersci.SnapshotFile{Path: source.Path, BlobDigest: rawProductionDigest(source.BlobDigest), SemanticIDs: ids})
	}
	manifest := plannersci.SnapshotManifest{SchemaVersion: plannersci.ManifestSchemaVersion, Files: files}
	manifest.Digest = manifest.ComputedDigest()
	if err := manifest.Validate(); err != nil {
		t.Fatalf("planner manifest fixture = %v", err)
	}
	return manifest
}
func productionEnvelope() resourceenvelope.Envelope {
	samples := []resourceenvelope.Sample{{CPUCoreNS: 1, WallNS: 10}, {CPUCoreNS: 10, WallNS: 10}, {CPUCoreNS: 20, WallNS: 10}, {CPUCoreNS: 30, WallNS: 10}, {CPUCoreNS: 40, WallNS: 10}, {CPUCoreNS: 50, WallNS: 10}}
	return resourceenvelope.Envelope{SchemaVersion: resourceenvelope.SchemaVersion, RunnerImageDigest: productionDigest("runner"), AllocatedCPUCount: 1, WarmupCount: 1, SampleCount: 5,
		Limits: resourceenvelope.Limits{CPUCoreNS: 100, PeakRSSBytes: 1000, ReadBytes: 1000, WriteBytes: 1000}, Samples: samples}
}
func productionID(value string) semantic.ID { return semantic.MustIdentity(value) }
func productionDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
func productionPrefixedDigest(value string) string { return "sha256:" + productionDigest(value) }
func rawProductionDigest(value string) string      { return strings.TrimPrefix(value, "sha256:") }
