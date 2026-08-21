package generation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

func (receipt *OperationReceipt) UnmarshalJSON(data []byte) error {
	type wire OperationReceipt
	var decoded wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return fmt.Errorf("decode operation receipt: %w", err)
	}
	if err := ensureIndicatorLedgerEOF(decoder); err != nil {
		return fmt.Errorf("decode operation receipt: %w", err)
	}
	candidate := OperationReceipt(decoded)
	if candidate.SchemaVersion != OperationReceiptSchemaVersion {
		return fmt.Errorf("unsupported operation receipt schema %q", candidate.SchemaVersion)
	}
	if !validIndicatorDecisionLedgerDigest(candidate.IndicatorDecisionLedgerDigest) ||
		candidate.IndicatorDecisionLedgerCount < 1 {
		return fmt.Errorf("invalid operation receipt indicator ledger provenance")
	}
	if !reflect.DeepEqual(candidate.Indicators,
		normalizeIndicatorReceipts(candidate.Indicators)) {
		return fmt.Errorf("operation receipt indicators are not canonical")
	}
	if !validDigest(candidate.ReceiptDigest) ||
		candidate.ReceiptDigest != operationReceiptDigest(candidate) {
		return fmt.Errorf("operation receipt digest mismatch")
	}
	*receipt = candidate
	return nil
}
