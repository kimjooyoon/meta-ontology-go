package languagegointeroperation

import "fmt"

func (registry CaseRegistry) Validate() error {
	if registry.Schema != RegistrySchema || registry.Version != RegistryVersion {
		return fmt.Errorf("invalid Go interoperation registry identity")
	}
	if len(registry.Cases) != FixedTotal {
		return fmt.Errorf("Go interoperation registry has %d cases, want %d", len(registry.Cases), FixedTotal)
	}
	seen := make(map[string]struct{}, len(registry.Cases))
	counts := map[CaseKind]int{}
	for _, definition := range registry.Cases {
		if err := validateDefinition(definition, seen); err != nil {
			return err
		}
		counts[definition.Kind]++
	}
	if counts[CaseGenerator] != 8 || counts[CaseGo127] != 8 || counts[CaseGuardrail] != 8 {
		return fmt.Errorf("Go interoperation registry partition is %d/%d/%d", counts[CaseGenerator], counts[CaseGo127], counts[CaseGuardrail])
	}
	return nil
}

func validateDefinition(definition Definition, seen map[string]struct{}) error {
	if definition.ID == "" || definition.Fixture == "" || definition.MetaOperation == "" {
		return fmt.Errorf("Go interoperation case has empty identity")
	}
	if _, exists := seen[definition.ID]; exists {
		return fmt.Errorf("duplicate Go interoperation case %q", definition.ID)
	}
	seen[definition.ID] = struct{}{}
	if definition.ProofChoice != "COHERENCE" && definition.ProofChoice != "REGRESSION" {
		return fmt.Errorf("case %q has unknown proof choice", definition.ID)
	}
	if definition.Kind != CaseGenerator && definition.Kind != CaseGo127 && definition.Kind != CaseGuardrail {
		return fmt.Errorf("case %q has unknown kind", definition.ID)
	}
	if definition.ExpectedOutcome != "ACCEPT" && definition.ExpectedOutcome != "REJECT" {
		return fmt.Errorf("case %q has unknown expected outcome", definition.ID)
	}
	return nil
}
