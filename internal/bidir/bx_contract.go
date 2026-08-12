package bidir

import (
	"errors"
	"fmt"
)

// BXEvidenceSchemaVersion identifies the hard BX evidence contract.
const BXEvidenceSchemaVersion = "bidir-bx-evidence/v1"

// BXBaseEvidenceInput supplies the six required base artifacts.
type BXBaseEvidenceInput struct {
	DSL        Document
	IR         Model
	Go         FactSet
	SourceMap  []SourceSpan
	Evidence   FactSet
	Provenance []SourceSpan
}

// BXArtifactEvidence records one observed base artifact digest.
type BXArtifactEvidence struct {
	Hash  string
	Count int
}

// BXBaseEvidence is the measured form of the six base artifacts.
type BXBaseEvidence struct {
	DSL        BXArtifactEvidence
	IR         BXArtifactEvidence
	Go         BXArtifactEvidence
	SourceMap  BXArtifactEvidence
	Evidence   BXArtifactEvidence
	Provenance BXArtifactEvidence
}

// BXLStat is the observed metadata required for a source write boundary.
type BXLStat struct {
	Path   string
	Size   int64
	Mode   uint32
	Exists bool
}

// BXFileSnapshot is one observer-confirmed source state.
type BXFileSnapshot struct {
	Bytes []byte
	LStat BXLStat
}

// BXWriteObservation records before/after file observations.
type BXWriteObservation struct {
	Observed bool
	Before   BXFileSnapshot
	After    BXFileSnapshot
}

// BXRejectedWriteObserver owns before/after snapshots around a rejected
// operation. The producer supplies only the operation, never its snapshots.
type BXRejectedWriteObserver interface {
	Kind() string
	ObserveRejected(operation func() error) (BXWriteObservation, error)
}

// BXEvidenceFixture extends a reconciliation fixture with hard evidence.
type BXEvidenceFixture interface {
	ReconciliationFixture
	BaseEvidence() BXBaseEvidenceInput
	ObserveAcceptedWrite(before, after Document) BXWriteObservation
	RejectedWriteObserver(document Document) (BXRejectedWriteObserver, error)
}

// BXStateEvidence records all transaction dimensions as digests.
type BXStateEvidence struct {
	Semantic string
	Source   string
	Region   string
	Slot     string
	Bytes    string
	LStat    string
}

// BXTransactionEvidence records an observed before/after transaction.
type BXTransactionEvidence struct {
	Before         BXStateEvidence
	After          BXStateEvidence
	ObserverKind   string
	Observed       bool
	Atomic         bool
	NoWrite        bool
	Deferred       bool
	DeferredReason string
}

// BXEvidenceSpanSet records evidence IDs and source-span cardinality.
type BXEvidenceSpanSet struct {
	IDs                 []string
	FactKeys            []string
	Spans               []SourceSpan
	Records             []BXEvidenceRecord
	IDCount             int
	SpanCount           int
	Hash                string
	EvidenceIDAuthority string
}

// BXDeltaEvidence records ordered delta and locality evidence.
type BXDeltaEvidence struct {
	SequenceHash          string
	OrderHash             string
	CanonicalJSON         string
	Added                 []string
	Removed               []string
	Locality              Locality
	LocalityClosureHash   string
	LocalityCanonicalJSON string
	ClosureMembers        []ID
	ClosureValid          bool
	Candidates            []string
	PortSequence          []string
	RelationSequence      []string
	PortOrderHash         string
	RelationOrderHash     string
	EvidenceSpans         BXEvidenceSpanSet
	EvidenceHash          string
	PartialObservation    bool
	RemovedCreated        bool
	CandidatePromoted     bool
}

// BXConflictEvidence records an expected partial-information rejection.
type BXConflictEvidence struct {
	Kind              ConflictKind
	Count             int
	Transactional     bool
	NoWriteObserved   bool
	RemovedCreated    bool
	CandidatePromoted bool
}

// BXEvidence is the stable, reviewable output of one reconciliation fixture.
type BXEvidence struct {
	SchemaVersion        string
	Fixture              string
	Base                 BXBaseEvidence
	GetPutPassed         bool
	PutGetPassed         bool
	SemanticEquivalent   bool
	AcceptedRelationAdds int
	Delta                BXDeltaEvidence
	PartialDelta         BXDeltaEvidence
	AcceptedTransaction  BXTransactionEvidence
	RejectedTransaction  BXTransactionEvidence
	Locality             Locality
	PartialConflict      BXConflictEvidence
	Deferred             []string
}

func (e BXEvidence) validate() error {
	if e.SchemaVersion != BXEvidenceSchemaVersion {
		return fmt.Errorf("unsupported evidence schema %q", e.SchemaVersion)
	}
	if e.Fixture == "" {
		return errors.New("evidence fixture name is empty")
	}
	if err := validateBaseEvidence(e.Base); err != nil {
		return err
	}
	if !e.GetPutPassed || !e.PutGetPassed || !e.SemanticEquivalent {
		return errors.New("round-trip semantic evidence is not green")
	}
	if err := validateDeltaEvidence(e.Delta); err != nil {
		return fmt.Errorf("accepted delta evidence: %w", err)
	}
	if err := validateDeltaEvidence(e.PartialDelta); err != nil {
		return fmt.Errorf("partial delta evidence: %w", err)
	}
	if !e.PartialDelta.PartialObservation {
		return errors.New("partial delta is not marked as a partial observation")
	}
	if err := validateTransaction(e.AcceptedTransaction, false); err != nil {
		return fmt.Errorf("accepted transaction evidence: %w", err)
	}
	if err := validateTransaction(e.RejectedTransaction, true); err != nil {
		return fmt.Errorf("rejected transaction evidence: %w", err)
	}
	if e.Delta.CandidatePromoted || e.Delta.RemovedCreated || e.PartialDelta.CandidatePromoted || e.PartialDelta.RemovedCreated || e.PartialConflict.RemovedCreated || e.PartialConflict.CandidatePromoted {
		return errors.New("partial observation changed authoritative semantic state")
	}
	if e.RejectedTransaction.Deferred || !e.RejectedTransaction.NoWrite || !e.PartialConflict.NoWriteObserved {
		return errors.New("rejected transaction lacks observed atomic no-write evidence")
	}
	if e.PartialConflict.Kind == "" || e.PartialConflict.Count == 0 || !e.PartialConflict.Transactional {
		return errors.New("partial delta did not produce a transactional rejection")
	}
	for _, required := range deferredBXSeams() {
		if !containsString(e.Deferred, required) {
			return fmt.Errorf("deferred seam %q is missing", required)
		}
	}
	return nil
}

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

func validateState(state BXStateEvidence) error {
	if state.Semantic == "" || state.Source == "" || state.Region == "" || state.Slot == "" || state.Bytes == "" || state.LStat == "" {
		return errors.New("semantic/source/region/slot/bytes/lstat digest is missing")
	}
	return nil
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sameIDs(left, right []ID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
