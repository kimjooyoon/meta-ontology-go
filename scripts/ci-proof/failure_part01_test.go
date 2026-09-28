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
	if manifest.Schema != failureSchema || manifest.Version != 2 || manifest.Code != "CI-TEST-001" || manifest.Scope != "pr" || manifest.BlockingScope != "local" || !manifest.Parallelizable || manifest.CatalogPath != failureCatalogPath || manifest.CatalogDigest != failureCatalogDigest || manifest.HeadBranch != "agent/ci-workflow" || manifest.NextOperation != "REPAIR_SOURCE" || len(manifest.ArtifactRefs) != 0 {
		t.Fatalf("unexpected failure manifest classification: %+v", manifest)
	}
	encoded, err := json.Marshal(manifest)
	if err != nil || !strings.Contains(string(encoded), `"catalog_path":"`+failureCatalogPath+`"`) {
		t.Fatalf("machine-readable catalog path is missing: %s", encoded)
	}
	if manifest.Provenance.WasGeneratedBy != manifest.Activity || manifest.Provenance.WasAssociatedWith != manifest.Agent || len(manifest.Provenance.WasDerivedFrom) != 2 || len(manifest.Provenance.HadPrimarySource) != 3 || len(manifest.EvidenceRefs) != 4 {
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
func TestFailureManifestBindsCatalogRefAndSystemOperation(t *testing.T) {
	manifest, err := buildFailureManifest(validFailureInput(), validFailureBinding())
	if err != nil {
		t.Fatal(err)
	}
	if manifest.CatalogRef != failureCatalogPath+"@"+validFailureBinding().HeadSHA || manifest.CatalogVersion != 2 || manifest.CatalogSHA256 != failureCatalogDigest || manifest.NextOperation != "REPAIR_SOURCE" {
		t.Fatalf("immutable catalog or system operation binding missing: %+v", manifest)
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "handoff_required") || strings.Contains(string(encoded), "handoff_owner") {
		t.Fatalf("failure manifest retained human handoff fields: %s", encoded)
	}
}
