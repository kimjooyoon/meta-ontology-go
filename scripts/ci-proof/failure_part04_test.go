package main

import (
	"strings"
	"testing"
)

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
