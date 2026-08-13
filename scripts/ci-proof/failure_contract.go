package main

import (
	"fmt"
	"sort"
)

type failureCatalogEntry struct {
	Class           string
	Severity        string
	BlockingScope   string
	Parallelizable  bool
	HandoffRequired bool
	Owner           string
}

type failureCatalogRecord struct {
	Code  string
	Entry failureCatalogEntry
}

var failureCatalogRecords = []failureCatalogRecord{
	{Code: "CI-TEST-001", Entry: failureCatalogEntry{Class: "test", Severity: "error", BlockingScope: "local", Parallelizable: true, HandoffRequired: false, Owner: "registered-path-owner"}},
	{Code: "CI-SCOPE-001", Entry: failureCatalogEntry{Class: "scope", Severity: "error", BlockingScope: "global", Parallelizable: false, HandoffRequired: false, Owner: "ci-policy"}},
	{Code: "CI-CAPS-001", Entry: failureCatalogEntry{Class: "caps", Severity: "error", BlockingScope: "global", Parallelizable: false, HandoffRequired: false, Owner: "ci-policy"}},
	{Code: "CI-CONTRACT-001", Entry: failureCatalogEntry{Class: "contract", Severity: "critical", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-DEPENDENCY-001", Entry: failureCatalogEntry{Class: "dependency", Severity: "warning", BlockingScope: "local", Parallelizable: true, HandoffRequired: true, Owner: "registered-path-owner"}},
	{Code: "CI-GATE-001", Entry: failureCatalogEntry{Class: "gate", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-ARTIFACT-001", Entry: failureCatalogEntry{Class: "artifact", Severity: "error", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-FRESHNESS-001", Entry: failureCatalogEntry{Class: "freshness", Severity: "error", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-PROVENANCE-001", Entry: failureCatalogEntry{Class: "provenance", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-OWNERSHIP-001", Entry: failureCatalogEntry{Class: "ownership", Severity: "blocked", BlockingScope: "local", Parallelizable: true, HandoffRequired: true, Owner: "branch-ownership"}},
	{Code: "CI-ROOT-OF-TRUST-001", Entry: failureCatalogEntry{Class: "trust-root", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-ROOT-OF-TRUST-BOOTSTRAP-001", Entry: failureCatalogEntry{Class: "trust-root", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
	{Code: "CI-UNCLASSIFIED-001", Entry: failureCatalogEntry{Class: "unclassified", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
}

var failureCatalog = buildFailureCatalog()

func buildFailureCatalog() map[string]failureCatalogEntry {
	catalog := make(map[string]failureCatalogEntry, len(failureCatalogRecords))
	for _, record := range failureCatalogRecords {
		catalog[record.Code] = record.Entry
	}
	return catalog
}

func validateFailureCodes(codes []string, primary string) error {
	if len(codes) == 0 {
		return fmt.Errorf("failure code set is empty")
	}
	seen := make(map[string]bool, len(codes))
	for index, code := range codes {
		if _, ok := failureCatalog[code]; !ok || seen[code] || code == "" || (index > 0 && codes[index-1] >= code) {
			return fmt.Errorf("failure code set is unknown, duplicated, or not canonical")
		}
		seen[code] = true
	}
	if !seen[primary] {
		return fmt.Errorf("primary failure code is absent from the complete failure set")
	}
	return nil
}

func validateFailureEvidence(manifest failureManifest) error {
	if manifest.ArtifactStatus != "verified" && manifest.ArtifactStatus != "missing" && manifest.ArtifactStatus != "not_applicable" {
		return fmt.Errorf("failure artifact status is missing or unknown")
	}
	if manifest.ArtifactReason == "" || containsUnknown(manifest.ArtifactReason) {
		return fmt.Errorf("failure artifact reason is missing or unknown")
	}
	if manifest.ArtifactStatus == "missing" {
		if manifest.Code != "CI-ARTIFACT-001" || len(manifest.Artifacts) != 0 {
			return fmt.Errorf("missing failure artifact evidence is not fail-closed")
		}
	}
	if manifest.ArtifactStatus == "not_applicable" && len(manifest.Artifacts) != 0 {
		return fmt.Errorf("non-applicable failure artifact evidence must be empty")
	}
	if manifest.ArtifactStatus == "verified" && len(manifest.Artifacts) == 0 {
		return fmt.Errorf("verified failure artifact evidence is empty")
	}
	if manifest.ArtifactStatus == "verified" && (len(manifest.Artifacts) != 1 || manifest.Artifacts[0].Name != fmt.Sprintf("ci-evidence-%d-%d", manifest.RunID, manifest.RunAttempt)) {
		return fmt.Errorf("verified failure artifact evidence is not the exact CI evidence artifact")
	}
	if manifest.Job.Name == "CI proof bundle" && manifest.ProofArtifactRef == nil && !isFailClosedMissingProof(manifest) {
		return fmt.Errorf("proof failure is missing its direct proof artifact reference")
	}
	if manifest.ProofArtifactRef != nil {
		proof := manifest.ProofArtifactRef
		if proof.ID <= 0 || proof.Name != fmt.Sprintf("ci-proof-%d-%d", manifest.RunID, manifest.RunAttempt) || proof.Size <= 0 || proof.Expired || !validArtifactDigest(proof.Digest) || proof.RunID != manifest.RunID || proof.RunAttempt != manifest.RunAttempt {
			return fmt.Errorf("proof artifact reference is stale, zero, or invalid")
		}
	}
	seenArtifacts := make(map[int64]bool, len(manifest.Artifacts))
	for _, artifact := range manifest.Artifacts {
		if artifact.ID <= 0 || seenArtifacts[artifact.ID] || artifact.Name == "" || containsUnknown(artifact.Name) || artifact.Size <= 0 || artifact.Expired || !validArtifactDigest(artifact.Digest) || artifact.RunID != manifest.RunID || artifact.RunAttempt != manifest.RunAttempt {
			return fmt.Errorf("failure artifact evidence is stale, duplicated, or invalid")
		}
		seenArtifacts[artifact.ID] = true
	}
	if manifest.Code == "CI-ARTIFACT-001" && manifest.ArtifactStatus != "missing" && len(manifest.Artifacts) == 0 {
		return fmt.Errorf("artifact failure has no specific artifact evidence")
	}
	for index, rejection := range manifest.Rejections {
		if rejection == "" || containsUnknown(rejection) {
			return fmt.Errorf("failure rejection set contains an unknown value")
		}
		if index > 0 && manifest.Rejections[index-1] == rejection {
			return fmt.Errorf("failure rejection set contains a duplicate")
		}
	}
	if !sort.StringsAreSorted(manifest.Rejections) {
		return fmt.Errorf("failure rejection set is not canonical")
	}
	for _, reason := range []string{manifest.MissingReasons.Protection, manifest.MissingReasons.Provenance} {
		if containsUnknown(reason) {
			return fmt.Errorf("failure missing reason is unknown")
		}
	}
	return nil
}

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

func isCanonicalFailureJob(name string) bool {
	for _, canonical := range []string{"gofmt", "go vet", "go test", "go test -race", "Semantic conformance", "CI policy"} {
		if name == canonical {
			return true
		}
	}
	return false
}

func containsCode(codes []string, target string) bool {
	for _, code := range codes {
		if code == target {
			return true
		}
	}
	return false
}

func sameArtifactInputs(left, right []artifactInput) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
