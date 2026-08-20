package main

import (
	"fmt"
	"time"
)

func validateInputIdentity(evidence evidenceInput, context contextInput, jobs []jobInput) error {
	if evidence.Schema != evidenceSchema || evidence.Repository == "" || evidence.Event == "" || !validEventRef(evidence.Event, evidence.EventRef) || evidence.CheckoutRef != evidence.HeadSHA || evidence.BaseRef == "" || evidence.RunID <= 0 || evidence.Attempt <= 0 {
		return fmt.Errorf("CI evidence identity is incomplete")
	}
	if context.Repository != evidence.Repository || context.Event != evidence.Event || context.Ref != evidence.EventRef || context.EventRef != evidence.EventRef || context.CheckoutRef != evidence.CheckoutRef || context.BaseRef != evidence.BaseRef || context.BaseSHA != evidence.BaseSHA || context.HeadSHA != evidence.HeadSHA || context.WorkflowSHA != evidence.WorkflowSHA || context.RunID != evidence.RunID || context.RunAttempt != evidence.Attempt {
		return fmt.Errorf("proof context does not match CI evidence identity")
	}
	if context.Ref == "" || context.EventRef == "" || context.CheckoutRef == "" || context.EventRef != context.Ref || context.CheckoutRef != evidence.HeadSHA || context.Actor == "" || context.Builder == "" || context.Gate == "" || !validSHA(evidence.BaseSHA) || !validSHA(evidence.HeadSHA) || !validSHA(evidence.WorkflowSHA) || !validSHA(context.CheckoutRef) || evidence.BaseSHA == evidence.HeadSHA {
		return fmt.Errorf("proof identity is missing, invalid, or identical")
	}
	if evidence.Event != "pull_request" && evidence.Event != "push" {
		return fmt.Errorf("unsupported proof event %q", evidence.Event)
	}
	if evidence.Event == "pull_request" && context.PRNumber <= 0 {
		return fmt.Errorf("pull-request proof number is required")
	}
	if evidence.Event == "push" && context.PRNumber != 0 {
		return fmt.Errorf("push proof cannot carry a pull request number")
	}
	if err := compareJobs(evidence.Jobs, jobs, evidence.HeadSHA, evidence.RunID, evidence.Attempt); err != nil {
		return err
	}
	if err := validateArtifacts(context.Artifacts, context.RunID, context.RunAttempt); err != nil {
		return err
	}
	if err := validateMissingReasons(context.MissingReasons, context.DomainEvidence.ProtectionStatus, context.DomainEvidence.ProvenanceStatus); err != nil {
		return err
	}
	return nil
}
func compareJobs(expected, actual []jobInput, head string, runID, runAttempt int64) error {
	if len(expected) != len(proofJobs) || len(actual) != len(proofJobs) {
		return fmt.Errorf("proof must contain exactly six canonical jobs")
	}
	seenIDs := make(map[int64]bool, len(expected))
	for index, name := range proofJobs {
		left, right := expected[index], actual[index]
		if left.Name != name || right.Name != name || left.ID != right.ID || left.ID <= 0 || seenIDs[left.ID] || left.Status != "completed" || right.Status != "completed" || left.Conclusion != "success" || right.Conclusion != "success" || right.HeadSHA != head || left.HeadSHA != head || left.RunID != runID || right.RunID != runID || left.RunAttempt != runAttempt || right.RunAttempt != runAttempt {
			return fmt.Errorf("proof job %q is missing or mismatched", name)
		}
		seenIDs[left.ID] = true
	}
	return nil
}
func validateEvidenceDigests(root string, evidence evidenceInput) error {
	if !validDigest(evidence.Digests.Source) || !validDigest(evidence.Digests.IR) || !validDigest(evidence.Digests.Generated) || !validDigest(evidence.Digests.Policy) || !validDigest(evidence.Digests.Toolchain) || !validDigest(evidence.Digests.Bundle) {
		return fmt.Errorf("CI evidence has missing or malformed digests")
	}
	if root == "" {
		return fmt.Errorf("proof repository root is required")
	}
	return nil
}
func validateBranchProtection(protection branchProtection, evidence evidenceInput, context contextInput) error {
	return validateBranchProtectionAt(protection, evidence, context, time.Now().UTC())
}
