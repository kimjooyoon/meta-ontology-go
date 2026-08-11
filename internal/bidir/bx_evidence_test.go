package bidir

import (
	"reflect"
	"testing"
)

type billingBXFixture struct{}

func (billingBXFixture) Name() string { return "billing" }

func (billingBXFixture) Document() Document { return billingDocument() }

func (billingBXFixture) AcceptedDelta() FactDelta {
	fact := NewSourcedFact(
		DeterministicFact,
		"billing://activity/pay-order",
		PredicateInvokes,
		"billing://activity/audit-payment",
		SourceSpan{File: "payment.go", Start: 42, End: 58},
	)
	fact.SubjectKind = ActivityKind
	fact.ObjectKind = ActivityKind
	return FactDelta{Added: FactSet{fact}}
}

func (billingBXFixture) PartialDelta() FactDelta {
	fact := NewFact(
		DeterministicFact,
		"billing://activity/pay-order",
		PredicateInvokes,
		"billing://activity/audit-payment",
	)
	return FactDelta{Added: FactSet{fact}}
}

func TestBXEvidenceMatchesBillingGolden(t *testing.T) {
	evidence, err := MeasureBXFixture(billingBXFixture{})
	if err != nil {
		t.Fatal(err)
	}
	want := "fixture=billing\n" +
		"get_put=pass\n" +
		"put_get=pass\n" +
		"semantic_equivalence=pass\n" +
		"accepted_relation_adds=1\n" +
		"touched=billing://activity/audit-payment,billing://activity/pay-order\n" +
		"affected=billing://activity/audit-payment,billing://activity/pay-order,billing://entity/audit,billing://entity/order,billing://entity/payment\n" +
		"partial_conflict=missing-source\n" +
		"partial_conflict_count=1\n" +
		"partial_transactional=pass\n"
	if got := evidence.Canonical(); got != want {
		t.Fatalf("golden evidence mismatch:\n got %s\nwant %s", got, want)
	}
}

func TestBXEvidenceSortsLocalityIDsBeforeCanonicalization(t *testing.T) {
	evidence := BXEvidence{
		Fixture:         "order",
		Locality:        Locality{Touched: []ID{"b", "a"}, Affected: []ID{"d", "c"}},
		PartialConflict: BXConflictEvidence{Transactional: true},
	}
	want := "fixture=order\n" +
		"get_put=fail\n" +
		"put_get=fail\n" +
		"semantic_equivalence=fail\n" +
		"accepted_relation_adds=0\n" +
		"touched=a,b\n" +
		"affected=c,d\n" +
		"partial_conflict=none\n" +
		"partial_conflict_count=0\n" +
		"partial_transactional=pass\n"
	if got := evidence.Canonical(); got != want {
		t.Fatalf("canonical evidence is not deterministic:\n got %s\nwant %s", got, want)
	}
	if !reflect.DeepEqual(evidence.Locality.Touched, []ID{"b", "a"}) {
		t.Fatal("canonicalization mutated the evidence")
	}
}
