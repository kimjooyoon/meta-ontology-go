package bindingcoverage

import (
	"encoding/json"
	"fmt"
	"io"
)

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
		return fmt.Errorf("binding coverage input must be an object")
	}
	for _, name := range []string{"schema_version", "contract_id", "snapshot_digest", "expected_snapshot_digest", "required_bindings", "partitions", "precedence_registry"} {
		if _, ok := fields[name]; !ok {
			return fmt.Errorf("binding coverage input missing %q", name)
		}
	}
	return nil
}
func schemaFromJSON(data []byte) (string, bool) {
	if rejectDuplicateKeys(data) != nil {
		return "", false
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil || fields == nil {
		return "", false
	}
	var schema string
	if err := json.Unmarshal(fields["schema_version"], &schema); err != nil {
		return "", false
	}
	return schema, true
}
