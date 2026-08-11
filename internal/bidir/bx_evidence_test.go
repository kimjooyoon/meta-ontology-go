package bidir

import (
	"reflect"
	"strings"
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

func (billingBXFixture) BaseEvidence() BXBaseEvidenceInput {
	return fixtureBaseEvidence(billingDocument())
}

func (billingBXFixture) ObserveAcceptedWrite(before, after Document) BXWriteObservation {
	return fixtureWriteObservation(before, after)
}

func (billingBXFixture) ObserveRejectedWrite(before Document) BXWriteObservation {
	return fixtureWriteObservation(before, before)
}

func fixtureBaseEvidence(document Document) BXBaseEvidenceInput {
	model, _ := Get(document)
	facts := ProjectFacts(model)
	spans := []SourceSpan{{File: "billing.gooo", Start: 1, End: 2}}
	return BXBaseEvidenceInput{DSL: document, IR: model, Go: facts, SourceMap: spans, Evidence: facts, Provenance: spans}
}

func fixtureWriteObservation(before, after Document) BXWriteObservation {
	return BXWriteObservation{
		Observed: true,
		Before:   fixtureSnapshot(before),
		After:    fixtureSnapshot(after),
	}
}

func fixtureSnapshot(document Document) BXFileSnapshot {
	bytes := documentSourceBytes(document)
	return BXFileSnapshot{Bytes: bytes, LStat: BXLStat{Path: "billing.gooo", Size: int64(len(bytes)), Mode: 0o644, Exists: true}}
}

func TestBXEvidenceMatchesBillingGolden(t *testing.T) {
	evidence, err := MeasureBXFixture(billingBXFixture{})
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{
		"schema_version=" + BXEvidenceSchemaVersion,
		"fixture=billing",
		"get_put=pass",
		"put_get=pass",
		"semantic_equivalence=pass",
		"accepted_sequence_hash=",
		"accepted_order_hash=",
		"accepted_locality_closure_hash=",
		"accepted_evidence_id_count=1",
		"accepted_evidence_span_count=1",
		"accepted_before=",
		"accepted_after=",
		"rejected_no_write=pass",
		"partial_conflict=missing-source",
		"partial_removed_created=false",
		"partial_candidate_promoted=false",
		"deferred=generic gooo:invokes lifting",
	} {
		if !strings.Contains(evidence.Canonical(), line) {
			t.Fatalf("golden evidence omitted %q:\n%s", line, evidence.Canonical())
		}
	}
}

func TestBXEvidenceSortsLocalityIDsBeforeCanonicalization(t *testing.T) {
	evidence := BXEvidence{Locality: Locality{Touched: []ID{"b", "a"}, Affected: []ID{"d", "c"}}}
	if got := joinIDs(evidence.Locality.Touched); got != "a,b" {
		t.Fatalf("canonical locality is not sorted: %s", got)
	}
	if got := joinIDs(evidence.Locality.Affected); got != "c,d" {
		t.Fatalf("canonical closure is not sorted: %s", got)
	}
	if !reflect.DeepEqual(evidence.Locality.Touched, []ID{"b", "a"}) {
		t.Fatal("canonicalization mutated the evidence")
	}
}

func TestBXEvidenceRejectsMissingHardContract(t *testing.T) {
	fixture := incompleteBXFixture{}
	if _, err := MeasureBXFixture(fixture); err == nil || !strings.Contains(err.Error(), "hard BX evidence contract") {
		t.Fatalf("missing evidence contract was accepted: %v", err)
	}
	if _, err := MeasureBXFixture(missingArtifactsFixture{}); err == nil {
		t.Fatalf("missing base artifacts were accepted: %v", err)
	}
}

func TestBXDeltaEvidencePreservesFactOrderAndAtomicState(t *testing.T) {
	first := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "a.go", Start: 1, End: 2})
	second := NewSourcedFact(DeterministicFact, "billing://activity/audit-payment", PredicateInvokes, "billing://activity/pay-order", SourceSpan{File: "b.go", Start: 3, End: 4})
	left := FactDelta{Added: FactSet{first, second}}
	right := FactDelta{Added: FactSet{second, first}}
	if factSequenceHash(left) == factSequenceHash(right) || factOrderHash(left) == factOrderHash(right) {
		t.Fatal("delta evidence lost source sequence/order")
	}
	evidence, err := MeasureBXFixture(billingBXFixture{})
	if err != nil {
		t.Fatal(err)
	}
	if !evidence.RejectedTransaction.NoWrite || evidence.PartialConflict.RemovedCreated || evidence.PartialConflict.CandidatePromoted {
		t.Fatalf("rejected partial observation was not atomic and non-authoritative: %#v", evidence)
	}
}

type incompleteBXFixture struct{}

func (incompleteBXFixture) Name() string { return "incomplete" }

func (incompleteBXFixture) Document() Document { return billingDocument() }

func (incompleteBXFixture) AcceptedDelta() FactDelta { return FactDelta{} }

func (incompleteBXFixture) PartialDelta() FactDelta { return FactDelta{} }

type missingArtifactsFixture struct{ billingBXFixture }

func (missingArtifactsFixture) Name() string { return "missing-artifacts" }

func (missingArtifactsFixture) BaseEvidence() BXBaseEvidenceInput {
	return BXBaseEvidenceInput{DSL: billingDocument()}
}
