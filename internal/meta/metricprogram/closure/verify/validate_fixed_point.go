package verify

import "fmt"

func validateFixedPoint(program programDocument, verification verificationDocument) error {
	activities := []string{"ObserveCounterfactualBoundary", "PreserveRepositoryWorkspace",
		"ReplayCounterfactual", "TerminateAtFixedPoint"}
	if program.Selection.MetaOperation == "preserve-non-promoting-terminal" {
		activities[3] = "PreserveNonPromotingTerminal"
	}
	for index, activity := range activities {
		if program.Steps[index].Activity != activity {
			return fmt.Errorf("fixed-point activity mismatch")
		}
	}
	if program.Selection.ProofChoice != "REGRESSION" {
		return fmt.Errorf("regression terminal selection mismatch")
	}
	if program.Selection.MetaOperation == "terminate-at-fixed-point" {
		if program.Selection.Decision != "HOLD_FIXED_POINT" {
			return fmt.Errorf("fixed-point selection mismatch")
		}
	} else if program.Selection.MetaOperation == "preserve-non-promoting-terminal" {
		if program.Selection.Decision != "PRESERVE_NON_PROMOTING_TERMINAL" ||
			program.Steps[3].OutputEntity != "NonPromotingTerminalReceipt" {
			return fmt.Errorf("non-promoting selection mismatch")
		}
	} else {
		return fmt.Errorf("unsupported terminal selection")
	}
	coverage := program.Coverage
	if coverage.BindingCount != canonicalBindingCount || coverage.ResolvedBindingCount != canonicalBindingCount ||
		coverage.RegistryOperationCount != 9 || coverage.ReferencedOperationCount != 9 ||
		!coverage.SelectionOperationResolved || coverage.Status != "COMPLETE" {
		return fmt.Errorf("program coverage mismatch")
	}
	return validateVerification(program, verification)
}
