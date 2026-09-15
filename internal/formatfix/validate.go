package formatfix

import "fmt"

func Validate(plan Plan) error {
	if plan.Schema != PlanSchema || plan.PlanDigest == "" || plan.PlanDigest != planDigest(plan) {
		return fmt.Errorf("format/fix plan binding invalid")
	}
	if plan.Decision != DecisionChangePlanned && plan.Decision != DecisionFixedPoint &&
		plan.Decision != DecisionFailClosed {
		return fmt.Errorf("format/fix decision unknown")
	}
	if plan.Resolution != ResolutionExact && plan.Resolution != ResolutionLower {
		return fmt.Errorf("format/fix resolution unknown")
	}
	if plan.DirectWrites != 0 || plan.MutationAuthorized {
		return fmt.Errorf("format/fix direct effect invalid")
	}
	if plan.Resolution == ResolutionLower {
		if plan.Decision != DecisionFailClosed || plan.ReasonCode == "" ||
			plan.Changed || len(plan.Edits) != 0 {
			return fmt.Errorf("format/fix lower resolution invalid")
		}
		return nil
	}
	if !plan.SemanticEqual || plan.SemanticBefore == "" ||
		plan.SemanticBefore != plan.SemanticAfter || plan.ResultDigest == "" {
		return fmt.Errorf("format/fix semantic evidence incomplete")
	}
	switch plan.Decision {
	case DecisionFixedPoint:
		if plan.Changed || len(plan.Edits) != 0 || plan.SourceDigest != plan.ResultDigest ||
			plan.SourceBytes != plan.ResultBytes {
			return fmt.Errorf("format/fix fixed point invalid")
		}
	case DecisionChangePlanned:
		if !plan.Changed || len(plan.Edits) != 1 || plan.Edits[0].Start != 0 ||
			plan.Edits[0].End != plan.SourceBytes ||
			len(plan.Edits[0].Replacement) != plan.ResultBytes {
			return fmt.Errorf("format/fix change plan invalid")
		}
	default:
		return fmt.Errorf("format/fix exact decision invalid")
	}
	return nil
}
