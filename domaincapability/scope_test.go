package domaincapability

import "testing"

func TestScopeDigestIsOrderIndependent(t *testing.T) {
	first := ScopeDigest([]string{"gooo.provenance.compare", "gooo.declaration.diff"})
	second := ScopeDigest([]string{"gooo.declaration.diff", "gooo.provenance.compare"})
	if first != second {
		t.Fatalf("scope digest changed with ordering: %q != %q", first, second)
	}
}

func TestScopeChangedRejectsBoundaryDrift(t *testing.T) {
	previous := ScopeDigest([]string{"gooo.declaration.diff"})
	current := ScopeDigest([]string{"gooo.declaration.diff", "gooo.provenance.compare"})
	if !ScopeChanged(previous, current) {
		t.Fatal("expected changed capability boundary")
	}
}