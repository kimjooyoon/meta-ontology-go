package main

import (
	"fmt"
	"strings"
)

const (
	proofRouteFeatureDev          = "feature_dev"
	proofRoutePromotionMain       = "promotion_main"
	proofRouteReconciliationMain  = "reconciliation_main"
	proofRouteFoundationPromotion = "foundation_promotion"
	proofRouteProtectedPushDev    = "protected_push_dev"
	proofRouteProtectedPushMain   = "protected_push_main"
)

func classifyProofRoute(event, baseRef string) (string, error) {
	switch event + ":" + baseRef {
	case "pull_request:dev":
		return proofRouteFeatureDev, nil
	case "pull_request:main":
		return proofRoutePromotionMain, nil
	case "push:dev":
		return proofRouteProtectedPushDev, nil
	case "push:main":
		return proofRouteProtectedPushMain, nil
	default:
		return "", fmt.Errorf("unsupported CI proof route tuple")
	}
}

func validContextProofRoute(context contextInput) bool {
	if context.Route == proofRouteFoundationPromotion {
		return isFoundationPromotionContext(context) && context.FoundationPromotion != nil && context.FoundationPromotion.HeadSHA == context.HeadSHA
	}
	route, err := classifyProofRouteWithHead(context.Event, context.BaseRef, context.HeadRef)
	return err == nil && context.Route == route && context.FoundationPromotion == nil
}

func validBundleProofRoute(bundle proofBundle) bool {
	if bundle.FoundationPromotion != nil {
		return isFoundationPromotionBundle(bundle)
	}
	route, err := classifyProofRouteWithHead(bundle.Event, bundle.BaseRef, bundle.HeadRef)
	return err == nil && route == expectedProofRoute(bundle.Event, bundle.BaseRef, bundle.HeadRef)
}

func isPromotionContext(context contextInput) bool {
	if context.Route == proofRouteFoundationPromotion || context.FoundationPromotion != nil {
		return false
	}
	route, err := classifyProofRouteWithHead(context.Event, context.BaseRef, context.HeadRef)
	return err == nil && (route == proofRoutePromotionMain || route == proofRouteReconciliationMain)
}

func isPromotionBundle(bundle proofBundle) bool {
	if bundle.FoundationPromotion != nil {
		return false
	}
	route, err := classifyProofRouteWithHead(bundle.Event, bundle.BaseRef, bundle.HeadRef)
	return err == nil && (route == proofRoutePromotionMain || route == proofRouteReconciliationMain)
}

func classifyProofRouteWithHead(event, baseRef, headRef string) (string, error) {
	const prefix = "agent/main-history-reconciliation-"
	if event == "pull_request" && baseRef == "main" && strings.HasPrefix(headRef, prefix) && len(headRef) > len(prefix) {
		return proofRouteReconciliationMain, nil
	}
	if event == "pull_request" && baseRef == "main" && headRef != "" && headRef != "dev" {
		return "", fmt.Errorf("ordinary agent-to-main route is not authorized")
	}
	return classifyProofRoute(event, baseRef)
}

func expectedProofRoute(event, baseRef, headRef string) string {
	route, err := classifyProofRouteWithHead(event, baseRef, headRef)
	if err != nil {
		return ""
	}
	return route
}

func isReconciliationContext(context contextInput) bool {
	return context.Route == proofRouteReconciliationMain && context.Event == "pull_request" && context.BaseRef == "main" && strings.HasPrefix(context.HeadRef, "agent/main-history-reconciliation-")
}

func isReconciliationBundle(bundle proofBundle) bool {
	return bundle.Event == "pull_request" && bundle.BaseRef == "main" && strings.HasPrefix(bundle.HeadRef, "agent/main-history-reconciliation-")
}
