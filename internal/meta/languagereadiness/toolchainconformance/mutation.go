package toolchainconformance

import (
	"encoding/json"
	"fmt"
)

func applyTamper(artifacts map[string][]byte,
	definition TamperDefinition) (map[string][]byte, error) {
	mutated := make(map[string][]byte, len(artifacts)+1)
	for id, raw := range artifacts {
		mutated[id] = append([]byte(nil), raw...)
	}
	if definition.Mutation == "MISSING_SURFACE" {
		delete(mutated, definition.Target)
		return mutated, nil
	}
	raw, ok := mutated[definition.Target]
	if !ok {
		return nil, fmt.Errorf("tamper target %q is missing", definition.Target)
	}
	if definition.Mutation == "UNEXPECTED_SURFACE" {
		mutated["unexpected-surface"] = raw
		return mutated, nil
	}
	changed, err := mutateEnvelope(raw, definition.Mutation)
	if err != nil {
		return nil, err
	}
	mutated[definition.Target] = changed
	return mutated, nil
}

func mutateEnvelope(raw []byte, mutation string) ([]byte, error) {
	value := map[string]any{}
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, err
	}
	switch mutation {
	case "SCHEMA":
		value["schema"] = "gooo/unknown/v1"
	case "HEAD":
		mutateHead(value)
	case "DECISION":
		value["decision"] = "UNKNOWN"
	case "RESOLUTION":
		value["resolution"] = ResolutionLower
	case "UNRESOLVED":
		mutateSummary(value, "unresolved", 1)
	case "CASE":
		mutateSummary(value, "satisfied", 0)
	case "INDICATOR":
		mutateBooleanList(value, "indicators", "satisfied")
	case "PROOF":
		mutateBooleanList(value, "proofs", "passed")
	case "DIGEST":
		value["report_digest"] = "unknown"
	case "WRITE":
		value["repository_writes"] = 1
	case "MUTATION":
		value["mutation_authorized"] = true
	default:
		return nil, fmt.Errorf("unknown mutation %q", mutation)
	}
	return json.Marshal(value)
}
