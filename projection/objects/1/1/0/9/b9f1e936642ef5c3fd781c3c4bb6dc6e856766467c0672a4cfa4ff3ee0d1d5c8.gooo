package main

import (
	"strings"
	"testing"
)

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
func TestFailureManifestRejectsDuplicateTerminalJobName(t *testing.T) {
	input := validFailureInput()
	duplicate := input.Job
	duplicate.ID = 12
	input.TerminalFailures = append(input.TerminalFailures, duplicate)
	input.TerminalFailureCodes = append(input.TerminalFailureCodes, "CI-TEST-001")
	if _, err := buildFailureManifest(input, validFailureBinding()); err == nil {
		t.Fatal("duplicate terminal job name was accepted")
	}
}
func TestFailureManifestRejectsStaleSecondaryTerminalJob(t *testing.T) {
	input := validFailureInput()
	input.TerminalFailures = append(input.TerminalFailures, failureJob{ID: 12, Name: "go vet", Status: "completed", Conclusion: "failure", HeadSHA: strings.Repeat("b", 40), RunID: 9, RunAttempt: 2})
	input.TerminalFailureCodes = append(input.TerminalFailureCodes, "CI-TEST-001")
	if _, err := buildFailureManifest(input, validFailureBinding()); err == nil {
		t.Fatal("stale secondary terminal job was accepted")
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
