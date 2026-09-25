package languagedeterministicquery

import "fmt"

func (registry PlanRegistry) Validate() error {
	if registry.Schema != RegistrySchema || registry.Version != RegistryVersion {
		return fmt.Errorf("invalid deterministic query registry identity")
	}
	if len(registry.Cases) != FixedTotal {
		return fmt.Errorf("deterministic query registry has %d cases, want %d", len(registry.Cases), FixedTotal)
	}
	seen := make(map[string]struct{}, len(registry.Cases))
	bindings, laws := 0, 0
	for _, definition := range registry.Cases {
		if err := validateDefinition(definition, seen); err != nil {
			return err
		}
		if definition.Kind == CaseBinding {
			bindings++
		} else {
			laws++
		}
	}
	if bindings != FixedBindingPlans || laws != FixedLawPlans {
		return fmt.Errorf("deterministic query registry partition is %d/%d", bindings, laws)
	}
	return nil
}

func validateDefinition(definition Definition, seen map[string]struct{}) error {
	if definition.ID == "" || definition.MetaOperation == "" {
		return fmt.Errorf("deterministic query case has empty identity")
	}
	if _, exists := seen[definition.ID]; exists {
		return fmt.Errorf("duplicate deterministic query case %q", definition.ID)
	}
	seen[definition.ID] = struct{}{}
	if definition.ProofChoice != "FOUNDATION" && definition.ProofChoice != "COHERENCE" && definition.ProofChoice != "REGRESSION" {
		return fmt.Errorf("case %q has unknown proof choice", definition.ID)
	}
	if definition.Kind == CaseBinding && (definition.BindingClass == "" || definition.Binding == "") {
		return fmt.Errorf("binding case %q is incomplete", definition.ID)
	}
	if definition.Kind != CaseBinding && definition.Kind != CaseLaw {
		return fmt.Errorf("case %q has unknown kind", definition.ID)
	}
	return nil
}
