package main

import (
	"fmt"
)

func isFailClosedMissingProof(manifest failureManifest) bool {
	if manifest.Code != "CI-ARTIFACT-001" || manifest.ArtifactStatus != "missing" {
		return false
	}
	if manifest.ArtifactReason != "artifact_missing" && manifest.ArtifactReason != "artifact_invalid" && manifest.ArtifactReason != "artifact_missing_or_invalid" {
		return false
	}
	for _, rejection := range manifest.Rejections {
		if rejection == "proof_artifact_missing" || rejection == "proof_artifact_invalid" {
			return true
		}
	}
	return false
}
func sameFailureJobs(jobs []failureJob, primary failureJob, codes []string) bool {
	if len(jobs) == 0 || len(jobs) != len(codes) || !sameFailureJob(jobs[0], primary) {
		return false
	}
	seenIDs := make(map[int64]bool, len(jobs))
	seenNames := make(map[string]bool, len(jobs))
	for index, job := range jobs {
		if index > 0 && (job.ID < jobs[index-1].ID || (job.ID == jobs[index-1].ID && job.Name < jobs[index-1].Name)) {
			return false
		}
		if job.ID <= 0 || seenIDs[job.ID] || seenNames[job.Name] || codes[index] == "" {
			return false
		}
		seenIDs[job.ID] = true
		seenNames[job.Name] = true
	}
	return true
}
func sameFailureJob(left, right failureJob) bool {
	return left == right
}
func validateTerminalFailureMapping(manifest failureManifest, binding failureBinding) error {
	if len(manifest.TerminalFailures) != len(manifest.TerminalFailureCodes) {
		return fmt.Errorf("terminal failure mapping is incomplete")
	}
	for index, job := range manifest.TerminalFailures {
		if err := validateFailureJob(job, binding); err != nil {
			return fmt.Errorf("terminal failure job is stale or mismatched: %w", err)
		}
		code := manifest.TerminalFailureCodes[index]
		if _, ok := failureCatalog[code]; !ok || !containsCode(manifest.FailureCodes, code) {
			return fmt.Errorf("terminal failure code is unknown or missing from the complete set")
		}
		if !isFailureJobName(job.Name) && code != "CI-UNCLASSIFIED-001" {
			return fmt.Errorf("unmapped terminal job lacks CI-UNCLASSIFIED-001")
		}
		if job.Name == "CI policy" && code != "CI-SCOPE-001" && code != "CI-CAPS-001" {
			return fmt.Errorf("CI policy terminal failure has a non-deterministic code")
		}
		if isCanonicalFailureJob(job.Name) && job.Name != "CI policy" && code != "CI-TEST-001" {
			return fmt.Errorf("canonical terminal job %q has a non-deterministic code", job.Name)
		}
		if job.Name == "CI proof bundle" && code == "CI-UNCLASSIFIED-001" {
			return fmt.Errorf("proof terminal job has an unclassified code")
		}
	}
	return nil
}
