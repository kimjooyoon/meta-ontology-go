package main

import (
	"strings"
	"testing"
)

func TestFailureManifestBindsProofArtifactReference(t *testing.T) {
	manifest, err := buildFailureManifest(validProofFailureInput(), validFailureBinding())
	if err != nil {
		t.Fatal(err)
	}
	if manifest.ProofArtifactRef == nil || manifest.ProofArtifactRef.Name != "ci-proof-9-2" || len(manifest.ArtifactURLs) != 2 || len(manifest.ArtifactRefs) != 2 || manifest.ArtifactRefs[1] != *manifest.ProofArtifactRef {
		t.Fatalf("proof artifact reference was not bound: %+v", manifest)
	}
	manifest.ProofArtifactRef.Digest = "sha256:" + strings.Repeat("0", 64)
	if err := validateFailureManifest(manifest, validFailureBinding()); err == nil {
		t.Fatal("tampered proof artifact reference was accepted")
	}
}
func TestFailureManifestAllowsMissingProofArtifactFailClosed(t *testing.T) {
	input := validProofFailureInput()
	input.Code = "CI-ARTIFACT-001"
	input.FailureCodes = []string{input.Code}
	input.Rejections = []string{"proof_artifact_missing"}
	input.ArtifactStatus = "missing"
	input.ArtifactReason = "artifact_missing"
	input.Artifacts = nil
	input.ProofArtifact = nil
	input.TerminalFailureCodes = []string{input.Code}
	manifest, err := buildFailureManifest(input, validFailureBinding())
	if err != nil {
		t.Fatal(err)
	}
	if manifest.ProofArtifactRef != nil || manifest.ArtifactStatus != "missing" || manifest.Code != "CI-ARTIFACT-001" {
		t.Fatalf("missing proof artifact was not preserved as fail-closed: %+v", manifest)
	}
}
func TestFailureManifestRejectsUnclassifiedMissingProofArtifact(t *testing.T) {
	input := validProofFailureInput()
	input.Code = "CI-ARTIFACT-001"
	input.FailureCodes = []string{input.Code}
	input.Rejections = []string{"artifact_evidence_not_verified"}
	input.ArtifactStatus = "missing"
	input.ArtifactReason = "canonical_job_failure"
	input.Artifacts = nil
	input.ProofArtifact = nil
	if _, err := buildFailureManifest(input, validFailureBinding()); err == nil {
		t.Fatal("unclassified missing proof artifact was accepted")
	}
}
func TestFailureManifestMapsUnknownTerminalJobToUnclassified(t *testing.T) {
	input := validFailureInput()
	input.Code = "CI-UNCLASSIFIED-001"
	input.FailureCodes = []string{input.Code}
	input.Job.Name = "future terminal check"
	input.TerminalFailures = []failureJob{input.Job}
	input.TerminalFailureCodes = []string{input.Code}
	manifest, err := buildFailureManifest(input, validFailureBinding())
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Code != input.Code {
		t.Fatalf("unexpected unknown-job code: %s", manifest.Code)
	}
	input.TerminalFailureCodes = []string{"CI-TEST-001"}
	if _, err := buildFailureManifest(input, validFailureBinding()); err == nil {
		t.Fatal("unknown terminal job was accepted without CI-UNCLASSIFIED-001")
	}
}
