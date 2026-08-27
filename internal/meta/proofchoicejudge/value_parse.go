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
	for _, forbidden := range []string{
		"choice", "numerator", "denominator", "evidence_kind", "admissible_routes",
		"observed", "predicate", "value", "slots", "observations", "dependencies",
		"provenance",
	} {
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
	if result.Schema != "gooo/proof-choice-input/v3" || result.ID == "" || result.Kind == "" {
		return value{}, fmt.Errorf("SEMANTIC_VALUE_METADATA_MISSING")
	}
	if result.Kind != "claim" && result.Kind != "metric" && result.Kind != "composition" {
		return value{}, fmt.Errorf("SEMANTIC_VALUE_KIND_UNKNOWN")
	}
	for _, route := range result.EvidenceCapabilities {
		if !validRoute(route) {
			return value{}, fmt.Errorf("EVIDENCE_CAPABILITY_UNKNOWN")
		}
	}
	return result, nil
}
