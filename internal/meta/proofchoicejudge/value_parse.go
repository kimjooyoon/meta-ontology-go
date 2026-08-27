package proofchoicejudge

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func decodeValue(raw string) (value, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return value{}, fmt.Errorf("SEMANTIC_VALUE_UNKNOWN")
	}
	for _, forbidden := range []string{"choice", "numerator", "denominator"} {
		if _, exists := fields[forbidden]; exists {
			return value{}, fmt.Errorf("SELF_DECLARED_PROOF_VALUE")
		}
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	var result value
	if err := decoder.Decode(&result); err != nil {
		return value{}, fmt.Errorf("SEMANTIC_VALUE_UNKNOWN")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return value{}, fmt.Errorf("SEMANTIC_VALUE_TRAILING")
	}
	if result.Schema != "gooo/proof-choice-input/v2" || result.ID == "" || result.Kind == "" {
		return value{}, fmt.Errorf("SEMANTIC_VALUE_METADATA_MISSING")
	}
	return result, nil
}
