package domaincapability

import "testing"

func TestPolicyCollectsBeforeInvestment(t *testing.T) {
	decision := (Policy{MinimumObservationCount: 3, RequireEvidenceDigest: true}).Decide(Measurement{
		ExpectedCapabilityCount:  4,
		ObservedCapabilityCount:  4,
		ObservationCount:         2,
		EvidenceBound:            true,
	})
	if decision != CollectMoreObservations {
		t.Fatalf("decision = %q, want %q", decision, CollectMoreObservations)
	}
}

func TestPolicyKeepsMissingBoundaryExplicit(t *testing.T) {
	decision := (Policy{MinimumObservationCount: 1, RequireEvidenceDigest: true}).Decide(Measurement{
		ExpectedCapabilityCount:   4,
		ObservedCapabilityCount:   3,
		UnresolvedCapabilityCount: 1,
		ObservationCount:          3,
		EvidenceBound:             true,
	})
	if decision != ReviewMissingCapabilities {
		t.Fatalf("decision = %q, want %q", decision, ReviewMissingCapabilities)
	}
}