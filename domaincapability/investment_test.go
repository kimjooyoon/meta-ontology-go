package domaincapability

import "testing"

func TestPlanDomainInvestmentUsesExplicitTargetInsteadOfCompletenessScore(t *testing.T) {
	plan := PlanDomainInvestment(InvestmentInput{
		DomainID:                "gooo.capability-discovery",
		ScopeDigest:             "sha256:scope",
		EvidenceDigest:          "sha256:evidence",
		DeclaredCapabilityCount: 10,
		ObservedCapabilityCount: 6,
		TargetNumerator:         8,
		TargetDenominator:       10,
		EvidenceComplete:        true,
		Comparable:              true,
	})
	if plan.Disposition != InvestmentInvest || plan.GapNumerator != 20 || plan.FirstBoundary != "observed_coverage_gap" {
		t.Fatalf("unexpected investment plan: %+v", plan)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("investment plan should validate: %v", err)
	}
}

func TestPlanDomainInvestmentPreservesUnresolvedBoundary(t *testing.T) {
	plan := PlanDomainInvestment(InvestmentInput{
		DomainID:                  "gooo.provenance",
		ScopeDigest:               "sha256:scope",
		EvidenceDigest:            "sha256:evidence",
		DeclaredCapabilityCount:   4,
		ObservedCapabilityCount:   3,
		UnresolvedCapabilityCount: 1,
		TargetNumerator:           3,
		TargetDenominator:         4,
		EvidenceComplete:          true,
		Comparable:                true,
	})
	if plan.Disposition != InvestmentInvestigate || plan.FirstBoundary != "unresolved_capability_boundary" {
		t.Fatalf("unexpected unresolved plan: %+v", plan)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("unresolved plan should validate: %v", err)
	}
}

func TestPlanDomainInvestmentDefersIncompleteEvidence(t *testing.T) {
	plan := PlanDomainInvestment(InvestmentInput{
		DomainID:                "gooo.syntax",
		ScopeDigest:             "sha256:scope",
		DeclaredCapabilityCount: 2,
		ObservedCapabilityCount: 1,
		TargetNumerator:         1,
		TargetDenominator:       2,
		Comparable:              true,
	})
	if plan.Disposition != InvestmentDeferred || plan.FirstBoundary != "evidence_completeness" {
		t.Fatalf("unexpected deferred plan: %+v", plan)
	}
	if err := plan.Validate(); err != nil {
		t.Fatalf("deferred plan should validate: %v", err)
	}
}
