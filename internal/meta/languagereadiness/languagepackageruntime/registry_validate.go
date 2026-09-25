package languagepackageruntime

import (
	"fmt"
	"slices"
)

func (registry Registry) Validate() error {
	if registry.Schema != RegistrySchema || registry.Version != RegistryVersion || len(registry.Cases) != FixedTotal {
		return fmt.Errorf("package runtime registry contract mismatch")
	}
	seen := map[string]bool{}
	positive, guardrails := 0, 0
	for _, definition := range registry.Cases {
		if definition.ID == "" || seen[definition.ID] || definition.MetaOperation == "" {
			return fmt.Errorf("invalid package runtime case %q", definition.ID)
		}
		seen[definition.ID] = true
		switch definition.Kind {
		case "POSITIVE":
			positive++
			if !knownAssertion(definition.Assertion) || definition.ProofChoice != "COHERENCE" {
				return fmt.Errorf("unknown positive case %q", definition.ID)
			}
		case "GUARDRAIL":
			guardrails++
			if !knownMutation(definition.Mutation) || definition.ExpectedCode == "" ||
				definition.ProofChoice != "REGRESSION" {
				return fmt.Errorf("unknown guardrail case %q", definition.ID)
			}
		default:
			return fmt.Errorf("unknown package runtime kind %q", definition.Kind)
		}
	}
	if positive != FixedPositive || guardrails != FixedGuardrails {
		return fmt.Errorf("package runtime case partition mismatch")
	}
	return nil
}

func knownAssertion(value string) bool {
	return contains([]string{"PACKAGE_GRAPH", "DIAMOND_ORDER", "MULTI_SOURCE", "ENTRY_CONTRACT",
		"PACKAGE_PERMUTATION", "IMPORT_PERMUTATION", "SOURCE_PERMUTATION", "CANONICAL_REPLAY",
		"SEMANTIC_BINDING", "ZERO_EFFECTS"}, value)
}

func knownMutation(value string) bool {
	return contains([]string{"UNKNOWN_SCHEMA", "DUPLICATE_PACKAGE", "UNKNOWN_IMPORT", "IMPORT_CYCLE",
		"HEADER_MISMATCH", "PARSE_ERROR", "UNKNOWN_ENTRY_PACKAGE", "UNKNOWN_ENTRY_ACTIVITY"}, value)
}

func contains(values []string, target string) bool {
	return slices.Contains(values, target)
}
