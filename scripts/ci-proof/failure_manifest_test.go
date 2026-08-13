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
	if manifest.ProofArtifactRef == nil || manifest.ProofArtifactRef.Name != "ci-proof-9-2" || len(manifest.ArtifactURLs) != 2 {
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

func TestFailureManifestMapsDAMPDRYTerminalFailure(t *testing.T) {
	input := validFailureInput()
	input.Code = "CI-CAPS-001"
	input.FailureCodes = []string{input.Code}
	input.Job = failureJob{ID: 11, Name: "CI policy", Status: "completed", Conclusion: "failure", HeadSHA: strings.Repeat("a", 40), RunID: 9, RunAttempt: 2}
	input.TerminalFailures = []failureJob{input.Job}
	input.TerminalFailureCodes = []string{input.Code}
	manifest, err := buildFailureManifest(input, validFailureBinding())
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Code != "CI-CAPS-001" {
		t.Fatalf("unexpected cap failure code: %s", manifest.Code)
	}
}

func TestFailureManifestRejectsUnorderedTerminalFailures(t *testing.T) {
	input := validFailureInput()
	input.TerminalFailures = append(input.TerminalFailures, failureJob{ID: 12, Name: "go vet", Status: "completed", Conclusion: "failure", HeadSHA: strings.Repeat("a", 40), RunID: 9, RunAttempt: 2})
	input.TerminalFailureCodes = append(input.TerminalFailureCodes, "CI-TEST-001")
	manifest, err := buildFailureManifest(input, validFailureBinding())
	if err != nil {
		t.Fatal(err)
	}
	manifest.TerminalFailures[0], manifest.TerminalFailures[1] = manifest.TerminalFailures[1], manifest.TerminalFailures[0]
	if err := validateFailureManifest(manifest, validFailureBinding()); err == nil {
		t.Fatal("unordered terminal failures were accepted")
	}
}

func validProofFailureInput() failureInput {
	input := validFailureInput()
	input.Code = "CI-GATE-001"
	input.FailureCodes = []string{"CI-GATE-001", "CI-PROVENANCE-001"}
	input.Rejections = []string{"branch_protection_missing", "provenance_evidence_not_verified"}
	input.MissingReasons = missingReasons{Protection: "protection observer not provisioned", Provenance: "provenance observer not provisioned"}
	input.ArtifactStatus = "verified"
	input.ArtifactReason = "proof_artifact_current_and_bound"
	input.Artifacts = []artifactInput{{ID: 12, Name: "ci-evidence-9-2", Size: 1, Digest: "sha256:" + strings.Repeat("a", 64), RunID: 9, RunAttempt: 2}}
	input.ProofArtifact = &artifactInput{ID: 13, Name: "ci-proof-9-2", Size: 1, Digest: "sha256:" + strings.Repeat("b", 64), RunID: 9, RunAttempt: 2}
	input.Job = failureJob{ID: 11, Name: "CI proof bundle", Status: "completed", Conclusion: "failure", HeadSHA: strings.Repeat("a", 40), RunID: 9, RunAttempt: 2}
	input.TerminalFailures = []failureJob{input.Job}
	input.TerminalFailureCodes = []string{"CI-GATE-001"}
	return input
}
