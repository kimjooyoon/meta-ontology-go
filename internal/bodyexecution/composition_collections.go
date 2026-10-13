package bodyexecution

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kimjooyoon/meta-ontology-go/internal/bodycodegen"
)

func canonicalRecordField(raw []byte, scalar, cardinality string) (json.RawMessage, error) {
	if cardinality != "many" {
		return canonicalScalar(raw, scalar)
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '[' {
		return nil, fmt.Errorf("repeated field requires an array; null is not a collection")
	}
	var values []json.RawMessage
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	if len(values) > bodycodegen.RepeatedFieldLimit {
		return nil, fmt.Errorf("repeated field has %d values, limit %d", len(values), bodycodegen.RepeatedFieldLimit)
	}
	for i, value := range values {
		canonical, err := canonicalScalar(value, scalar)
		if err != nil {
			return nil, fmt.Errorf("element %d: %w", i, err)
		}
		values[i] = canonical
	}
	return json.Marshal(values)
}
