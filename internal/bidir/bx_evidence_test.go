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

func (billingBXFixture) RejectedWriteObserver(document Document) (BXRejectedWriteObserver, error) {
	return NewBXMemoryRejectedWriteObserver(document), nil
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
		"accepted_port_order_hash=",
		"accepted_relation_order_hash=",
		"accepted_evidence_id_count=1",
		"accepted_evidence_span_count=1",
		"accepted_before=",
		"accepted_after=",
		"rejected_observer=memory-source",
		"rejected_deferred=fail",
		"partial_no_write=pass",
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
	if !evidence.Delta.ClosureValid || !sameIDs(evidence.Delta.ClosureMembers, evidence.Delta.Locality.Affected) {
		t.Fatalf("delta locality closure was not verified: %#v", evidence.Delta)
	}
	if evidence.Delta.PortOrderHash == "" || evidence.Delta.RelationOrderHash == "" || !strings.Contains(evidence.Delta.CanonicalJSON, "\"candidates\"") {
		t.Fatalf("delta canonical order/candidate schema is incomplete: %#v", evidence.Delta)
	}
	if evidence.RejectedTransaction.Deferred || !evidence.RejectedTransaction.NoWrite || !evidence.PartialConflict.NoWriteObserved || evidence.PartialConflict.RemovedCreated || evidence.PartialConflict.CandidatePromoted {
		t.Fatalf("rejected partial transaction was not observer-proven and non-authoritative: %#v", evidence)
	}
}

func TestBXEvidenceRejectsMissingCanonicalDeltaFields(t *testing.T) {
	evidence, err := MeasureBXFixture(billingBXFixture{})
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*BXDeltaEvidence){
		"closure-members":    func(delta *BXDeltaEvidence) { delta.ClosureMembers = nil },
		"closure-membership": func(delta *BXDeltaEvidence) { delta.ClosureMembers = []ID{"not-in-closure"} },
		"candidates":         func(delta *BXDeltaEvidence) { delta.Candidates = nil },
		"port-sequence":      func(delta *BXDeltaEvidence) { delta.PortSequence = nil },
		"evidence-hash":      func(delta *BXDeltaEvidence) { delta.EvidenceHash = "" },
	} {
		t.Run(name, func(t *testing.T) {
			copyEvidence := evidence
			mutate(&copyEvidence.Delta)
			if err := copyEvidence.validate(); err == nil {
				t.Fatal("missing canonical delta field was accepted")
			}
		})
	}
}

func TestEvidenceRetainsSameFactKeyWithDistinctIDsAndSpans(t *testing.T) {
	first := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "a.go", Start: 1, End: 2})
	second := NewSourcedFact(DeterministicFact, "billing://activity/pay-order", PredicateInvokes, "billing://activity/audit-payment", SourceSpan{File: "b.go", Start: 3, End: 4})
	evidence := evidenceSpans(FactSet{first, second})
	if evidence.IDCount != 2 || evidence.SpanCount != 2 || evidence.IDs[0] == evidence.IDs[1] {
		t.Fatalf("same FactKey evidence was collapsed: %#v", evidence)
	}
	if evidence.FactKeys[0] != evidence.FactKeys[1] || evidence.Hash == "" {
		t.Fatalf("same-edge evidence boundary lost key/hash: %#v", evidence)
	}
}

func TestBXRejectedObserverOwnsNoWriteSnapshots(t *testing.T) {
	document := billingDocument()
	observer := NewBXMemoryRejectedWriteObserver(document)
	called := false
	observation, err := observer.ObserveRejected(func() error {
		called = true
		return nil
	})
	if err != nil || !called || observer.Kind() != "memory-source" {
		t.Fatalf("observer did not run its operation: err=%v called=%t kind=%q", err, called, observer.Kind())
	}
	if !observation.Observed || !reflect.DeepEqual(observation.Before, observation.After) {
		t.Fatalf("observer did not prove no-write: %#v", observation)
	}
	observation.Before.Bytes[0] ^= 1
	second, err := observer.ObserveRejected(func() error { return nil })
	if err != nil || !reflect.DeepEqual(second.Before, second.After) {
		t.Fatalf("observer snapshots were not isolated: err=%v observation=%#v", err, second)
	}
}

func TestBXDeltaEvidenceRetainsRemovedCandidates(t *testing.T) {
	candidate := NewSourcedFact(CandidateFact, "billing://activity/pay-order", PredicateWasDerivedFrom, "billing://entity/order", SourceSpan{File: "removed.go", Start: 1, End: 2})
	evidence := makeDeltaEvidenceUnchecked(FactDelta{Removed: FactSet{candidate}}, Locality{}, true, Model{}, Model{})
	if len(evidence.Candidates) != 1 || !strings.Contains(evidence.CanonicalJSON, "\"candidates\"") {
		t.Fatalf("removed candidate was omitted from canonical evidence: %#v", evidence)
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
