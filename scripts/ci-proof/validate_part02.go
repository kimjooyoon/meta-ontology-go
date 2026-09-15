package main

import (
	"encoding/json"
	"fmt"
)

func validateProofDigest(bundle proofBundle) error {
	recorded := bundle.Digests.Bundle
	authorizationDigest := ""
	if bundle.PromotionAuthorization != nil {
		authorizationDigest = bundle.PromotionAuthorization.ProofDigest
	}
	bundle.Digests.Bundle = ""
	if bundle.PromotionAuthorization != nil {
		authorization := *bundle.PromotionAuthorization
		authorization.ProofDigest = ""
		bundle.PromotionAuthorization = &authorization
	}
	payload, err := json.Marshal(bundle)
	if err != nil {
		return err
	}
	if digestBytes(payload) != recorded {
		return fmt.Errorf("proof bundle digest mismatch")
	}
	if bundle.PromotionAuthorization != nil && authorizationDigest != recorded {
		return fmt.Errorf("promotion authorization is not signed by the proof digest")
	}
	return nil
}
func validatePredecessors(bundle proofBundle) error {
	seen := make(map[string]bool, len(bundle.Predecessors))
	for _, predecessor := range bundle.Predecessors {
		if !validDigest(predecessor) || predecessor == bundle.HeadSHA || seen[predecessor] {
			return fmt.Errorf("stale or cyclic predecessor")
		}
		seen[predecessor] = true
	}
	return nil
}
func validateArtifacts(artifacts []artifactInput, runID, runAttempt int64) error {
	if len(artifacts) != 1 {
		return fmt.Errorf("artifact inventory must contain exactly one current CI evidence artifact")
	}
	expectedName := fmt.Sprintf("ci-evidence-%d-%d", runID, runAttempt)
	seenIDs := make(map[int64]bool, len(artifacts))
	for _, artifact := range artifacts {
		if artifact.ID <= 0 || seenIDs[artifact.ID] || artifact.Name != expectedName || artifact.Size <= 0 || artifact.Expired || !validArtifactDigest(artifact.Digest) || artifact.RunID != runID || artifact.RunAttempt != runAttempt {
			return fmt.Errorf("artifact inventory contains missing, zero, or expired artifact")
		}
		seenIDs[artifact.ID] = true
	}
	return nil
}
func validateMissingReasons(reasons missingReasons, protectionStatus, provenanceStatus string) error {
	for _, item := range []struct {
		label  string
		status string
		reason string
	}{{"protection", protectionStatus, reasons.Protection}, {"provenance", provenanceStatus, reasons.Provenance}} {
		label, status, reason := item.label, item.status, item.reason
		if status == "unavailable" || status == "missing" {
			if reason == "" {
				return fmt.Errorf("%s evidence is missing without a reason", label)
			}
			continue
		}
		if reason != "" {
			return fmt.Errorf("%s evidence has a reason despite status %q", label, status)
		}
	}
	return nil
}
