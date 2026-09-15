package fullsoundness

import (
	"encoding/json"
	"fmt"
	"io"
)

func scanJSONArray(decoder *json.Decoder) error {
	for decoder.More() {
		if err := scanJSONValue(decoder); err != nil {
			return err
		}
	}
	_, err := decoder.Token()
	return err
}
func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple JSON values")
		}
		return err
	}
	return nil
}
func requireInputFields(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return fmt.Errorf("full soundness input must be an object")
	}
	for _, name := range inputFieldNames {
		if _, exists := fields[name]; !exists {
			return fmt.Errorf("full soundness input missing %q", name)
		}
	}
	return nil
}

var inputFieldNames = []string{
	"schema_version", "snapshot_digest", "policy_digest", "registry_digest", "selection_digest",
	"toolchain_digest", "runner_digest", "obligations", "commands", "impacted_obligation_ids",
	"selected_command_ids", "selection_receipt", "full_outcomes", "selected_outcomes",
	"full_resource_receipts", "selected_resource_receipts", "execution_authorized", "ci_authorized",
}
