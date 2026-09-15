package main

import (
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/verify"
)

func isFoundationPromotionContext(context contextInput) bool {
	return context.Route == proofRouteFoundationPromotion && context.Event == "pull_request" &&
		context.Repository == verify.FoundationPromotionRepository && context.PRNumber == verify.FoundationPromotionPRNumber &&
		context.BaseRef == verify.FoundationPromotionBaseBranch && context.BaseSHA == verify.FoundationPromotionBaseSHA
}

func foundationPromotionRejections(context contextInput) []string {
	if !isFoundationPromotionContext(context) {
		return nil
	}
	if err := validateFoundationPromotionEvidence(context.FoundationPromotion, context); err != nil {
		return []string{"foundation_promotion_" + err.Error()}
	}
	return nil
}

func validateFoundationPromotionBundle(bundle proofBundle) error {
	if isFoundationPromotionBundle(bundle) {
		if err := validateFoundationPromotionEvidence(bundle.FoundationPromotion, contextInput{Repository: bundle.Repository, Event: bundle.Event, Route: proofRouteFoundationPromotion, BaseRef: bundle.BaseRef, BaseSHA: bundle.BaseSHA, HeadSHA: bundle.HeadSHA, PRNumber: bundle.PRNumber, RunID: bundle.RunID, RunAttempt: bundle.RunAttempt}); err != nil {
			return fmt.Errorf("foundation promotion evidence is invalid: %w", err)
		}
		return nil
	}
	if bundle.FoundationPromotion != nil {
		return fmt.Errorf("foundation promotion evidence is not allowed on an ordinary proof")
	}
	return nil
}

func isFoundationPromotionBundle(bundle proofBundle) bool {
	return bundle.FoundationPromotion != nil && bundle.Event == "pull_request" &&
		bundle.Repository == verify.FoundationPromotionRepository && bundle.PRNumber == verify.FoundationPromotionPRNumber &&
		bundle.BaseRef == verify.FoundationPromotionBaseBranch && bundle.BaseSHA == verify.FoundationPromotionBaseSHA &&
		bundle.FoundationPromotion.HeadRef == verify.FoundationPromotionHeadBranch && bundle.FoundationPromotion.HeadSHA == bundle.HeadSHA
}
