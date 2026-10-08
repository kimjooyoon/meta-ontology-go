package toolchainrelease

import (
	"fmt"
	"slices"
	"strings"
)

func validateJointFillRejection(r jointSmokeFillRejection, index int) error {
	id, stage, reason := "bad_cap_type", "TYPECHECK", "cannot use true"
	boundary, cap := "input.used >= input.limit", "true"
	if index == 2 {
		id, stage, reason = "bad_boundary_value", "TRAINING_EVALUATION", "zero divisor for integer operator /"
		boundary, cap = "(input.limit / (input.used - input.used)) > 0", "input.limit"
	}
	if r.ID != id || r.Stage != stage || !strings.Contains(r.Reason, reason) ||
		!slices.Equal(r.Holes, []jointSmokeFillHole{{"boundary", boundary}, {"cap", cap}}) {
		return fmt.Errorf("rejected-fill assignment, stage or reason differs")
	}
	return nil
}

func validateJointRejectedFill(a jointSmokeAttempt, index int, sourceSHA string) error {
	if a.Rejection == nil || a.Rejection.Stage != "LOCAL_SOURCE_FILL" || !jointSmokeInt(a.Rejection.Slot, 0) ||
		a.Rejection.Activity != "PlanBudget" || len(a.Candidates) != 0 || len(a.SearchCandidates) != 0 || len(a.FillCandidates) != 1 ||
		len(a.Masks) != 1 || a.Masks[0] != index || !jointSmokeInt(a.Passed, 0) || !jointSmokeInt(a.Total, 0) {
		return fmt.Errorf("rejected-fill local rejection or budget selector differs")
	}
	c := a.FillCandidates[0]
	if c.Schema != "gooo/fill-candidate/v1" || c.Activity != "PlanBudget" || c.ActivityID != "budgetplan://activity/plan-budget" ||
		c.InputSHA != sourceSHA || c.SelectedSHA == "" || c.PlanSHA == "" || !jointSmokeInt(c.Count, 5) ||
		c.Method != "caller_selected_assignment" || c.Rejection == nil || c.ID != c.Rejection.ID ||
		a.Rejection.CandidateID != c.ID || a.Rejection.Reason != c.Rejection.Reason || !slices.Equal(c.Holes, c.Rejection.Holes) ||
		!jointSmokeInt(c.Passed, 0) || !jointSmokeInt(c.Total, 0) || !jointSmokeInt(c.HoldoutPassed, 0) || !jointSmokeInt(c.HoldoutTotal, 0) ||
		len(c.Cases)+len(c.Values)+len(c.Holdout)+len(c.ValueHoldout) != 0 {
		return fmt.Errorf("rejected-fill identity or unscored fields differ")
	}
	if err := validateJointFillRejection(*c.Rejection, index); err != nil {
		return err
	}
	r := a.Runtime
	if r.Stage != "" || r.Failure != "" || r.SHA != "" || len(r.Traces) != 0 || len(r.Runs) != 0 ||
		!jointSmokeBool(r.Build.Started, false) || !jointSmokeInt(r.Calls, 0) || !jointSmokeInt(r.Passed, 0) || !jointSmokeInt(r.Total, 0) ||
		!jointSmokeBool(r.Projection, false) || !jointSmokeBool(r.Replay, false) {
		return fmt.Errorf("rejected-fill has native execution or score")
	}
	return nil
}
