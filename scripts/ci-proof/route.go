package main

import "fmt"

const (
	proofRouteFeatureDev        = "feature_dev"
	proofRoutePromotionMain     = "promotion_main"
	proofRouteProtectedPushDev  = "protected_push_dev"
	proofRouteProtectedPushMain = "protected_push_main"
)

func classifyProofRoute(event, baseRef string) (string, error) {
	switch event + ":" + baseRef {
	case "pull_request:dev":
		return proofRouteFeatureDev, nil
	case "pull_request:main":
		return proofRoutePromotionMain, nil
	case "workflow_dispatch:main":
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
	route, err := classifyProofRoute(context.Event, context.BaseRef)
	return err == nil && context.Route == route && validPromotionDispatchBinding(context.Event, context.Ref, context.PRNumber, context.BaseRef, context.HeadRef) && (context.BaseRef != "main" || (context.Event != "pull_request" && context.Event != "workflow_dispatch") || validMainPromotionHead(context.HeadRef))
}

func validBundleProofRoute(bundle proofBundle) bool {
	route, err := classifyProofRoute(bundle.Event, bundle.BaseRef)
	return err == nil && route == expectedProofRoute(bundle.Event, bundle.BaseRef) && validPromotionDispatchBinding(bundle.Event, bundle.Ref, bundle.PRNumber, bundle.BaseRef, bundle.HeadRef) && (bundle.BaseRef != "main" || (bundle.Event != "pull_request" && bundle.Event != "workflow_dispatch") || validMainPromotionHead(bundle.HeadRef))
}

func validPromotionDispatchBinding(event, ref string, prNumber int64, baseRef, headRef string) bool {
	if event != "workflow_dispatch" {
		return true
	}
	return baseRef == "main" && prNumber > 0 && headRef != "" && ref == "refs/heads/"+headRef
}

func isPromotionContext(context contextInput) bool {
	route, err := classifyProofRoute(context.Event, context.BaseRef)
	return err == nil && route == proofRoutePromotionMain && validPromotionDispatchBinding(context.Event, context.Ref, context.PRNumber, context.BaseRef, context.HeadRef) && validMainPromotionHead(context.HeadRef)
}

func isPromotionBundle(bundle proofBundle) bool {
	route, err := classifyProofRoute(bundle.Event, bundle.BaseRef)
	return err == nil && route == proofRoutePromotionMain && validPromotionDispatchBinding(bundle.Event, bundle.Ref, bundle.PRNumber, bundle.BaseRef, bundle.HeadRef) && validMainPromotionHead(bundle.HeadRef)
}

func validMainPromotionHead(headRef string) bool {
	return headRef == "dev" || validPromotionSnapshotHeadBranch(headRef)
}

func expectedProofRoute(event, baseRef string) string {
	route, err := classifyProofRoute(event, baseRef)
	if err != nil {
		return ""
	}
	return route
}
