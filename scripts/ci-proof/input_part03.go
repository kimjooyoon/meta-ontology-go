package main

import (
	"encoding/json"
	"fmt"
	"time"
)

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
	if !isPromotionContext(context) {
		if protection.Branch != context.BaseRef || protection.TokenSource != "not_observed" || protection.ReadStatus != "unavailable" || protection.Exists || protection.Strict || len(protection.RequiredChecks) != 0 || len(protection.RequiredCheckBindings) != 0 || protection.MissingReason != "trusted_guardian_required" || protection.EventRef != context.EventRef || protection.CheckoutRef != context.CheckoutRef || protection.RunID != evidence.RunID || protection.RunAttempt != evidence.Attempt || protection.WorkflowSHA != evidence.WorkflowSHA || protection.ObservedAt != nil || protection.ValidUntil != nil {
			return fmt.Errorf("non-promotion proof must keep branch protection explicitly unobserved")
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
func digestBranchProtection(protection branchProtection) string {
	protection.Digest = ""
	data, _ := json.Marshal(protection)
	return digestBytes(data)
}
