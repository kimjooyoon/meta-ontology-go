package semantic

import (
	"errors"
	"testing"
)

func TestSameEdgeEvidenceRetainsDistinctSpans(t *testing.T) {
	ir := candidateHashIR(t, false, false)
	fact := NewUsedFact(
		MustIdentity("candidate-hash://activity/compile"),
		MustIdentity("candidate-hash://entity/source"),
	)
	before := ir.StableHash()
	beforeEvidenceHash := ir.EvidenceHash()
	first, err := NewEvidence(
		MustIdentity("candidate-hash://evidence/first"), GoVerifierID,
		VerificationEvidence, fact.Key(), StableHashString("edge evidence"),
	)
	if err != nil {
		t.Fatal(err)
	}
	first = first.WithSpan(Span{File: "first.gooo", Start: Position{Offset: 1}, End: Position{Offset: 5}})
	second, err := NewEvidence(
		MustIdentity("candidate-hash://evidence/second"), GoVerifierID,
		VerificationEvidence, fact.Key(), StableHashString("edge evidence"),
	)
	if err != nil {
		t.Fatal(err)
	}
	second = second.WithSpan(Span{File: "second.gooo", Start: Position{Offset: 8}, End: Position{Offset: 13}})
	if err := ir.AddEvidence(first); err != nil {
		t.Fatal(err)
	}
	if err := ir.AddEvidence(second); err != nil {
		t.Fatal(err)
	}
	if err := ir.Validate(); err != nil {
		t.Fatal(err)
	}
	if ir.StableHash() != before || ir.EvidenceHash() == beforeEvidenceHash || len(ir.Evidence()) != 2 {
		t.Fatal("evidence records changed the semantic edge or were not retained")
	}
	if ir.Evidence()[0].Span == ir.Evidence()[1].Span {
		t.Fatal("distinct evidence spans were collapsed")
	}
	conflict := first.WithSpan(Span{File: "changed.gooo", Start: Position{Offset: 20}, End: Position{Offset: 25}})
	if err := ir.AddEvidence(conflict); !errors.Is(err, ErrEvidenceConflict) {
		t.Fatalf("same evidence ID with changed span error = %v, want ErrEvidenceConflict", err)
	}
}
