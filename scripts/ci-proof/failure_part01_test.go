package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFailureManifestBuildsCanonicalPROVRelations(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Schema != failureSchema || manifest.Version != 1 || manifest.Code != "CI-TEST-001" || manifest.Scope != "pr" || manifest.BlockingScope != "local" || !manifest.Parallelizable || manifest.CatalogPath != failureCatalogPath || manifest.CatalogDigest != failureCatalogDigest || manifest.OwnerBranch != "agent/ci-workflow" || manifest.OwnerRef == "" || len(manifest.ArtifactRefs) != 0 {
		t.Fatalf("unexpected failure manifest classification: %+v", manifest)
	}
	encoded, err := json.Marshal(manifest)
	if err != nil || !strings.Contains(string(encoded), `"catalog_path":"`+failureCatalogPath+`"`) {
		t.Fatalf("machine-readable catalog path is missing: %s", encoded)
	}
	if manifest.Provenance.WasGeneratedBy != manifest.Activity || manifest.Provenance.WasAssociatedWith != manifest.Agent || len(manifest.Provenance.WasDerivedFrom) != 2 || len(manifest.Provenance.HadPrimarySource) != 4 || len(manifest.EvidenceRefs) != 5 {
		t.Fatal("PROV relations were not canonicalized")
	}
	if err := validateFailureManifest(manifest, binding); err != nil {
		t.Fatal(err)
	}
}
func TestFailureManifestRejectsTamperedCatalogPath(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.CatalogPath = "scripts/ci-proof/docs/other-reasons.md"
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("tampered failure catalog path was accepted")
	}
}
func TestFailureManifestRejectsTamperedCatalogDigest(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.CatalogDigest = "sha256:" + strings.Repeat("0", 64)
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("tampered failure catalog digest was accepted")
	}
}
func TestFailureManifestBindsCatalogRefAndHandoffParity(t *testing.T) {
	manifest, err := buildFailureManifest(validFailureInput(), validFailureBinding())
	if err != nil {
		t.Fatal(err)
	}
	if manifest.CatalogRef != failureCatalogPath+"@"+validFailureBinding().HeadSHA || manifest.CatalogVersion != 1 || manifest.CatalogSHA256 != failureCatalogDigest || manifest.HandoffOwner != "registered-path-owner" {
		t.Fatalf("immutable catalog or owner binding missing: %+v", manifest)
	}
	if !failureCatalog["CI-ARTIFACT-001"].HandoffRequired || !failureCatalog["CI-FRESHNESS-001"].HandoffRequired {
		t.Fatal("artifact and freshness failures must require handoff")
	}
}
