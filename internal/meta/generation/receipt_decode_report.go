package generation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

func (report *ReceiptReport) UnmarshalJSON(data []byte) error {
	type wire ReceiptReport
	var decoded wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return fmt.Errorf("decode receipt report: %w", err)
	}
	if err := ensureIndicatorLedgerEOF(decoder); err != nil {
		return fmt.Errorf("decode receipt report: %w", err)
	}
	candidate := ReceiptReport(decoded)
	if candidate.SchemaVersion != ReceiptReportSchemaVersion {
		return fmt.Errorf("unsupported receipt report schema %q", candidate.SchemaVersion)
	}
	if !validReceiptLedgerProvenance(candidate.IndicatorDecisionLedgerDigest,
		candidate.IndicatorDecisionLedgerCount) {
		return fmt.Errorf("invalid receipt report indicator ledger provenance")
	}
	canonical := candidate
	canonical.InputDigest, canonical.ReportDigest, canonical.ReplayDigest = "", "", ""
	canonical = finishReceiptReport(canonical)
	if !reflect.DeepEqual(candidate, canonical) {
		return fmt.Errorf("receipt report canonical replay mismatch")
	}
	*report = candidate
	return nil
}

func validReceiptLedgerProvenance(digest string, count int) bool {
	if digest == "" {
		return count == 0
	}
	return count >= 0 && validIndicatorDecisionLedgerDigest(digest)
}
