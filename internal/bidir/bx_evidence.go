package bidir

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ReconciliationFixture supplies the source view and two Go-side updates for
// a measurable bidirectional contract experiment.
type ReconciliationFixture interface {
	Name() string
	Document() Document
	AcceptedDelta() FactDelta
	PartialDelta() FactDelta
}

// MeasureBXFixture runs the hard evidence contract for one fixture.
func MeasureBXFixture(fixture ReconciliationFixture) (BXEvidence, error) {
	if fixture == nil {
		return BXEvidence{}, errors.New("reconciliation fixture is nil")
	}
	contract, ok := fixture.(BXEvidenceFixture)
	if !ok {
		return BXEvidence{}, errors.New("fixture does not implement the hard BX evidence contract")
	}
	evidence := BXEvidence{SchemaVersion: BXEvidenceSchemaVersion, Fixture: strings.TrimSpace(fixture.Name())}
	if evidence.Fixture == "" {
		return BXEvidence{}, errors.New("reconciliation fixture name is empty")
	}
	document := fixture.Document()
	base, err := Get(document)
	if err != nil {
		return evidence, fmt.Errorf("get base document: %w", err)
	}
	evidence.Base, err = baseEvidence(contract.BaseEvidence(), document, base)
	if err != nil {
		return evidence, err
	}
	evidence.GetPutPassed = CheckGetPut(document) == nil
	acceptedDelta := fixture.AcceptedDelta()
	accepted, err := Reconcile(base, acceptedDelta)
	if err != nil {
		return evidence, fmt.Errorf("reconcile accepted delta: %w", err)
	}
	evidence.AcceptedRelationAdds = len(accepted.Delta.AddedRelations)
	evidence.Locality = accepted.Locality
	evidence.Delta, err = makeDeltaEvidence(acceptedDelta, accepted.Locality, false, base, accepted.Model)
	if err != nil {
		return evidence, fmt.Errorf("accepted delta evidence: %w", err)
	}
	updatedDocument, err := Put(document, accepted.Model)
	if err != nil {
		return evidence, fmt.Errorf("put accepted model: %w", err)
	}
	observed, err := Get(updatedDocument)
	if err != nil {
		return evidence, fmt.Errorf("get updated document: %w", err)
	}
	evidence.PutGetPassed = true
	evidence.SemanticEquivalent = SemanticEquivalent(accepted.Model, observed)
	evidence.AcceptedTransaction, err = acceptedTransaction(contract, document, updatedDocument, base, accepted)
	if err != nil {
		return evidence, err
	}
	observer, err := contract.RejectedWriteObserver(document)
	if err != nil {
		return evidence, fmt.Errorf("rejected write observer: %w", err)
	}
	evidence.PartialConflict, evidence.RejectedTransaction, evidence.PartialDelta, err = partialEvidence(document, base, fixture.PartialDelta(), observer)
	if err != nil {
		return evidence, err
	}
	evidence.Deferred = deferredBXSeams()
	if err := evidence.validate(); err != nil {
		return evidence, err
	}
	return evidence, nil
}

func acceptedTransaction(contract BXEvidenceFixture, before, after Document, base Model, result ReconcileResult) (BXTransactionEvidence, error) {
	observation := contract.ObserveAcceptedWrite(before, after)
	if err := observationMatches(observation, before, after); err != nil {
		return BXTransactionEvidence{}, fmt.Errorf("accepted write observation: %w", err)
	}
	return BXTransactionEvidence{
		Before:       stateEvidence(base, before, result.Locality, observation.Before),
		After:        stateEvidence(result.Model, after, result.Locality, observation.After),
		ObserverKind: "accepted-fixture",
		Observed:     observation.Observed,
		Atomic:       true,
	}, nil
}

func partialEvidence(document Document, base Model, delta FactDelta, observer BXRejectedWriteObserver) (BXConflictEvidence, BXTransactionEvidence, BXDeltaEvidence, error) {
	if observer == nil {
		return BXConflictEvidence{}, BXTransactionEvidence{}, BXDeltaEvidence{}, errors.New("rejected write observer is nil")
	}
	result := ReconcileResult{Model: base}
	var reconcileErr error
	called := false
	observation, observerErr := observer.ObserveRejected(func() error {
		called = true
		result, reconcileErr = Reconcile(base, delta)
		return reconcileErr
	})
	if observerErr != nil {
		return BXConflictEvidence{}, BXTransactionEvidence{}, BXDeltaEvidence{}, fmt.Errorf("observe rejected write: %w", observerErr)
	}
	if !called {
		return BXConflictEvidence{}, BXTransactionEvidence{}, BXDeltaEvidence{}, errors.New("rejected write observer did not run operation")
	}
	partial := makeDeltaEvidenceUnchecked(delta, LocalityBetween(base, result.Model), true, base, result.Model)
	before := stateEvidence(base, document, LocalityBetween(base, result.Model), observation.Before)
	after := stateEvidence(result.Model, document, LocalityBetween(base, result.Model), observation.After)
	transaction := BXTransactionEvidence{
		Before: before, After: after, ObserverKind: observer.Kind(), Observed: observation.Observed,
		Atomic: before == after, NoWrite: before == after,
	}
	evidence := BXConflictEvidence{Transactional: SemanticEquivalent(base, result.Model)}
	evidence.RemovedCreated = removedCreated(base, result.Model, delta)
	evidence.CandidatePromoted = candidatePromoted(base, delta, result.Model)
	var conflictErr *ReconcileError
	if !errors.As(reconcileErr, &conflictErr) {
		evidence.NoWriteObserved = transaction.NoWrite
		return evidence, transaction, partial, nil
	}
	evidence.Count = len(conflictErr.Conflicts)
	if evidence.Count > 0 {
		evidence.Kind = conflictErr.Conflicts[0].Kind
	}
	partial.RemovedCreated = evidence.RemovedCreated
	partial.CandidatePromoted = evidence.CandidatePromoted
	evidence.NoWriteObserved = transaction.NoWrite
	return evidence, transaction, partial, nil
}

func removedCreated(base, after Model, delta FactDelta) bool {
	allowed := make(map[string]struct{}, len(delta.Removed))
	for _, fact := range delta.Removed {
		allowed[fact.SemanticKey()] = struct{}{}
	}
	for _, relation := range base.Relations {
		if _, exists := findRelation(after, relation.Kind, relation.Source, relation.Target); exists {
			continue
		}
		if _, explicit := allowed[relationKey(relation.Kind, relation.Source, relation.Target)]; !explicit {
			return true
		}
	}
	return false
}

func candidatePromoted(base Model, delta FactDelta, model Model) bool {
	for _, fact := range delta.Added.ByLayer(CandidateFact) {
		if _, exists := findRelation(base, fact.Predicate, fact.Subject, fact.Object); exists {
			continue
		}
		if _, exists := findRelation(model, fact.Predicate, fact.Subject, fact.Object); exists {
			return true
		}
	}
	return false
}

func deferredBXSeams() []string {
	return []string{
		"generic gooo:invokes lifting",
		"PROV-O adapter mapping policy",
		"Go-lift/CLI delta atomicity",
		"rejected transaction filesystem/inode observer",
		"three-way merge",
	}
}

// Canonical returns line-oriented golden evidence with deterministic IDs.
func (e BXEvidence) Canonical() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "schema_version=%s\n", e.SchemaVersion)
	fmt.Fprintf(&builder, "fixture=%s\n", e.Fixture)
	fmt.Fprintf(&builder, "base_dsl=%s/%d\n", e.Base.DSL.Hash, e.Base.DSL.Count)
	fmt.Fprintf(&builder, "base_ir=%s/%d\n", e.Base.IR.Hash, e.Base.IR.Count)
	fmt.Fprintf(&builder, "base_go=%s/%d\n", e.Base.Go.Hash, e.Base.Go.Count)
	fmt.Fprintf(&builder, "base_source_map=%s/%d\n", e.Base.SourceMap.Hash, e.Base.SourceMap.Count)
	fmt.Fprintf(&builder, "base_evidence=%s/%d\n", e.Base.Evidence.Hash, e.Base.Evidence.Count)
	fmt.Fprintf(&builder, "base_provenance=%s/%d\n", e.Base.Provenance.Hash, e.Base.Provenance.Count)
	fmt.Fprintf(&builder, "get_put=%s\n", evidenceStatus(e.GetPutPassed))
	fmt.Fprintf(&builder, "put_get=%s\n", evidenceStatus(e.PutGetPassed))
	fmt.Fprintf(&builder, "semantic_equivalence=%s\n", evidenceStatus(e.SemanticEquivalent))
	fmt.Fprintf(&builder, "accepted_relation_adds=%d\n", e.AcceptedRelationAdds)
	writeDeltaCanonical(&builder, "accepted", e.Delta)
	writeDeltaCanonical(&builder, "partial", e.PartialDelta)
	fmt.Fprintf(&builder, "touched=%s\n", joinIDs(e.Locality.Touched))
	fmt.Fprintf(&builder, "affected=%s\n", joinIDs(e.Locality.Affected))
	writeTransactionCanonical(&builder, "accepted", e.AcceptedTransaction)
	writeTransactionCanonical(&builder, "rejected", e.RejectedTransaction)
	fmt.Fprintf(&builder, "partial_conflict=%s\n", conflictStatus(e.PartialConflict.Kind))
	fmt.Fprintf(&builder, "partial_conflict_count=%d\n", e.PartialConflict.Count)
	fmt.Fprintf(&builder, "partial_transactional=%s\n", evidenceStatus(e.PartialConflict.Transactional))
	fmt.Fprintf(&builder, "partial_no_write=%s\n", evidenceStatus(e.PartialConflict.NoWriteObserved))
	fmt.Fprintf(&builder, "partial_removed_created=%t\n", e.PartialConflict.RemovedCreated)
	fmt.Fprintf(&builder, "partial_candidate_promoted=%t\n", e.PartialConflict.CandidatePromoted)
	fmt.Fprintf(&builder, "deferred=%s\n", strings.Join(e.Deferred, ","))
	return builder.String()
}

func writeDeltaCanonical(builder *strings.Builder, label string, delta BXDeltaEvidence) {
	fmt.Fprintf(builder, "%s_sequence_hash=%s\n", label, delta.SequenceHash)
	fmt.Fprintf(builder, "%s_order_hash=%s\n", label, delta.OrderHash)
	fmt.Fprintf(builder, "%s_json=%s\n", label, delta.CanonicalJSON)
	fmt.Fprintf(builder, "%s_added=%s\n", label, strings.Join(delta.Added, ","))
	fmt.Fprintf(builder, "%s_removed=%s\n", label, strings.Join(delta.Removed, ","))
	fmt.Fprintf(builder, "%s_locality_closure_hash=%s\n", label, delta.LocalityClosureHash)
	fmt.Fprintf(builder, "%s_locality_json=%s\n", label, delta.LocalityCanonicalJSON)
	fmt.Fprintf(builder, "%s_candidates=%s\n", label, strings.Join(delta.Candidates, ","))
	fmt.Fprintf(builder, "%s_port_sequence=%s\n", label, strings.Join(delta.PortSequence, ","))
	fmt.Fprintf(builder, "%s_relation_sequence=%s\n", label, strings.Join(delta.RelationSequence, ","))
	fmt.Fprintf(builder, "%s_port_order_hash=%s\n", label, delta.PortOrderHash)
	fmt.Fprintf(builder, "%s_relation_order_hash=%s\n", label, delta.RelationOrderHash)
	fmt.Fprintf(builder, "%s_evidence_ids=%s\n", label, strings.Join(delta.EvidenceSpans.IDs, ","))
	fmt.Fprintf(builder, "%s_evidence_fact_keys=%s\n", label, strings.Join(delta.EvidenceSpans.FactKeys, ","))
	fmt.Fprintf(builder, "%s_evidence_spans=%s\n", label, strings.Join(spanTexts(delta.EvidenceSpans.Spans), ","))
	fmt.Fprintf(builder, "%s_evidence_records=%s\n", label, strings.Join(canonicalEvidenceRecordTexts(delta.EvidenceSpans.Records), ","))
	fmt.Fprintf(builder, "%s_evidence_id_count=%d\n", label, delta.EvidenceSpans.IDCount)
	fmt.Fprintf(builder, "%s_evidence_span_count=%d\n", label, delta.EvidenceSpans.SpanCount)
	fmt.Fprintf(builder, "%s_evidence_id_authority=%s\n", label, delta.EvidenceSpans.EvidenceIDAuthority)
	fmt.Fprintf(builder, "%s_evidence_span_hash=%s\n", label, delta.EvidenceSpans.Hash)
	fmt.Fprintf(builder, "%s_evidence_hash=%s\n", label, delta.EvidenceHash)
}

func canonicalEvidenceRecordTexts(records []BXEvidenceRecord) []string {
	values := make([]string, len(records))
	for index, record := range records {
		values[index] = strings.Join([]string{record.EvidenceID, record.FactKey, spanText(record.Span), fmt.Sprintf("%t", record.HasSpan)}, "|")
	}
	return values
}

func writeTransactionCanonical(builder *strings.Builder, label string, transaction BXTransactionEvidence) {
	fmt.Fprintf(builder, "%s_before=%s\n", label, stateCanonical(transaction.Before))
	fmt.Fprintf(builder, "%s_after=%s\n", label, stateCanonical(transaction.After))
	fmt.Fprintf(builder, "%s_observer=%s\n", label, transaction.ObserverKind)
	fmt.Fprintf(builder, "%s_observed=%s\n", label, evidenceStatus(transaction.Observed))
	fmt.Fprintf(builder, "%s_atomic=%s\n", label, evidenceStatus(transaction.Atomic))
	fmt.Fprintf(builder, "%s_no_write=%s\n", label, evidenceStatus(transaction.NoWrite))
	fmt.Fprintf(builder, "%s_deferred=%s\n", label, evidenceStatus(transaction.Deferred))
	fmt.Fprintf(builder, "%s_deferred_reason=%s\n", label, transaction.DeferredReason)
}

func stateCanonical(state BXStateEvidence) string {
	return strings.Join([]string{state.Semantic, state.Source, state.Region, state.Slot, state.Bytes, state.LStat}, "/")
}

func evidenceStatus(passed bool) string {
	if passed {
		return "pass"
	}
	return "fail"
}

func conflictStatus(kind ConflictKind) string {
	if kind == "" {
		return "none"
	}
	return string(kind)
}

func joinIDs(ids []ID) string {
	copyIDs := append([]ID(nil), ids...)
	sort.Slice(copyIDs, func(i, j int) bool { return copyIDs[i] < copyIDs[j] })
	values := make([]string, len(copyIDs))
	for index, id := range copyIDs {
		values[index] = string(id)
	}
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ",")
}
