package domaincapability

import "testing"

func TestCompareScopesBlocksCrossBoundaryComparison(t *testing.T) {
	comparison := CompareScopes(
		[]string{"gooo.declaration.diff"},
		[]string{"gooo.declaration.diff", "gooo.provenance.compare"},
	)
	if comparison.Status != ScopeChangedStatus {
		t.Fatalf("status = %q, want %q", comparison.Status, ScopeChangedStatus)
	}
	if comparison.PreviousScopeDigest == comparison.CurrentScopeDigest {
		t.Fatal("scope digests unexpectedly match")
	}
}

func TestCompareScopesAllowsSameBoundaryInDifferentOrder(t *testing.T) {
	comparison := CompareScopes(
		[]string{"gooo.provenance.compare", "gooo.declaration.diff"},
		[]string{"gooo.declaration.diff", "gooo.provenance.compare"},
	)
	if comparison.Status != ScopeComparable {
		t.Fatalf("status = %q, want %q", comparison.Status, ScopeComparable)
	}
}