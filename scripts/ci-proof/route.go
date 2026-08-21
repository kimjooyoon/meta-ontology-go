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
	return err == nil && context.Route == route
}

func validBundleProofRoute(bundle proofBundle) bool {
	_, err := classifyProofRoute(bundle.Event, bundle.BaseRef)
	return err == nil
}

func isPromotionContext(context contextInput) bool {
	route, err := classifyProofRoute(context.Event, context.BaseRef)
	return err == nil && route == proofRoutePromotionMain
}

func isPromotionBundle(bundle proofBundle) bool {
	route, err := classifyProofRoute(bundle.Event, bundle.BaseRef)
	return err == nil && route == proofRoutePromotionMain
}
