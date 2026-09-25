package metainvocation

import "fmt"

func validateProgram(program Program) error {
	requiredEntities := map[string]string{
		"ChangeSet":           "gooo://meta/ci-plan/entity/change-set",
		"CheckPlan":           "gooo://meta/ci-plan/entity/check-plan",
		"VerificationReceipt": "gooo://meta/ci-plan/entity/verification-receipt",
	}
	for name, id := range requiredEntities {
		if program.Entities[name] != id {
			return fmt.Errorf("entity %q must bind id %q", name, id)
		}
	}
	requiredOperations := map[string]string{
		"SelectGoCheck":   operationGoRule,
		"SelectDocsCheck": operationDocsRule,
		"SelectYAMLCheck": operationYAMLRule,
		"PlanCI":          operationPlan,
		"VerifyCIPlan":    operationVerify,
	}
	for activity, operation := range requiredOperations {
		if program.Operations[activity].Program != operation {
			return fmt.Errorf("activity %q must bind operation %q", activity, operation)
		}
	}
	return nil
}
