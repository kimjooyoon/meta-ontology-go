package toolchainrelease

import (
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"
)

func validateJointFillRejectionIdentity(r jointSmokeOutput, source []byte, budget int, replay bool, selected string) error {
	c, e := r.Construction, r.Evaluation
	index, decision, stop := 4, "COMPLETE_FINITE", "LOCAL_AND_CALLER_CASES_MATCHED"
	if budget == 3 {
		index, decision, stop = 0, "PARTIAL_FINITE", "PROGRAM_BUDGET_EXHAUSTED"
	}
	if budget != 3 && budget != 5 || replay && budget != 5 || !jointSmokeBool(r.Generated, !replay) ||
		c.Schema != "gooo/joint-construction/v5" || c.Stage != "COMPLETE" || c.Failure != "" || c.Decision != decision || c.StopReason != stop ||
		!jointSmokeInt(c.Budget, budget) || c.Space != "5" || !jointSmokeInt(c.SelectedAttempt, index) || len(c.Attempts) != budget ||
		!slices.Equal(c.Kinds, []string{"source_fill_index"}) || c.Source == "" || c.Selected.SHA == "" ||
		c.OriginalSHA != fmt.Sprintf("sha256:%x", sha256.Sum256(source)) ||
		strings.Contains(c.Source, "__GOOO_BODY_HOLE_") || strings.Contains(c.Source, "assembling") ||
		c.Initial.Model.Loaded != nil && *c.Initial.Model.Loaded || c.Initial.FillModel.Loaded != nil && *c.Initial.FillModel.Loaded ||
		!jointSmokeBool(e.Replayed, replay) || !jointSmokeInt(e.Calls, 0) ||
		!jointSmokeInt(e.Separation.Unique, 4) || !jointSmokeInt(e.Separation.Duplicate, 0) ||
		!jointSmokeInt(e.Separation.Consumed, 0) || !jointSmokeInt(e.Separation.Other, 4) {
		return fmt.Errorf("rejected-fill identity, budget or input separation differs")
	}
	if replay && (selected == "" || selected != c.Selected.SHA) {
		return fmt.Errorf("rejected-fill replay changed the selected program")
	}
	return nil
}
