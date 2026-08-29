package verify

import "fmt"

func validateFixedPoint(program programDocument, verification verificationDocument) error {
	activities := []string{"ObserveCounterfactualBoundary", "PreserveRepositoryWorkspace",
		"ReplayCounterfactual", "TerminateAtFixedPoint"}
	for index, activity := range activities {
		if program.Steps[index].Activity != activity {
			return fmt.Errorf("fixed-point activity mismatch")
		}
	}
	if program.Selection.ProofChoice != "REGRESSION" ||
		program.Selection.Decision != "HOLD_FIXED_POINT" ||
		program.Selection.MetaOperation != "terminate-at-fixed-point" {
		return fmt.Errorf("fixed-point selection mismatch")
	}
	coverage := program.Coverage
	if coverage.BindingCount != canonicalBindingCount || coverage.ResolvedBindingCount != canonicalBindingCount ||
		coverage.RegistryOperationCount != 9 || coverage.ReferencedOperationCount != 9 ||
		!coverage.SelectionOperationResolved || coverage.Status != "COMPLETE" {
		return fmt.Errorf("program coverage mismatch")
	}
	return validateVerification(program, verification)
}
