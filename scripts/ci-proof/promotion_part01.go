package main

import (
	"fmt"
	"strings"
)

const (
	promotionObservationCode   = "CI-PROMOTION-OBSERVATION-001"
	promotionAuthorizationCode = "CI-PROMOTION-AUTH-001"
)

func validPromotionObservation(observation *promotionObservation, repository string, prNumber int64, baseSHA, headSHA string) bool {
	return observation != nil && observation.Repository == repository && observation.PRNumber == prNumber && observation.Action != "" && observation.State == "open" && !observation.Draft && !observation.Merged && observation.Mergeable && observation.MergeableState == "clean" && observation.BaseRepo == repository && observation.BaseRef == "main" && observation.BaseSHA == baseSHA && observation.HeadRepo == repository && observation.HeadRef == "dev" && observation.HeadSHA == headSHA && validSHA(observation.BaseSHA) && validSHA(observation.HeadSHA) && validSHA(observation.LiveDevSHA) && validSHA(observation.LiveMainSHA) && observation.LiveDevSHA == headSHA && observation.LiveMainSHA == baseSHA && observation.Topology.Status == "ahead" && observation.Topology.AheadBy > 0 && observation.Topology.BehindBy == 0 && observation.Topology.MergeBaseSHA == baseSHA
}
func validPromotionObservationForContext(context contextInput) bool {
	if isReconciliationContext(context) {
		return validReconciliationObservation(context.PromotionObservation, context.Repository, context.PRNumber, context.BaseSHA, context.HeadSHA, context.HeadRef)
	}
	if !isPromotionContext(context) {
		return context.PromotionObservation == nil
	}
	return validPromotionObservation(context.PromotionObservation, context.Repository, context.PRNumber, context.BaseSHA, context.HeadSHA)
}

func validReconciliationObservation(observation *promotionObservation, repository string, prNumber int64, baseSHA, headSHA, headRef string) bool {
	return observation != nil && observation.Repository == repository && observation.PRNumber == prNumber && observation.Action != "" && observation.State == "open" && !observation.Draft && !observation.Merged && observation.Mergeable && observation.MergeableState == "clean" && observation.BaseRepo == repository && observation.BaseRef == "main" && observation.BaseSHA == baseSHA && observation.HeadRepo == repository && observation.HeadRef == headRef && strings.HasPrefix(headRef, "agent/main-history-reconciliation-") && observation.HeadSHA == headSHA && validSHA(observation.BaseSHA) && validSHA(observation.HeadSHA) && validSHA(observation.LiveDevSHA) && validSHA(observation.LiveMainSHA) && observation.LiveMainSHA == baseSHA && observation.Topology.MergeBaseSHA != ""
}
func validatePromotionObservation(observation *promotionObservation, bundle proofBundle) error {
	if !isPromotionBundle(bundle) {
		if observation != nil {
			return fmt.Errorf("promotion observation is not allowed on a non-promotion proof")
		}
		return nil
	}
	valid := validPromotionObservation(observation, bundle.Repository, bundle.PRNumber, bundle.BaseSHA, bundle.HeadSHA)
	if isReconciliationBundle(bundle) {
		valid = validReconciliationObservation(observation, bundle.Repository, bundle.PRNumber, bundle.BaseSHA, bundle.HeadSHA, bundle.HeadRef)
	}
	if !valid {
		return fmt.Errorf("promotion PR observation is not open, clean, exact dev-to-main, or live-topology bound")
	}
	return nil
}
func promotionProofCoreReady(bundle proofBundle) bool {
	if bundle.Decision != "PASS" || !branchProtectionReadyFor(bundle.BranchProtection, "main") || validateGuardianEvidence(bundle.GuardianEvidence, bundle) != nil || len(bundle.Jobs) != len(proofJobs) {
		return false
	}
	for index, job := range bundle.Jobs {
		if job.Name != proofJobs[index] || job.Status != "completed" || job.Conclusion != "success" || job.HeadSHA != bundle.HeadSHA || job.RunID != bundle.RunID || job.RunAttempt != bundle.RunAttempt {
			return false
		}
	}
	return validateArtifacts(bundle.Artifacts, bundle.RunID, bundle.RunAttempt) == nil
}
func promotionAuthorizationFor(bundle proofBundle) *promotionAuthorization {
	if !isPromotionBundle(bundle) {
		return nil
	}
	operation := "fast_forward"
	if isReconciliationBundle(bundle) {
		operation = "squash_linear"
	}
	authorization := &promotionAuthorization{Decision: "FAIL_CLOSED", Code: new(promotionAuthorizationCode), Operation: operation, Source: "dev", Target: "main", BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA}
	if validatePromotionObservation(bundle.PromotionObservation, bundle) != nil {
		authorization.Code = new(promotionObservationCode)
	} else if promotionProofCoreReady(bundle) {
		authorization.Decision = "PASS"
		authorization.Code = nil
	}
	return authorization
}
