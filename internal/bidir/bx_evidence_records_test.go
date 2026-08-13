package bidir

import (
	"reflect"
	"strings"
	"testing"
)

func TestEvidenceRecordsPreserveDuplicateFactKeysByPermutation(t *testing.T) {
	first := duplicateEvidenceFact("evidence-b", 30)
	second := duplicateEvidenceFact("evidence-a", 10)
	left := evidenceSpans(FactSet{first, second})
	right := evidenceSpans(FactSet{second, first})
	if !reflect.DeepEqual(left, right) {
		t.Fatalf("permuted duplicate evidence is not deterministic:\nleft %#v\nright %#v", left, right)
	}
	if left.EvidenceIDAuthority != "explicit" || len(left.Records) != 2 || left.Records[0].EvidenceID != second.EvidenceID || !left.Records[0].HasSpan {
		t.Fatalf("duplicate evidence records lost order or authority: %#v", left)
	}
	if left.Records[0].FactKey != left.Records[1].FactKey || left.IDCount != 2 || left.SpanCount != 2 {
		t.Fatalf("same-edge cardinality was not retained: %#v", left)
	}
}

func TestEvidenceAuthorityDistinguishesExplicitAndDerivedRecords(t *testing.T) {
	explicit := evidenceSpans(FactSet{duplicateEvidenceFact("explicit", 1)})
	derivedFact := duplicateEvidenceFact("", 1)
	derived := evidenceSpans(FactSet{derivedFact})
	mixed := evidenceSpans(FactSet{duplicateEvidenceFact("explicit", 1), derivedFact})
	if explicit.EvidenceIDAuthority != "explicit" || derived.EvidenceIDAuthority != "derived-non-authoritative" || mixed.EvidenceIDAuthority != "mixed-non-authoritative" {
		t.Fatalf("evidence authority classification is not fail-closed: explicit=%q derived=%q mixed=%q", explicit.EvidenceIDAuthority, derived.EvidenceIDAuthority, mixed.EvidenceIDAuthority)
	}
}

func TestBXDeltaCanonicalExportsEvidenceRecordsAndCounts(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	changes := FactDelta{Added: FactSet{duplicateEvidenceFact("evidence-a", 10), duplicateEvidenceFact("evidence-b", 20)}}
	evidence := makeDeltaEvidenceUnchecked(changes, Locality{}, false, base, base)
	if err := validateDeltaEvidence(evidence); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"evidence_records", "evidence_id_count", "evidence_span_count", "evidence_id_authority"} {
		if !strings.Contains(evidence.CanonicalJSON, `"`+field+`"`) {
			t.Fatalf("canonical delta omitted %q: %s", field, evidence.CanonicalJSON)
		}
	}
}

func TestBXDeltaEvidencePreservesOrderAndCanonicalizesModelPermutation(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	relation := Relation{Kind: PredicateInvokes, Source: "billing://activity/pay-order", Target: "billing://activity/audit-payment", Span: SourceSpan{File: "order.gooo", Start: 4, End: 8}}
	after, err := base.Apply(Delta{AddedRelations: []Relation{relation}})
	if err != nil {
		t.Fatal(err)
	}
	permuted := after.Clone()
	reverseModelCollections(&permuted)
	locality := LocalityBetween(base, after)
	changes := FactDelta{Added: FactSet{duplicateEvidenceFact("evidence-a", 10), duplicateEvidenceFact("evidence-b", 20)}}
	left := makeDeltaEvidenceUnchecked(changes, locality, false, base, after)
	right := makeDeltaEvidenceUnchecked(FactDelta{Added: FactSet{changes.Added[1], changes.Added[0]}}, locality, false, base, permuted)
	if left.SequenceHash == right.SequenceHash {
		t.Fatal("fact source sequence hash ignored observation order")
	}
	if left.OrderHash == right.OrderHash {
		t.Fatal("canonical order hash ignored observation order")
	}
	if err := validateDeltaEvidence(left); err != nil {
		t.Fatalf("left permutation evidence failed self-consistency: %v", err)
	}
	if err := validateDeltaEvidence(right); err != nil {
		t.Fatalf("right permutation evidence failed self-consistency: %v", err)
	}
	if left.EvidenceHash != right.EvidenceHash || !reflect.DeepEqual(left.EvidenceSpans, right.EvidenceSpans) {
		t.Fatal("permuted duplicate evidence changed canonical evidence boundary")
	}
	if left.PortOrderHash != right.PortOrderHash || left.RelationOrderHash != right.RelationOrderHash || !reflect.DeepEqual(left.RelationSequence, right.RelationSequence) {
		t.Fatal("model collection permutation changed canonical source order")
	}
}

func TestBXDeltaEvidenceDoesNotMutateInputsOrShareLocality(t *testing.T) {
	base, err := Get(billingDocument())
	if err != nil {
		t.Fatal(err)
	}
	relation := Relation{Kind: PredicateInvokes, Source: "billing://activity/pay-order", Target: "billing://activity/audit-payment", Span: SourceSpan{File: "no-write.gooo", Start: 4, End: 8}}
	after, err := base.Apply(Delta{AddedRelations: []Relation{relation}})
	if err != nil {
		t.Fatal(err)
	}
	locality := LocalityBetween(base, after)
	changes := FactDelta{Added: FactSet{duplicateEvidenceFact("evidence-a", 10)}}
	beforeChanges := cloneFactDelta(changes)
	beforeLocality := detachedLocality(locality)
	evidence := makeDeltaEvidenceUnchecked(changes, locality, false, base, after)
	evidence.Locality.Touched[0] = "mutated"
	evidence.EvidenceSpans.Records[0].EvidenceID = "mutated"
	if !reflect.DeepEqual(changes, beforeChanges) || !reflect.DeepEqual(locality, beforeLocality) {
		t.Fatal("delta evidence mutated an input")
	}
	if evidence.EvidenceSpans.Records[0].EvidenceID == changes.Added[0].EvidenceID {
		t.Fatal("evidence records share mutable state with input")
	}
}

func TestBXDeltaEvidenceRejectsTamperedEvidenceRecord(t *testing.T) {
	evidence, err := MeasureBXFixture(billingBXFixture{})
	if err != nil {
		t.Fatal(err)
	}
	copyEvidence := evidence
	copyEvidence.Delta.EvidenceSpans.Records = append([]BXEvidenceRecord(nil), evidence.Delta.EvidenceSpans.Records...)
	copyEvidence.Delta.EvidenceSpans.Records[0].FactKey = "tampered"
	if err := copyEvidence.validate(); err == nil {
		t.Fatal("tampered evidence record was accepted")
	}
}

func duplicateEvidenceFact(evidenceID string, start int) Fact {
	fact := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "duplicate.gooo", Start: start, End: start + 4})
	fact.EvidenceID = evidenceID
	fact.SubjectKind = ActivityKind
	fact.ObjectKind = ActivityKind
	return fact
}
