package selectiveci

import (
	"encoding/json"
)

func objectFields(raw json.RawMessage) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return nil, failure(ReasonInvalidInput, "required field must be an object")
	}
	return fields, nil
}
func requireFields(fields map[string]json.RawMessage, names ...string) error {
	for _, name := range names {
		if _, ok := fields[name]; !ok {
			return failure(ReasonInvalidInput, "missing required field "+name)
		}
	}
	return nil
}
func requireArrayObjectFields(raw json.RawMessage, names ...string) error {
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil || values == nil {
		return failure(ReasonInvalidInput, "required field must be an array")
	}
	for _, value := range values {
		fields, err := objectFields(value)
		if err != nil {
			return err
		}
		if err := requireFields(fields, names...); err != nil {
			return err
		}
	}
	return nil
}
