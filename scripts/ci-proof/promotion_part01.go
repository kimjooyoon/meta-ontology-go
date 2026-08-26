package main

import (
	"fmt"
)

const (
	promotionObservationCode   = "CI-PROMOTION-OBSERVATION-001"
	promotionAuthorizationCode = "CI-PROMOTION-AUTH-001"
)

func validPromotionObservation(observation *promotionObservation, repository string, prNumber int64, baseSHA, headSHA string) bool {
	return observation != nil && observation.Repository == repository && observation.PRNumber == prNumber && observation.Action != "" && observation.State == "open" && !observation.Draft && !observation.Merged && observation.Mergeable && observation.MergeableState == "clean" && observation.BaseRepo == repository && observation.BaseRef == "main" && observation.BaseSHA == baseSHA && observation.HeadRepo == repository && observation.HeadRef == "dev" && observation.HeadSHA == headSHA && validSHA(observation.BaseSHA) && validSHA(observation.HeadSHA) && validSHA(observation.LiveDevSHA) && validSHA(observation.LiveMainSHA) && observation.LiveDevSHA == headSHA && observation.LiveMainSHA == baseSHA && observation.Topology.Status == "ahead" && observation.Topology.AheadBy > 0 && observation.Topology.BehindBy == 0 && observation.Topology.MergeBaseSHA == baseSHA
}
func validPromotionObservationForContext(context contextInput) bool {
	if !isPromotionContext(context) {
		return context.PromotionObservation == nil
	}
	return validPromotionObservation(context.PromotionObservation, context.Repository, context.PRNumber, context.BaseSHA, context.HeadSHA)
}
func validatePromotionObservation(observation *promotionObservation, bundle proofBundle) error {
	if !isPromotionBundle(bundle) {
		if observation != nil {
			return fmt.Errorf("promotion observation is not allowed on a non-promotion proof")
		}
		return nil
	}
	if !validPromotionObservation(observation, bundle.Repository, bundle.PRNumber, bundle.BaseSHA, bundle.HeadSHA) {
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
	authorization := &promotionAuthorization{Decision: "FAIL_CLOSED", Code: new(promotionAuthorizationCode), Operation: "fast_forward", Source: "dev", Target: "main", BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA}
	if !validPromotionObservation(bundle.PromotionObservation, bundle.Repository, bundle.PRNumber, bundle.BaseSHA, bundle.HeadSHA) {
		authorization.Code = new(promotionObservationCode)
	} else if promotionProofCoreReady(bundle) {
		authorization.Decision = "PASS"
		authorization.Code = nil
	}
	return authorization
}
