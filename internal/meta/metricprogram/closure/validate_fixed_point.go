package closure

import "fmt"

func validateProgramFixedPoint(program programDocument) error {
	for index, activity := range expectedActivities() {
		if program.Steps[index].Activity != activity {
			return fmt.Errorf("step %d activity %q does not match %q",
				index+1, program.Steps[index].Activity, activity)
		}
	}
	selection := program.Selection
	if selection.ProofChoice != "REGRESSION" ||
		selection.Decision != "HOLD_FIXED_POINT" ||
		selection.MetaOperation != "terminate-at-fixed-point" {
		return fmt.Errorf("program selection is not the regression fixed point")
	}
	coverage := program.Coverage
	if coverage.BindingCount != canonicalBindingCount || coverage.ResolvedBindingCount != canonicalBindingCount ||
		coverage.RegistryOperationCount != 9 || coverage.ReferencedOperationCount != 9 ||
		!coverage.SelectionOperationResolved || coverage.Status != "COMPLETE" {
		return fmt.Errorf("program coverage is incomplete")
	}
	return nil
}
