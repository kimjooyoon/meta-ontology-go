package main

import (
	"fmt"
)

func validatePromotionAuthorization(bundle proofBundle) error {
	if bundle.BaseRef != "main" {
		if bundle.PromotionAuthorization != nil || bundle.PromotionObservation != nil {
			return fmt.Errorf("promotion authorization is not allowed on a feature proof")
		}
		return nil
	}
	authorization := bundle.PromotionAuthorization
	if authorization == nil || authorization.Operation != "fast_forward" || authorization.Source != "dev" || authorization.Target != "main" || authorization.BaseSHA != bundle.BaseSHA || authorization.HeadSHA != bundle.HeadSHA || !validSHA(authorization.BaseSHA) || !validSHA(authorization.HeadSHA) || !validDigest(authorization.ProofDigest) || authorization.ProofDigest != bundle.Digests.Bundle {
		return fmt.Errorf("promotion authorization is missing or not bound to the proof digest")
	}
	if authorization.Decision == "FAIL_CLOSED" {
		if authorization.Code == nil || (*authorization.Code != promotionAuthorizationCode && *authorization.Code != promotionObservationCode) {
			return fmt.Errorf("fail-closed promotion authorization has no reason code")
		}
		return nil
	}
	if authorization.Decision != "PASS" || authorization.Code != nil {
		return fmt.Errorf("promotion authorization decision is malformed")
	}
	if !promotionProofCoreReady(bundle) {
		return fmt.Errorf("promotion authorization is not backed by a complete passing proof")
	}
	return validatePromotionObservation(bundle.PromotionObservation, bundle)
}
func stringPointer(value string) *string {
	return &value
}
