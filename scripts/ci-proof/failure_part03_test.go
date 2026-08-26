package main

import (
	"strings"
	"testing"
)

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
