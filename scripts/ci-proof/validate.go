package main

import (
	"encoding/json"
	"fmt"
)

func validateProof(bundle proofBundle) error {
	if bundle.Schema != proofSchema || bundle.Repository == "" || bundle.Event == "" || bundle.Ref == "" || bundle.EventRef == "" || bundle.CheckoutRef == "" || bundle.EventRef != bundle.Ref || bundle.CheckoutRef != bundle.HeadSHA || bundle.BaseRef == "" || bundle.RunID <= 0 || bundle.RunAttempt <= 0 || bundle.PRNumber < 0 || bundle.WriteEffect != "none" || !bundle.NoWrite {
		return fmt.Errorf("proof metadata is incomplete")
	}
	if !validSHA(bundle.BaseSHA) || !validSHA(bundle.HeadSHA) || !validSHA(bundle.WorkflowSHA) || bundle.BaseSHA == bundle.HeadSHA {
		return fmt.Errorf("proof revisions are invalid or identical")
	}
	if len(bundle.Jobs) != len(proofJobs) || len(bundle.Artifacts) == 0 {
		return fmt.Errorf("proof requires six jobs and a non-empty artifact inventory")
	}
	for index, job := range bundle.Jobs {
		if job.Name != proofJobs[index] || job.ID <= 0 || job.Conclusion != "success" || !validSHA(job.HeadSHA) || job.HeadSHA != bundle.HeadSHA {
			return fmt.Errorf("proof job %q is incomplete", job.Name)
		}
	}
	if bundle.Actors.Actor == "" || bundle.Actors.Builder == "" || bundle.Actors.Gate == "" || bundle.Actors.Builder != bundle.Actors.Actor {
		return fmt.Errorf("proof actor roles are incomplete")
	}
	if err := validateBranchProtection(bundle.BranchProtection, evidenceInput{Repository: bundle.Repository, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA, RunID: bundle.RunID, Attempt: bundle.RunAttempt, WorkflowSHA: bundle.WorkflowSHA, Digests: evidenceDigests{Policy: bundle.Digests.Policy}}, contextInput{BaseRef: bundle.BaseRef, EventRef: bundle.EventRef, CheckoutRef: bundle.CheckoutRef}); err != nil {
		return err
	}
	if err := validateDomainEvidence(bundle.DomainEvidence, evidenceInput{Repository: bundle.Repository, Event: bundle.Event, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA, RunID: bundle.RunID, Attempt: bundle.RunAttempt, WorkflowSHA: bundle.WorkflowSHA, Digests: evidenceDigests{Source: bundle.Digests.Source, IR: bundle.Digests.Semantic, Generated: bundle.Digests.Projection, Bundle: bundle.DomainEvidence.Digests.BundleSHA256}}, contextInput{EventRef: bundle.EventRef, CheckoutRef: bundle.CheckoutRef}); err != nil {
		return err
	}
	if len(bundle.Fixtures.Paths) == 0 || bundle.Fixtures.Status == "" || bundle.Fixtures.Source == "" || bundle.Fixtures.Semantic == "" || bundle.Fixtures.Provenance == "" || bundle.Scope.Decision == "" {
		return fmt.Errorf("proof fixture or scope evidence is incomplete")
	}
	if err := validateArtifacts(bundle.Artifacts); err != nil {
		return err
	}
	if err := validateCache(bundle.Cache, evidenceInput{HeadSHA: bundle.HeadSHA}); err != nil {
		return err
	}
	if err := validatePredecessors(bundle); err != nil {
		return err
	}
	if err := validateProofDigests(bundle.Digests); err != nil {
		return err
	}
	recorded := bundle.Digests.Bundle
	bundle.Digests.Bundle = ""
	payload, err := json.Marshal(bundle)
	if err != nil {
		return err
	}
	if digestBytes(payload) != recorded {
		return fmt.Errorf("proof bundle digest mismatch")
	}
	if bundle.Decision != "PASS" && bundle.Decision != "FAIL_CLOSED" {
		return fmt.Errorf("unknown proof decision %q", bundle.Decision)
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

func validateArtifacts(artifacts []artifactInput) error {
	for _, artifact := range artifacts {
		if artifact.ID <= 0 || artifact.Name == "" || artifact.Size <= 0 || artifact.Expired || !validDigest(artifact.Digest) {
			return fmt.Errorf("artifact inventory contains missing, zero, or expired artifact")
		}
	}
	return nil
}

func validateProofDigests(digests proofDigests) error {
	for _, digest := range []string{digests.Source, digests.Semantic, digests.Provenance, digests.Projection, digests.Build, digests.Policy, digests.Schema, digests.Toolchain, digests.Target, digests.Bundle} {
		if !validDigest(digest) {
			return fmt.Errorf("proof digest is missing, zero, or malformed")
		}
	}
	return nil
}
