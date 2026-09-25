package externalconformanceactivation

import "testing"

func capsule(input *Input, name string) *[]byte {
	switch name {
	case "assurance":
		return &input.Assurance
	case "eligibility":
		return &input.Eligibility
	default:
		return &input.Merge
	}
}

func assertBlocked(t *testing.T, receipt Receipt, resolution, reason string) {
	t.Helper()
	if receipt.Decision != DecisionFailClosed || receipt.Resolution != resolution ||
		receipt.EnforcementEffect != EffectBlock || receipt.Reason != reason ||
		receipt.TransitionApplied != 0 || receipt.RepositoryWrites != 0 {
		t.Fatalf("receipt=%+v", receipt)
	}
}
