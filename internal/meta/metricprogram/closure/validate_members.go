package closure

import "fmt"

func validateProgramMembers(program programDocument) error {
	if len(program.Operations) != 8 || len(program.Bindings) != 15 ||
		len(program.Steps) != 4 {
		return fmt.Errorf("program cardinality must be operations=8 bindings=15 steps=4")
	}
	operations := make(map[string]struct{}, len(program.Operations))
	for _, operation := range program.Operations {
		if operation.ID == "" || operation.Activity == "" ||
			operation.RepositoryWrites || operation.PromotionAuthorized {
			return fmt.Errorf("invalid operation %q", operation.ID)
		}
		if _, exists := operations[operation.ID]; exists {
			return fmt.Errorf("duplicate operation %q", operation.ID)
		}
		operations[operation.ID] = struct{}{}
	}
	for _, binding := range program.Bindings {
		if binding.IndicatorID == "" {
			return fmt.Errorf("binding indicator is empty")
		}
		if _, exists := operations[binding.OperationID]; !exists {
			return fmt.Errorf("binding operation %q is unknown", binding.OperationID)
		}
	}
	return validateProgramFixedPoint(program)
}
