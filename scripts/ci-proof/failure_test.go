package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
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

func TestFailureManifestRejectsTamperedClassificationAndHandoff(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.Class = "gate"
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("tampered failure classification was accepted")
	}
	manifest.Class = "test"
	manifest.HandoffOwner = "gate"
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("tampered failure handoff owner was accepted")
	}
}

func TestFailureCatalogMatchesCheckedInDocument(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not locate failure catalog test source")
	}
	document, err := os.ReadFile(filepath.Join(filepath.Dir(source), "docs", "failure-reasons.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := "sha256:" + digestBytes(document); got != failureCatalogDigest {
		t.Fatalf("catalog digest is not bound to document bytes: got %s want %s", failureCatalogDigest, got)
	}
	if err := validateFailureCatalog(); err != nil {
		t.Fatal(err)
	}
	counts := make(map[string]int)
	for _, line := range strings.Split(string(document), "\n") {
		if !strings.HasPrefix(line, "| `CI-") {
			continue
		}
		parts := strings.Split(line, "`")
		if len(parts) < 2 {
			t.Fatalf("malformed catalog row: %s", line)
		}
		counts[parts[1]]++
	}
	if len(counts) != len(failureCatalogRecords) {
		t.Fatalf("catalog code count mismatch: docs=%d machine=%d", len(counts), len(failureCatalogRecords))
	}
	for _, record := range failureCatalogRecords {
		if counts[record.Code] != 1 {
			t.Fatalf("catalog code %s is missing or duplicated: %d", record.Code, counts[record.Code])
		}
	}
	for code := range counts {
		if _, ok := failureCatalog[code]; !ok {
			t.Fatalf("catalog contains unknown code %s", code)
		}
	}
}

func TestFailureOwnerRegistryRequiresRegisteredCIOwner(t *testing.T) {
	if err := validateFailureOwnerRegistry("agent/ci-workflow"); err != nil {
		t.Fatal(err)
	}
	if err := validateFailureOwnerRegistry("agent/bidir"); err == nil {
		t.Fatal("non-CI owner was accepted for failure manifests")
	}
}

func TestFailureManifestRejectsTamperedOwnerBinding(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.OwnerBranch = "agent/other"
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("tampered failure owner branch was accepted")
	}
}

func TestFailureManifestRejectsTamperedArtifactReference(t *testing.T) {
	binding := validFailureBinding()
	input := validFailureInput()
	input.ArtifactStatus = "verified"
	input.ArtifactReason = "proof_artifact_current_and_bound"
	input.Artifacts = []artifactInput{{ID: 12, Name: "ci-evidence-9-2", Size: 1, Digest: "sha256:" + strings.Repeat("a", 64), RunID: 9, RunAttempt: 2}}
	manifest, err := buildFailureManifest(input, binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.ArtifactRefs[0].ID++
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("tampered failure artifact reference was accepted")
	}
}

func TestFailureManifestRejectsDuplicateRejection(t *testing.T) {
	binding := validFailureBinding()
	input := validFailureInput()
	input.Rejections = []string{"artifact_inventory_missing", "artifact_inventory_missing"}
	if _, err := buildFailureManifest(input, binding); err == nil {
		t.Fatal("duplicate rejection was accepted")
	}
}

func TestFailureManifestRejectsZeroArtifactEvidence(t *testing.T) {
	binding := validFailureBinding()
	input := validFailureInput()
	input.ArtifactStatus = "verified"
	input.ArtifactReason = "proof artifact was expected"
	if _, err := buildFailureManifest(input, binding); err == nil {
		t.Fatal("zero verified artifacts were accepted")
	}
}

func TestFailureManifestRejectsStaleArtifactEvidence(t *testing.T) {
	binding := validFailureBinding()
	input := validFailureInput()
	input.ArtifactStatus = "verified"
	input.ArtifactReason = "proof_artifact_current_and_bound"
	input.Artifacts = []artifactInput{{ID: 12, Name: "ci-evidence-9-2", Size: 1, Digest: "sha256:" + strings.Repeat("a", 64), RunID: 8, RunAttempt: 2}}
	if _, err := buildFailureManifest(input, binding); err == nil {
		t.Fatal("stale failure artifact evidence was accepted")
	}
}

func TestFailureManifestRejectsCallerSuppliedFailureCodeSet(t *testing.T) {
	binding := validFailureBinding()
	input := validFailureInput()
	input.FailureCodes = []string{"CI-TEST-001", "CI-TEST-001"}
	if _, err := buildFailureManifest(input, binding); err == nil {
		t.Fatal("duplicate caller-supplied failure codes were accepted")
	}
}

func TestFailureManifestAllowsMissingArtifactOnlyFailClosed(t *testing.T) {
	binding := validFailureBinding()
	input := validFailureInput()
	input.Code = "CI-ARTIFACT-001"
	input.FailureCodes = []string{"CI-ARTIFACT-001"}
	input.TerminalFailureCodes = []string{"CI-ARTIFACT-001"}
	input.Job = failureJob{ID: 11, Name: "CI proof bundle", Status: "completed", Conclusion: "failure", HeadSHA: strings.Repeat("a", 40), RunID: 9, RunAttempt: 2}
	input.TerminalFailures = []failureJob{input.Job}
	input.Rejections = []string{"proof_artifact_missing"}
	input.ArtifactStatus = "missing"
	input.ArtifactReason = "artifact_missing"
	input.Message = "proof artifact is missing for the exact run"
	input.Remediation = "rerun the exact head and publish the proof artifact"
	manifest, err := buildFailureManifest(input, binding)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.ArtifactStatus != "missing" || len(manifest.Artifacts) != 0 {
		t.Fatalf("missing artifact was not preserved fail-closed: %+v", manifest)
	}
}

func TestFailureManifestRejectsStaleHead(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.HeadSHA = strings.Repeat("b", 40)
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("stale failure head was accepted")
	}
}

func TestFailureManifestRejectsCheckoutMismatch(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.CheckoutRef = strings.Repeat("b", 40)
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("checkout ref mismatch was accepted")
	}
}

func TestFailureManifestRejectsTamperedProvenance(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.Provenance.WasGeneratedBy = "urn:gooo:ci-run:replayed"
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("tampered provenance relation was accepted")
	}
}

func TestFailureManifestRejectsCallerSuppliedTuple(t *testing.T) {
	binding := validFailureBinding()
	manifest, err := buildFailureManifest(validFailureInput(), binding)
	if err != nil {
		t.Fatal(err)
	}
	manifest.RunID++
	if err := validateFailureManifest(manifest, binding); err == nil {
		t.Fatal("caller-supplied run tuple was accepted")
	}
}

func TestFailureManifestRejectsMismatchedJobBinding(t *testing.T) {
	binding := validFailureBinding()
	input := validFailureInput()
	input.Job.HeadSHA = strings.Repeat("b", 40)
	if _, err := buildFailureManifest(input, binding); err == nil {
		t.Fatal("mismatched failure job head was accepted")
	}
}

func TestFailureManifestRejectsUnknownBinding(t *testing.T) {
	binding := validFailureBinding()
	binding.Actor = "unknown"
	if _, err := buildFailureManifest(validFailureInput(), binding); err == nil {
		t.Fatal("unknown agent binding was accepted")
	}
}
