package pressurecoverage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func validateInput(input Input) error {
	if input.Schema != SchemaVersion {
		return fmt.Errorf("pressure coverage input has invalid schema")
	}
	seen := make(map[string]struct{}, len(input.RequiredPressureIDs))
	for _, id := range input.RequiredPressureIDs {
		if !validID(id) {
			return fmt.Errorf("invalid required pressure ID %q", id)
		}
		if _, exists := seen[id]; exists {
			return fmt.Errorf("duplicate required pressure ID %q", id)
		}
		seen[id] = struct{}{}
	}
	records := make(map[string]PressureRecord, len(input.PressureRecords))
	for _, record := range input.PressureRecords {
		if !validID(record.PressureID) || !validID(record.CategoryID) ||
			!optionalID(record.IndependenceGroupID) || !optionalID(record.ApplicabilityRuleID) {
			return fmt.Errorf("invalid pressure record ID %q", record.PressureID)
		}
		if prior, exists := records[record.PressureID]; exists {
			if prior == record {
				return fmt.Errorf("duplicate pressure ID %q", record.PressureID)
			}
			return fmt.Errorf("conflicting pressure ID %q", record.PressureID)
		}
		records[record.PressureID] = record
	}
	return nil
}
func optionalID(value string) bool { return value == "" || validID(value) }
func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing JSON value")
	}
	return nil
}
