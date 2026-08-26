package bidir

import (
	"errors"
	"fmt"
)

func validateBaseEvidence(base BXBaseEvidence) error {
	artifacts := []struct {
		name string
		item BXArtifactEvidence
	}{
		{"dsl", base.DSL}, {"ir", base.IR}, {"go", base.Go},
		{"source-map", base.SourceMap}, {"evidence", base.Evidence},
		{"provenance", base.Provenance},
	}
	for _, artifact := range artifacts {
		if artifact.item.Hash == "" || artifact.item.Count == 0 {
			return fmt.Errorf("base %s artifact is missing", artifact.name)
		}
	}
	return nil
}
func validateDeltaEvidence(delta BXDeltaEvidence) error {
	if delta.SequenceHash == "" || delta.OrderHash == "" || delta.CanonicalJSON == "" {
		return errors.New("delta sequence/order/canonical evidence is missing")
	}
	if delta.LocalityClosureHash == "" || delta.LocalityCanonicalJSON == "" || !delta.ClosureValid {
		return errors.New("locality closure evidence is missing")
	}
	if delta.ClosureMembers == nil || !sameIDs(delta.ClosureMembers, delta.Locality.Affected) {
		return errors.New("locality closure membership is incomplete")
	}
	if delta.Candidates == nil || delta.PortSequence == nil || delta.RelationSequence == nil {
		return errors.New("delta candidate/port/relation sequence is missing")
	}
	if delta.PortOrderHash == "" || delta.RelationOrderHash == "" {
		return errors.New("ordered port/relation hashes are missing")
	}
	if delta.EvidenceHash == "" || delta.EvidenceHash != delta.EvidenceSpans.Hash || !validEvidenceAuthority(delta.EvidenceSpans.EvidenceIDAuthority) || delta.EvidenceSpans.IDCount != len(delta.EvidenceSpans.IDs) || delta.EvidenceSpans.IDCount != len(delta.EvidenceSpans.FactKeys) || delta.EvidenceSpans.SpanCount != len(delta.EvidenceSpans.Spans) || !uniqueStrings(delta.EvidenceSpans.IDs) {
		return errors.New("evidence ID/span set is incomplete")
	}
	if err := validateEvidenceRecords(delta.EvidenceSpans); err != nil {
		return err
	}
	return validateDeltaConsistency(delta)
}
func validateTransaction(transaction BXTransactionEvidence, requireNoWrite bool) error {
	if requireNoWrite {
		if err := validateState(transaction.Before); err != nil {
			return fmt.Errorf("before state: %w", err)
		}
		if err := validateState(transaction.After); err != nil {
			return fmt.Errorf("after state: %w", err)
		}
		if transaction.ObserverKind == "" || !transaction.Observed || !transaction.Atomic || !transaction.NoWrite || transaction.Deferred || transaction.Before != transaction.After {
			return errors.New("rejected transaction observer did not prove atomic no-write")
		}
		return nil
	}
	if err := validateState(transaction.Before); err != nil {
		return fmt.Errorf("before state: %w", err)
	}
	if err := validateState(transaction.After); err != nil {
		return fmt.Errorf("after state: %w", err)
	}
	if !transaction.Observed || !transaction.Atomic {
		return errors.New("transaction was not observed atomically")
	}
	return nil
}
