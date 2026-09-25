package bidir

import (
	"fmt"
	"strings"
)

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
