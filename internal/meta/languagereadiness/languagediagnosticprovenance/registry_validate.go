package languagediagnosticprovenance

import "fmt"

func (registry CaseRegistry) Validate() error {
	if registry.Schema != RegistrySchema || registry.Version != RegistryVersion {
		return fmt.Errorf("diagnostic provenance registry identity mismatch")
	}
	if len(registry.Cases) != FixedTotal {
		return fmt.Errorf("diagnostic provenance registry total = %d", len(registry.Cases))
	}
	counts := map[CaseKind]int{}
	seen := map[string]bool{}
	for _, definition := range registry.Cases {
		if seen[definition.ID] || definition.ID == "" || definition.Fixture == "" {
			return fmt.Errorf("diagnostic provenance registry duplicate or empty case %q", definition.ID)
		}
		seen[definition.ID] = true
		counts[definition.Kind]++
		if err := validateDefinition(definition); err != nil {
			return err
		}
	}
	if counts[CaseSyntax] != FixedSyntaxCases || counts[CaseType] != FixedTypeCases ||
		counts[CaseSourceMap] != FixedSourceMapCases ||
		counts[CaseGuardrail] != FixedGuardrailCases {
		return fmt.Errorf("diagnostic provenance registry partition mismatch")
	}
	return nil
}

func validateDefinition(definition Definition) error {
	if definition.MetaOperation == "" || definition.ProofChoice == "" ||
		definition.ExpectedStage == "" {
		return fmt.Errorf("diagnostic provenance case %q has incomplete meta binding", definition.ID)
	}
	if definition.Kind == CaseGuardrail {
		if definition.ExpectedOutcome != "REJECT" || definition.ExpectedReason == "" ||
			definition.GuardrailClass == "" {
			return fmt.Errorf("diagnostic provenance guardrail %q is incomplete", definition.ID)
		}
		return nil
	}
	if definition.ExpectedOutcome != "TRACE" || definition.ExpectedReason != "" {
		return fmt.Errorf("diagnostic provenance positive case %q is incomplete", definition.ID)
	}
	return nil
}
