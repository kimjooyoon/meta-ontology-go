package bidir

type billingBXFixture struct{}

func (billingBXFixture) Name() string       { return "billing" }
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
