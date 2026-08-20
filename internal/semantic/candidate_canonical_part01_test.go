package semantic

import (
	"testing"
)

func TestCandidateFactsStayOutOfAuthoritativeCanonical(t *testing.T) {
	for repetition := range 10 {
		reverse := repetition%2 == 1
		authoritative := candidateHashIR(t, false, reverse)
		observed := candidateHashIR(t, true, reverse)
		if err := authoritative.Validate(); err != nil {
			t.Fatal(err)
		}
		if err := observed.Validate(); err != nil {
			t.Fatal(err)
		}
		if authoritative.SemanticCanonical() != observed.SemanticCanonical() {
			t.Fatal("candidate fact entered authoritative semantic canonical form")
		}
		if authoritative.StableHash() != observed.StableHash() {
			t.Fatal("candidate fact changed authoritative stable hash")
		}
		comparison := CompareIR(authoritative, observed)
		if !comparison.Equivalent() {
			t.Fatalf("candidate-only observation changed IR comparison: %#v", comparison)
		}
		if len(observed.Graph.Candidates()) != 1 || observed.Graph.Canonical() == authoritative.Graph.Canonical() {
			t.Fatal("candidate observation was not retained in the audit projection")
		}
	}
}
func TestCandidatePromotionChangesAuthoritativeCanonicalExplicitly(t *testing.T) {
	ir := candidateHashIR(t, true, false)
	candidate := candidateHashFact()
	before := ir.StableHash()
	if _, err := ir.Graph.PromoteCandidate(candidate.Key()); err != nil {
		t.Fatal(err)
	}
	if len(ir.Graph.Candidates()) != 0 || len(ir.Graph.Facts()) != 2 {
		t.Fatalf("promotion state = facts:%d candidates:%d", len(ir.Graph.Facts()), len(ir.Graph.Candidates()))
	}
	if ir.StableHash() == before {
		t.Fatal("explicit candidate promotion did not change authoritative hash")
	}
	expected := candidateHashIR(t, false, false)
	candidate.Status = FactDeterministic
	if err := expected.AddFact(candidate); err != nil {
		t.Fatal(err)
	}
	if ir.StableHash() != expected.StableHash() {
		t.Fatal("promoted authoritative hash differs from deterministic graph")
	}
}
