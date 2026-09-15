package main

import (
	"fmt"
	"slices"
	"sort"
)

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
	if slices.ContainsFunc([]string{manifest.MissingReasons.Protection, manifest.MissingReasons.Provenance}, containsUnknown) {
		return fmt.Errorf("failure missing reason is unknown")
	}
	return nil
}
