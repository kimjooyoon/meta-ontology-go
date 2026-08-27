package ciplanusecase

import "github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"

func exactClaimTransitions(report metainvocation.Report) bool {
	if len(report.Claims) != 3 || report.Claims[0].ID != "source-program-binding" || report.Claims[0].Status != metainvocation.ClaimDischarged {
		return false
	}
	ruleClaim, planClaim := report.Claims[1], report.Claims[2]
	switch report.Decision {
	case metainvocation.DecisionPass:
		return ruleClaim.Status == metainvocation.ClaimDischarged && planClaim.Status == metainvocation.ClaimDischarged
	case metainvocation.DecisionClosed:
		return ruleClaim.Status == metainvocation.ClaimRefuted && planClaim.Status == metainvocation.ClaimOpen && planClaim.Reason == "DEPENDENCY_BLOCKED"
	case metainvocation.DecisionUnknown:
		return ruleClaim.Status == metainvocation.ClaimOpen && ruleClaim.Reason == "UNKNOWN_AT_RULE_SELECTION" && planClaim.Status == metainvocation.ClaimOpen && planClaim.Reason == "DEPENDENCY_BLOCKED"
	default:
		return false
	}
}

func exactUnknown(cause metainvocation.UnknownCause) bool {
	return cause.Stage == "RULE_SELECTION" && cause.Step == "classify-change" && cause.Reason == "NO_REGISTERED_RULE" && cause.File != ""
}
