package pathclosure

import (
	"bytes"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/semantic"
)

const CodeR4ProofValid = "PROOF_VALID_FINITE_BOUNDARY"

func r4Result(status Status, code, reason string, required []semantic.ID, cost int) R4Result {
	return R4Result{Status: status, Code: code, Reason: reason, RequiredPathIDs: sortedR4IDs(required), Cost: cost}
}

func r4Fail(code, reason string, required []semantic.ID, cost int) R4Result {
	return r4Result(FAIL_CLOSED, code, reason, required, cost)
}

func r4Unknown(code, reason string, required []semantic.ID, cost int) R4Result {
	return r4Result(UNKNOWN, code, reason, required, cost)
}

func invalidR4ID(value semantic.ID, label string) error {
	if _, err := semantic.ParseIdentity(value.String()); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	return nil
}

func duplicateR4IDs(values []semantic.ID) semantic.ID {
	seen := map[semantic.ID]struct{}{}
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return value
		}
		seen[value] = struct{}{}
	}
	return ""
}

// EvaluateR4 recomputes only the declared finite paths. It never discovers
// paths, accepts no expected/display label in the R4 input, and never emits
// promotion authorization.
func EvaluateR4(input R4Input) R4Result {
	required := sortedR4IDs(input.Boundary.RequiredPathIDs)
	cost := len(input.Records) + len(input.Receipts) + len(input.Paths)
	if input.Schema != R4SchemaVersion {
		return r4Fail(CodeInvalidPath, "invalid R4 schema", required, cost)
	}
	if duplicate := duplicateR4IDs(required); duplicate != "" {
		return r4Fail(CodeInvalidPath, "duplicate required path "+duplicate.String(), required, cost)
	}
	if len(required) == 0 {
		return r4Unknown(CodeMissingRequiredPaths, "no finite required paths were declared", required, cost)
	}
	if input.Boundary.OpenWorld {
		return r4Unknown(CodeOpenWorld, "open-world path closure is not a finite proof boundary", required, cost)
	}
	if !input.Boundary.Exhausted {
		return r4Unknown(CodeUnexhaustedBoundary, "finite path boundary was not explicitly exhausted", required, cost)
	}

	records, recordErr := indexR4Records(input.Records)
	if recordErr != nil {
		return r4Fail(CodeInvalidPath, recordErr.Error(), required, cost)
	}
	receipts, receiptErr := indexR4Receipts(input.Receipts)
	if receiptErr != nil {
		return r4Fail(CodeConflictingReceipt, receiptErr.Error(), required, cost)
	}
	paths, pathErr := indexR4Paths(input.Paths)
	if pathErr != nil {
		return r4Fail(CodeInvalidPath, pathErr.Error(), required, cost)
	}

	for _, pathID := range required {
		path, exists := paths[pathID]
		if !exists {
			return r4Unknown(CodeMissingRecord, "missing required path "+pathID.String(), required, cost)
		}
		status, code, reason := evaluateR4Path(path, records, receipts)
		if status == FAIL_CLOSED {
			return r4Fail(code, reason, required, cost)
		}
		if status == UNKNOWN {
			return r4Unknown(code, reason, required, cost)
		}
	}
	result := r4Result(PASS, CodeR4ProofValid, "all finite declared paths are covered", required, cost)
	result.CoveredPathIDs = append([]semantic.ID(nil), required...)
	result.ProofValid = true
	// This package has no promotion authority by contract. Keep this explicit
	// so a future caller cannot mistake finite proof validity for authorization.
	result.PromotionAuthorized = false
	return result
}

func indexR4Records(values []R4Record) (map[semantic.ID]R4Record, error) {
	result := make(map[semantic.ID]R4Record, len(values))
	for _, raw := range values {
		record, err := normalizeR4Record(raw)
		if err != nil {
			return nil, err
		}
		if _, exists := result[record.ID]; exists {
			return nil, fmt.Errorf("duplicate record %s", record.ID)
		}
		result[record.ID] = record
	}
	return result, nil
}

func indexR4Receipts(values []R4Receipt) (map[semantic.ID]R4Receipt, error) {
	result := make(map[semantic.ID]R4Receipt, len(values))
	events := make(map[semantic.ID]struct{}, len(values))
	for _, raw := range values {
		receipt, err := normalizeR4Receipt(raw)
		if err != nil {
			return nil, err
		}
		if receipt.ID == "" || receipt.EventID == "" {
			return nil, fmt.Errorf("receipt and append-only event IDs are required")
		}
		if _, exists := result[receipt.ID]; exists {
			return nil, fmt.Errorf("duplicate receipt %s", receipt.ID)
		}
		if _, exists := events[receipt.EventID]; exists {
			return nil, fmt.Errorf("conflicting append-only event %s", receipt.EventID)
		}
		result[receipt.ID] = receipt
		events[receipt.EventID] = struct{}{}
	}
	return result, nil
}

func indexR4Paths(values []R4Path) (map[semantic.ID]R4Path, error) {
	result := make(map[semantic.ID]R4Path, len(values))
	for _, raw := range values {
		path, err := normalizeR4Path(raw)
		if err != nil {
			return nil, err
		}
		if _, exists := result[path.ID]; exists {
			return nil, fmt.Errorf("duplicate path %s", path.ID)
		}
		result[path.ID] = path
	}
	return result, nil
}

func evaluateR4Path(path R4Path, records map[semantic.ID]R4Record, receipts map[semantic.ID]R4Receipt) (Status, string, string) {
	if len(path.RecordIDs) == 0 || len(path.RecordIDs) != len(path.RecordBytes) {
		return FAIL_CLOSED, CodeInvalidPath, "path record IDs and canonical record bytes are not a non-empty equal-length sequence"
	}
	if err := invalidR4ID(path.StartID, "path start ID"); err != nil {
		return FAIL_CLOSED, CodeInvalidPath, err.Error()
	}
	if err := invalidR4ID(path.EndID, "path end ID"); err != nil {
		return FAIL_CLOSED, CodeInvalidPath, err.Error()
	}
	var previous R4Record
	for index, recordID := range path.RecordIDs {
		record, exists := records[recordID]
		if !exists {
			return UNKNOWN, CodeMissingRecord, "missing record " + recordID.String()
		}
		provided, err := decodeCanonicalR4Record([]byte(path.RecordBytes[index]))
		if err != nil {
			return FAIL_CLOSED, CodeInvalidPath, fmt.Sprintf("record %s bytes: %v", recordID, err)
		}
		if provided.ID != record.ID {
			return FAIL_CLOSED, CodeInvalidPath, "record ID does not match ordered record ID"
		}
		actualBytes, _ := record.CanonicalRecordBytes()
		providedBytes, _ := provided.CanonicalRecordBytes()
		if !bytes.Equal(actualBytes, providedBytes) || !bytes.Equal(actualBytes, []byte(path.RecordBytes[index])) {
			return FAIL_CLOSED, CodeInvalidPath, "canonical record bytes do not match the record identity and fields"
		}
		if index == 0 {
			if record.PredecessorID != "" {
				return FAIL_CLOSED, CodeInvalidPath, "root record has a predecessor"
			}
			if record.SubjectID != path.StartID {
				return FAIL_CLOSED, CodeInvalidPath, "path start does not match first record subject"
			}
		} else {
			if record.PredecessorID != previous.ID {
				return FAIL_CLOSED, CodeInvalidPath, "record predecessor does not match ordered predecessor"
			}
			if previous.ObjectID != record.SubjectID {
				return FAIL_CLOSED, CodeInvalidPath, "ordered edge subject/object endpoints do not join"
			}
		}
		if index == len(path.RecordIDs)-1 && record.ObjectID != path.EndID {
			return FAIL_CLOSED, CodeInvalidPath, "path end does not match last record object"
		}
		if status, code, reason := validateR4Binding(record, receipts); status != PASS {
			return status, code, reason
		}
		previous = record
	}
	return PASS, CodeR4ProofValid, "path covered"
}

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
