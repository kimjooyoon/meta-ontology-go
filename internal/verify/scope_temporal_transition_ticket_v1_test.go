package verify

import "testing"

func TestTemporalTransitionTicketV1ScopeIsExplicit(t *testing.T) {
	paths, ok := BranchScope(temporalTransitionTicketV1Branch)
	if !ok || len(paths) != 12 {
		t.Fatalf("temporal transition ticket scope: known=%t paths=%d", ok, len(paths))
	}
	allowed := []string{
		".github/workflows/temporal-transition-ticket-v1.yml",
		"contracts/temporal-transition-ticket-v1.json",
		"examples/temporal-transition-ticket/main.gooo",
		"fixtures/temporal-transition-ticket/evidence.json",
		"scripts/evaluate-temporal-transition-ticket-v1.sh",
	}
	if err := CheckPathScopeForBranch(allowed, temporalTransitionTicketV1Branch); err != nil {
		t.Fatalf("representative paths rejected: %v", err)
	}
	if err := CheckPathScopeForBranch([]string{"scripts/semantic-conformance.sh"}, temporalTransitionTicketV1Branch); err == nil {
		t.Fatal("unrelated script accepted")
	}
}
