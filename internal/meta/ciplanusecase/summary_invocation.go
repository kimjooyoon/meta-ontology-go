package ciplanusecase

import "github.com/kimjooyoon/meta-ontology-go/internal/metainvocation"

func applyInvocationSummary(summary *Summary, report metainvocation.Report) {
	switch report.Decision {
	case metainvocation.DecisionPass:
		summary.PassDecisions++
	case metainvocation.DecisionClosed:
		summary.FailClosedDecisions++
	case metainvocation.DecisionUnknown:
		summary.UnknownDecisions++
	}
	for _, check := range report.Plan.Checks {
		summary.RuleEvidenceRefs += len(check.Reasons)
	}
	summary.PersistentClaims += len(report.Claims)
	for _, claim := range report.Claims {
		if claim.Reason == "UNKNOWN_AT_RULE_SELECTION" {
			summary.DirectUnknownClaims++
		}
		if claim.Reason == "DEPENDENCY_BLOCKED" {
			summary.DependencyBlocked++
		}
		if claim.Status == metainvocation.ClaimRefuted {
			summary.RefutedClaims++
		}
	}
	summary.RepositoryWrites += report.Effects.RepositoryWrites
	if report.Effects.MutationAuthority {
		summary.MutationAuthority++
	}
}
