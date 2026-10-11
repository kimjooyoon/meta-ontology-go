package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestFailureManifestConsumesSemanticEvidenceProofChain(t *testing.T) {
	path := "testdata/failure-summary-chain.json"
	if root := os.Getenv("GOOO_FAILURE_TEST_DRAFT"); root != "" {
		path = filepath.Join(root, "scripts/ci-proof", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var input failureInput
	if err := json.Unmarshal(data, &input); err != nil {
		t.Fatal(err)
	}
	manifest, err := buildFailureManifest(input, validFailureBinding())
	if err != nil {
		t.Fatal(err)
	}
	if err := validateFailureManifest(manifest, validFailureBinding()); err != nil {
		t.Fatal(err)
	}
	if manifest.Job.Name != "Semantic conformance" || manifest.ArtifactStatus != "missing" || manifest.ProofArtifactRef != nil {
		t.Fatalf("failure origin or missing proof changed: %+v", manifest)
	}
	if !slices.Equal(manifest.TerminalFailureCodes, []string{"CI-TEST-001", "CI-ARTIFACT-001", "CI-ARTIFACT-001"}) || manifest.NextOperation != "REBUILD_ARTIFACT" {
		t.Fatalf("failure chain was not retained: %+v", manifest)
	}
}

func TestTerminalFailureMappingsMatchProducerCodes(t *testing.T) {
	allowed := map[string][]string{
		"gofmt":                {"CI-TEST-001"},
		"go vet":               {"CI-TEST-001"},
		"go test":              {"CI-TEST-001"},
		"go test -race":        {"CI-TEST-001"},
		"Semantic conformance": {"CI-TEST-001"},
		"CI policy":            {"CI-SCOPE-001", "CI-CAPS-001", "CI-CONTRACT-001"},
		"CI evidence":          {"CI-TEST-001", "CI-ARTIFACT-001"},
		"future check":         {"CI-UNCLASSIFIED-001"},
	}
	for name, codes := range allowed {
		for _, catalog := range failureCatalogRecords {
			t.Run(name+"/"+catalog.Code, func(t *testing.T) {
				job := validFailureInput().Job
				job.Name = name
				manifest := failureManifest{TerminalFailures: []failureJob{job}, TerminalFailureCodes: []string{catalog.Code}, FailureCodes: []string{catalog.Code}}
				err := validateTerminalFailureMapping(manifest, validFailureBinding())
				if (err == nil) != slices.Contains(codes, catalog.Code) {
					t.Fatalf("mapping %s/%s: %v", name, catalog.Code, err)
				}
			})
		}
	}
}

func TestFailureManifestPolicyContractFailureRemainsUnresolved(t *testing.T) {
	input := validFailureInput()
	input.Job.Name = "CI policy"
	input.Code = "CI-CONTRACT-001"
	input.FailureCodes = []string{input.Code}
	input.TerminalFailures = []failureJob{input.Job}
	input.TerminalFailureCodes = []string{input.Code}
	manifest, err := buildFailureManifest(input, validFailureBinding())
	if err != nil {
		t.Fatal(err)
	}
	if manifest.NextOperation != "RETAIN_UNRESOLVED" || manifest.Severity != "critical" {
		t.Fatalf("contract failure was converted to success: %+v", manifest)
	}
}
