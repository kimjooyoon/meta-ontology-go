package main

import (
	"fmt"
)

const (
	promotionObservationCode   = "CI-PROMOTION-OBSERVATION-001"
	promotionAuthorizationCode = "CI-PROMOTION-AUTH-001"
)

func validPromotionObservation(observation *promotionObservation, repository string, prNumber int64, baseSHA, headSHA, headRef string) bool {
	// GitHub can report "blocked" while a workflow_dispatch check suite is still
	// completing: the required job results are bound to the exact head below, but
	// the PR aggregate is reconciled only after the run closes. This observation
	// alone never authorizes a merge; promotionAuthorizationFor also requires the
	// complete passing proof, and the native PR merge API enforces branch rules.
	if observation == nil ||
		observation.Repository != repository || observation.PRNumber != prNumber || observation.Action == "" ||
		observation.State != "open" || observation.Draft || observation.Merged || !observation.Mergeable ||
		(observation.MergeableState != "clean" && observation.MergeableState != "unstable" && observation.MergeableState != "blocked") ||
		observation.BaseRepo != repository || observation.BaseRef != "main" || observation.BaseSHA != baseSHA ||
		observation.HeadRepo != repository || observation.HeadRef != headRef || observation.HeadSHA != headSHA ||
		!validSHA(observation.BaseSHA) || !validSHA(observation.HeadSHA) || !validSHA(observation.HeadParentSHA) ||
		!validSHA(observation.HeadTreeSHA) || !validSHA(observation.LiveDevSHA) || !validSHA(observation.LiveDevTreeSHA) ||
		!validSHA(observation.LiveMainSHA) || observation.LiveMainSHA != baseSHA ||
		observation.Topology.Status != "ahead" || observation.Topology.AheadBy == 0 ||
		observation.Topology.BehindBy != 0 || observation.Topology.MergeBaseSHA != baseSHA {
		return false
	}
	switch observation.Mode {
	case "DIRECT_DEV_HEAD":
		return observation.HeadRef == "dev" && observation.HeadSHA == observation.LiveDevSHA && observation.HeadTreeSHA == observation.LiveDevTreeSHA
	case "DEV_TREE_SNAPSHOT":
		return validPromotionSnapshotHeadBranch(observation.HeadRef) &&
			observation.HeadRef == "agent/main-promotion-snapshot-"+observation.LiveDevSHA &&
			observation.HeadSHA != observation.LiveDevSHA && observation.HeadParentSHA == baseSHA &&
			observation.HeadTreeSHA == observation.LiveDevTreeSHA && observation.Topology.AheadBy == 1
	default:
		return false
	}
}
func validPromotionObservationForContext(context contextInput) bool {
	if !isPromotionContext(context) {
		return context.PromotionObservation == nil
	}
	return validPromotionObservation(context.PromotionObservation, context.Repository, context.PRNumber, context.BaseSHA, context.HeadSHA, context.HeadRef)
}

func validatePromotionObservation(observation *promotionObservation, bundle proofBundle) error {
	if !isPromotionBundle(bundle) {
		if observation != nil {
			return fmt.Errorf("promotion observation is not allowed on a non-promotion proof")
		}
		return nil
	}
	if !validPromotionObservation(observation, bundle.Repository, bundle.PRNumber, bundle.BaseSHA, bundle.HeadSHA, bundle.HeadRef) {
		return fmt.Errorf("promotion PR observation is not open, clean, source-tree exact, or live-topology bound")
	}
	return nil
}
func promotionProofCoreReady(bundle proofBundle) bool {
	if bundle.Decision != "PASS" || len(bundle.Jobs) != len(proofJobs) {
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
	if validatePromotionObservation(bundle.PromotionObservation, bundle) != nil {
		authorization.Code = new(promotionObservationCode)
	} else if promotionProofCoreReady(bundle) {
		authorization.Decision = "PASS"
		authorization.Code = nil
	}
	return authorization
}
