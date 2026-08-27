package proofchoicealgebra

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func decodeValue(raw string) (Value, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return Value{}, fmt.Errorf("SEMANTIC_VALUE_UNKNOWN")
	}
	for _, forbidden := range []string{
		"choice", "numerator", "denominator", "evidence_kind", "admissible_routes",
		"observed", "predicate", "value", "slots", "observations", "dependencies",
		"provenance",
	} {
		if _, exists := fields[forbidden]; exists {
			return Value{}, fmt.Errorf("SELF_DECLARED_PROOF_VALUE")
		}
	}
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.DisallowUnknownFields()
	var value Value
	if err := decoder.Decode(&value); err != nil {
		return Value{}, fmt.Errorf("SEMANTIC_VALUE_UNKNOWN")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Value{}, fmt.Errorf("SEMANTIC_VALUE_TRAILING")
	}
	if value.Schema != InputSchema || value.ID == "" || value.Kind == "" {
		return Value{}, fmt.Errorf("SEMANTIC_VALUE_METADATA_MISSING")
	}
	if value.Kind != ClaimKind && value.Kind != MetricKind && value.Kind != CompositionKind {
		return Value{}, fmt.Errorf("SEMANTIC_VALUE_KIND_UNKNOWN")
	}
	for _, route := range value.EvidenceCapabilities {
		if !route.Valid() {
			return Value{}, fmt.Errorf("EVIDENCE_CAPABILITY_UNKNOWN")
		}
	}
	return value, nil
}
