package pathclosure

import (
	"bytes"
	"fmt"
	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

func decodeCanonicalR4Record(data []byte) (R4Record, error) {
	var wire r4WireRecord
	if err := decodeStrictR4(data, &wire); err != nil {
		return R4Record{}, err
	}
	canonical, err := marshalR4Record(wire)
	if err != nil {
		return R4Record{}, err
	}
	if !bytes.Equal(data, canonical) {
		return R4Record{}, fmt.Errorf("non-canonical JSON")
	}
	record := R4Record{ID: semantic.ID(wire.ID), SubjectID: semantic.ID(wire.SubjectID), ObjectID: semantic.ID(wire.ObjectID), ProviderID: semantic.ID(wire.ProviderID), ProviderDigest: wire.ProviderDigest, Phase: R4Phase(wire.Phase), PhaseDigest: wire.PhaseDigest, PredecessorID: semantic.ID(wire.PredecessorID), ReceiptID: semantic.ID(wire.ReceiptID), Writes: wire.Writes, Effect: wire.Effect}
	return normalizeR4Record(record)
}
func validateR4Binding(record R4Record, receipts map[semantic.ID]R4Receipt) (Status, string, string) {
	if record.ProviderID == "" || record.ProviderDigest == "" {
		return UNKNOWN, CodeMissingProvider, "record " + record.ID.String() + " has no provider binding"
	}
	if record.ReceiptID == "" {
		return UNKNOWN, CodeMissingEvidence, "record " + record.ID.String() + " has no evidence receipt binding"
	}
	receipt, exists := receipts[record.ReceiptID]
	if !exists {
		return UNKNOWN, CodeMissingEvidence, "missing evidence receipt " + record.ReceiptID.String()
	}
	if receipt.RecordID != record.ID || receipt.Writes != record.Writes || receipt.Effect != record.Effect {
		return FAIL_CLOSED, CodeConflictingReceipt, "receipt " + receipt.ID.String() + " conflicts with record " + record.ID.String()
	}
	if receipt.ProviderID == "" || receipt.ProviderDigest == "" {
		return UNKNOWN, CodeMissingProvider, "receipt " + receipt.ID.String() + " has no provider binding"
	}
	if receipt.ProviderID != record.ProviderID || receipt.ProviderDigest != record.ProviderDigest || receipt.Phase != record.Phase || receipt.PhaseDigest != record.PhaseDigest {
		return UNKNOWN, CodePhaseMismatch, "provider or phase digest is stale for record " + record.ID.String()
	}
	if (record.Effect != "" || !record.Writes) && receipt.ObserverID == "" {
		return UNKNOWN, CodeMissingObserver, "record " + record.ID.String() + " makes a no-write/effect claim without an observer-owned receipt"
	}
	if record.Phase != R4CompilePhase && record.Phase != R4RuntimePhase {
		return UNKNOWN, CodePhaseMismatch, "record " + record.ID.String() + " has no explicit supported phase binding"
	}
	if !validR4Digest(record.ProviderDigest) || !validR4Digest(record.PhaseDigest) || !validR4Digest(receipt.ProviderDigest) || !validR4Digest(receipt.PhaseDigest) {
		return UNKNOWN, CodePhaseMismatch, "provider or phase digest is not a current canonical digest for record " + record.ID.String()
	}
	return PASS, CodeR4ProofValid, "receipt binding is valid"
}

// R4Costs reports deterministic work units without using a weighted score in
// the decision. Costs are evidence, never compensation for an invalid path.
func R4Costs(input R4Input) int { return len(input.Records) + len(input.Receipts) + len(input.Paths) }
