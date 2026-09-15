package closure

import "fmt"

func validateProgramFixedPoint(program programDocument) error {
	activities := expectedActivities()
	if program.Selection.MetaOperation == "preserve-non-promoting-terminal" {
		activities = []string{"ObserveCounterfactualBoundary", "PreserveRepositoryWorkspace",
			"ReplayCounterfactual", "PreserveNonPromotingTerminal"}
	}
	for index, activity := range activities {
		if program.Steps[index].Activity != activity {
			return fmt.Errorf("step %d activity %q does not match %q",
				index+1, program.Steps[index].Activity, activity)
		}
	}
	selection := program.Selection
	if selection.ProofChoice != "REGRESSION" {
		return fmt.Errorf("program selection is not a regression terminal")
	}
	if selection.MetaOperation == "terminate-at-fixed-point" {
		if selection.Decision != "HOLD_FIXED_POINT" {
			return fmt.Errorf("program selection is not the regression fixed point")
		}
	} else if selection.MetaOperation == "preserve-non-promoting-terminal" {
		if selection.Decision != "PRESERVE_NON_PROMOTING_TERMINAL" {
			return fmt.Errorf("program selection is not the non-promoting terminal")
		}
		if program.Steps[3].OutputEntity != "NonPromotingTerminalReceipt" {
			return fmt.Errorf("non-promoting terminal output is not distinct")
		}
	} else {
		return fmt.Errorf("program selection is unsupported")
	}
	coverage := program.Coverage
	if coverage.BindingCount != canonicalBindingCount || coverage.ResolvedBindingCount != canonicalBindingCount ||
		coverage.RegistryOperationCount != 9 || coverage.ReferencedOperationCount != 9 ||
		!coverage.SelectionOperationResolved || coverage.Status != "COMPLETE" {
		return fmt.Errorf("program coverage is incomplete")
	}
	return nil
}
