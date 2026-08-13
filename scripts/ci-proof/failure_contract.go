package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
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
	{Code: "CI-UNCLASSIFIED-001", Entry: failureCatalogEntry{Class: "unclassified", Severity: "blocked", BlockingScope: "global", Parallelizable: false, HandoffRequired: true, Owner: "gate"}},
}

var failureCatalog = buildFailureCatalog()

var failureCatalogDigest = immutableFailureCatalogDigest()

func buildFailureCatalog() map[string]failureCatalogEntry {
	catalog := make(map[string]failureCatalogEntry, len(failureCatalogRecords))
	for _, record := range failureCatalogRecords {
		catalog[record.Code] = record.Entry
	}
	return catalog
}

func immutableFailureCatalogDigest() string {
	var payload strings.Builder
	payload.WriteString(failureCatalogPath)
	payload.WriteByte('\n')
	for _, record := range failureCatalogRecords {
		fmt.Fprintf(&payload, "%s|%s|%s|%s|%t|%t|%s\n", record.Code, record.Entry.Class, record.Entry.Severity, record.Entry.BlockingScope, record.Entry.Parallelizable, record.Entry.HandoffRequired, record.Entry.Owner)
	}
	digest := sha256.Sum256([]byte(payload.String()))
	return "sha256:" + hex.EncodeToString(digest[:])
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
	for _, reason := range []string{manifest.MissingReasons.Protection, manifest.MissingReasons.Approval, manifest.MissingReasons.Provenance} {
		if containsUnknown(reason) {
			return fmt.Errorf("failure missing reason is unknown")
		}
	}
	return nil
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
