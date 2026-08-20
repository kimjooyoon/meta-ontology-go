package main

import (
	"fmt"
	"time"

	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

func readInputs(root, governancePath, evidencePath, jobsPath, schedulerPath, contextPath string) (proofInputs, error) {
	matrix, err := verify.ReadGovernanceMatrix(governancePath)
	if err != nil {
		return proofInputs{}, err
	}
	evidence, err := readJSON[evidenceInput](evidencePath)
	if err != nil {
		return proofInputs{}, fmt.Errorf("read CI evidence: %w", err)
	}
	scheduler, err := readScheduler(schedulerPath)
	if err != nil {
		return proofInputs{}, err
	}
	if err := validateSchedulerAgreement(evidence.Scheduler, scheduler, evidence.HeadSHA, evidence.RunID, evidence.Attempt); err != nil {
		return proofInputs{}, err
	}
	jobs, err := readJobsFor(jobsPath, evidence.Scheduler, evidence.HeadSHA, evidence.RunID, evidence.Attempt)
	if err != nil {
		return proofInputs{}, err
	}
	context, err := readJSON[contextInput](contextPath)
	if err != nil {
		return proofInputs{}, fmt.Errorf("read proof context: %w", err)
	}
	if err := validateInputIdentity(evidence, context, jobs); err != nil {
		return proofInputs{}, err
	}
	if err := validateEvidenceDigests(root, evidence); err != nil {
		return proofInputs{}, err
	}
	if err := validateBranchProtection(context.BranchProtection, evidence, context); err != nil {
		return proofInputs{}, err
	}
	if err := validateDomainEvidence(context.DomainEvidence, evidence, context); err != nil {
		return proofInputs{}, err
	}
	return proofInputs{Governance: governanceInput{
		Schema:           matrix.Schema,
		RequiredContexts: governanceContexts{Dev: matrix.RequiredContexts.Dev, Main: matrix.RequiredContexts.Main},
		GuardianContexts: guardianContexts{DevShadow: matrix.GuardianContexts.DevShadow, MainRequired: matrix.GuardianContexts.MainRequired},
		ProofJobs:        matrix.ProofJobs,
		Promotion:        promotionInput{Source: matrix.Promotion.Source, Target: matrix.Promotion.Target, RequiredChecks: matrix.Promotion.RequiredChecks, BranchProtectionRequired: matrix.Promotion.BranchProtectionRequired},
	}, Evidence: evidence, Scheduler: scheduler, Jobs: jobs, Context: context}, nil
}

func readScheduler(filename string) ([]schedulerInput, error) {
	scheduler, err := readJSON[[]schedulerInput](filename)
	if err != nil {
		return nil, fmt.Errorf("read proof-side scheduler evidence: %w", err)
	}
	return scheduler, nil
}

func readJobs(filename string) ([]jobInput, error) {
	return readJobsFor(filename, nil, "", 0, 0)
}

func readJobsFor(filename string, scheduler []schedulerInput, head string, runID, runAttempt int64) ([]jobInput, error) {
	jobs, err := readJSON[[]jobInput](filename)
	if err != nil {
		return nil, fmt.Errorf("read workflow jobs: %w", err)
	}
	byName := make(map[string]jobInput, len(proofJobs))
	seenIDs := make(map[int64]bool, len(proofJobs))
	for _, job := range jobs {
		if !isProofJob(job.Name) {
			continue
		}
		if _, exists := byName[job.Name]; exists {
			return nil, fmt.Errorf("duplicate canonical proof job %q", job.Name)
		}
		if job.ID <= 0 || seenIDs[job.ID] {
			return nil, fmt.Errorf("duplicate or invalid canonical proof job id %d", job.ID)
		}
		seenIDs[job.ID] = true
		byName[job.Name] = job
	}
	result := make([]jobInput, 0, len(proofJobs))
	schedulerByName := map[string]schedulerInput{}
	if scheduler != nil {
		schedulerByName, err = validateSchedulerInputs(scheduler, head, runID, runAttempt)
		if err != nil {
			return nil, err
		}
	}
	for _, name := range proofJobs {
		job, ok := byName[name]
		if !ok || job.ID <= 0 || job.Status == nil || !validSHA(job.HeadSHA) || job.RunID <= 0 || job.RunAttempt <= 0 {
			return nil, fmt.Errorf("canonical proof job %q is missing or unsuccessful", name)
		}
		state, err := jobObservationState(job, schedulerByName[name])
		if err != nil {
			return nil, fmt.Errorf("canonical proof job %q: %w", name, err)
		}
		job.ObservationState = state
		result = append(result, job)
	}
	return result, nil
}

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
	if err := compareJobs(evidence.Jobs, jobs, evidence.Scheduler, evidence.HeadSHA, evidence.RunID, evidence.Attempt); err != nil {
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

func compareJobs(expected, actual []jobInput, scheduler []schedulerInput, head string, runID, runAttempt int64) error {
	if len(expected) != len(proofJobs) || len(actual) != len(proofJobs) {
		return fmt.Errorf("proof must contain exactly six canonical jobs")
	}
	seenIDs := make(map[int64]bool, len(expected))
	schedulerByName, err := validateSchedulerInputs(scheduler, head, runID, runAttempt)
	if err != nil {
		return err
	}
	for index, name := range proofJobs {
		left, right := expected[index], actual[index]
		if left.Name != name || right.Name != name || left.ID != right.ID || left.ID <= 0 || seenIDs[left.ID] || !sameOptionalString(left.Status, right.Status) || !sameOptionalString(left.Conclusion, right.Conclusion) || left.ObservationState != right.ObservationState || right.HeadSHA != head || left.HeadSHA != head || left.RunID != runID || right.RunID != runID || left.RunAttempt != runAttempt || right.RunAttempt != runAttempt {
			return fmt.Errorf("proof job %q is missing or mismatched", name)
		}
		if state, err := jobObservationState(left, schedulerByName[name]); err != nil || state != left.ObservationState {
			return fmt.Errorf("proof job %q has an invalid observer state", name)
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

func validateBranchProtectionAt(protection branchProtection, evidence evidenceInput, context contextInput, now time.Time) error {
	if protection.Repository != evidence.Repository || protection.PolicySHA != evidence.Digests.Policy || !validSHA(protection.BaseSHA) || !validSHA(protection.HeadSHA) || protection.BaseSHA != evidence.BaseSHA || protection.HeadSHA != evidence.HeadSHA {
		return fmt.Errorf("branch protection snapshot is missing or unbound")
	}
	if protection.ReadStatus != "verified" && protection.ReadStatus != "unavailable" {
		return fmt.Errorf("branch protection snapshot source or status is invalid")
	}
	if protection.ReadStatus == "unavailable" && protection.MissingReason == "" || protection.ReadStatus == "verified" && protection.MissingReason != "" {
		return fmt.Errorf("branch protection missing reason is inconsistent with read status")
	}
	if protection.Digest != digestBranchProtection(protection) {
		return fmt.Errorf("branch protection snapshot digest mismatch")
	}
	if context.BaseRef == "dev" {
		if protection.Branch != "dev" || protection.TokenSource != "not_observed" || protection.ReadStatus != "unavailable" || protection.Exists || protection.Strict || len(protection.RequiredChecks) != 0 || len(protection.RequiredCheckBindings) != 0 || protection.MissingReason != "trusted_guardian_required" || protection.EventRef != context.EventRef || protection.CheckoutRef != context.CheckoutRef || protection.RunID != evidence.RunID || protection.RunAttempt != evidence.Attempt || protection.WorkflowSHA != evidence.WorkflowSHA || protection.ObservedAt != nil || protection.ValidUntil != nil {
			return fmt.Errorf("feature proof must keep branch protection explicitly unobserved")
		}
		return nil
	}
	if context.BaseRef != "main" || protection.Branch != "main" || protection.TokenSource != "github_app_installation" || protection.ReadStatus != "verified" || !branchProtectionReadyForAt(protection, "main", now) {
		return fmt.Errorf("main proof requires the trusted Guardian branch protection snapshot")
	}
	return nil
}

func validateTrustedBranchProtection(protection branchProtection, evidence evidenceInput, branch string) error {
	return validateTrustedBranchProtectionAt(protection, evidence, branch, time.Now().UTC())
}

func validateTrustedBranchProtectionAt(protection branchProtection, evidence evidenceInput, branch string, now time.Time) error {
	if protection.Repository != evidence.Repository || protection.Branch != branch || protection.PolicySHA != evidence.Digests.Policy || protection.TokenSource != "github_app_installation" || protection.ReadStatus != "verified" || !validSHA(protection.BaseSHA) || !validSHA(protection.HeadSHA) || protection.BaseSHA != evidence.BaseSHA || protection.HeadSHA != evidence.HeadSHA || protection.EventRef != evidence.EventRef || protection.CheckoutRef != evidence.CheckoutRef || protection.RunID != evidence.RunID || protection.RunAttempt != evidence.Attempt || protection.WorkflowSHA != evidence.WorkflowSHA || !branchProtectionReadyForAt(protection, branch, now) {
		return fmt.Errorf("trusted %s branch protection snapshot is missing or unbound", branch)
	}
	return nil
}

func branchProtectionReady(protection branchProtection) bool {
	return branchProtectionReadyFor(protection, "dev")
}

func requiredContextsForBase(base string) []string {
	if base == "main" {
		return append(append([]string(nil), proofJobs...), "CI guardian")
	}
	return append(append([]string(nil), proofJobs...), "CI guardian shadow")
}

func branchProtectionReadyFor(protection branchProtection, base string) bool {
	return branchProtectionReadyForAt(protection, base, time.Now().UTC())
}

func branchProtectionReadyForAt(protection branchProtection, base string, now time.Time) bool {
	return protection.ReadStatus == "verified" && protection.TokenSource == "github_app_installation" && protection.AppInstallationID > 0 && protection.AppSlug != "" && protection.Exists && protection.Strict && protection.EnforceAdmins && protection.RequiredReviews == 0 && !protection.DismissStaleReviews && !protection.RequireLastPushApproval && protection.LinearHistory && !protection.AllowForcePushes && !protection.AllowDeletions && !protection.RequiredSignatures && !protection.RequiredConversationResolution && !protection.BlockCreations && !protection.LockBranch && !protection.AllowForkSyncing && protection.Restrictions == nil && sameStringSet(protection.RequiredChecks, requiredContextsForBase(base)) && validRequiredCheckBindings(protection.RequiredCheckBindings, requiredContextsForBase(base)) && validObserverFreshness(protection.ObservedAt, protection.ValidUntil, now)
}
