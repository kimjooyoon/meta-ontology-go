package semantic

import (
	"errors"
	"os"
	"testing"
)

func TestEvidenceFreshnessFixtureRejectsChangedPayload(t *testing.T) {
	payload, err := os.ReadFile("testdata/evidence-freshness.txt")
	if err != nil {
		t.Fatal(err)
	}
	const wantDigest = "9ea86bfd939cc1218a0e2352a4623282e0d44739c88fe3f97f8f762efee68de0"
	if got := StableHash(payload); got != wantDigest {
		t.Fatalf("freshness fixture digest = %s, want %s", got, wantDigest)
	}
	fact := FactKey{
		Subject: MustIdentity("bootstrap://activity/verify"), Predicate: Used,
		Object: MustIdentity("bootstrap://entity/output"),
	}
	evidence, err := NewEvidence(
		MustIdentity("bootstrap://evidence/freshness"), GoVerifierID,
		VerificationEvidence, fact, StableHash(payload),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := evidence.ValidateFresh(payload); err != nil {
		t.Fatalf("fresh fixture was rejected: %v", err)
	}
	stalePayload := append(append([]byte(nil), payload...), '\n')
	if err := evidence.ValidateFresh(stalePayload); !errors.Is(err, ErrStaleEvidence) {
		t.Fatalf("changed fixture was accepted: %v", err)
	}
}
func TestPROVCanonicalDigestIsPinnedAndOrderIndependent(t *testing.T) {
	left := conformanceGraph(t, false)
	right := conformanceGraph(t, true)
	if left.SemanticCanonical() != right.SemanticCanonical() {
		t.Fatal("PROV canonical form changed with insertion order")
	}
	const wantDigest = "aa0e5232c314cf31b63caf6268578c3dc7b7763f97b9600d911ba776c568b156"
	if got := left.StableHash(); got != wantDigest {
		t.Fatalf("PROV canonical digest = %s, want %s", got, wantDigest)
	}
}
