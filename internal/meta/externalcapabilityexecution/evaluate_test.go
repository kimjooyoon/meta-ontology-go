package externalcapabilityexecution

import "testing"

func TestUnknownParentDecisionFailsClosed(t *testing.T) {
	report := RunCase("subject", "parent-decision-unknown")
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionUnknown {
		t.Fatalf("got %s/%s", report.Decision, report.Resolution)
	}
}

func TestExactCapabilityDoesNotPromoteParent(t *testing.T) {
	report := RunCase("subject", "exact")
	if report.Decision != DecisionExecutable || report.PromotionCount != 0 {
		t.Fatalf("decision=%s promotions=%d", report.Decision, report.PromotionCount)
	}
	if report.Parent.Decision != DecisionFailClosed || report.Parent.Completed != 6 {
		t.Fatalf("parent=%s completed=%d", report.Parent.Decision, report.Parent.Completed)
	}
}

func TestFixedDenominators(t *testing.T) {
	if MetricDenominator != 10 || len(caseIDs) != SuiteDenominator {
		t.Fatalf("metrics=%d cases=%d", MetricDenominator, len(caseIDs))
	}
}

func TestInvalidPrincipalEvidenceFailsClosedWithoutGrantingCapability(t *testing.T) {
	binding, err := BindPrincipal(principalSourceDigest, "spiffe://ci.example.org", "spiffe://ci.example.org/workload/gooo", "gooo://external-capability/execute", principalNonce)
	if err != nil {
		t.Fatal(err)
	}
	binding.Subject = "spiffe://ci.example.org/workload/tampered"
	observation := exactObservation("subject")
	observation.Principal = &binding
	report := Evaluate(observation)
	if report.Decision != DecisionFailClosed || report.Resolution != ResolutionUnknown || report.EnforcementEffect != EffectBlock || report.Reason != ReasonPrincipal {
		t.Fatalf("invalid principal crossed capability boundary: %#v", report)
	}
	if report.Principal == nil || report.Principal.Subject != binding.Subject || report.PromotionCount != 0 || report.RepositoryWrites != 0 || report.ExternalRepositoryWrites != 0 {
		t.Fatalf("invalid principal report lost provenance or changed authority: %#v", report)
	}
}
