package toolchainrelease

import (
	"encoding/json"
	"fmt"
)

func validateRelativeV2Header(r typedSmokeOutput, source, feedback []byte, replay bool, selected string) error {
	c, e := r.Construction, r.Evaluation
	declared, err := json.Marshal(c.ConstructionCases)
	if err != nil {
		return err
	}
	if !jointSmokeBool(r.Generated, !replay) || c.Schema != "gooo/joint-construction/v7" || c.Stage != "COMPLETE" || c.Decision != "COMPLETE_FINITE" ||
		c.StopReason != "LOCAL_AND_CALLER_CASES_MATCHED" || c.OriginalSourceSHA256 != scalarPreflightDigest(source) ||
		!samePackageSourceValue(declared, feedback) || c.ProgramBudget != 4 || c.SelectedAttempt != 0 || len(c.Attempts) != 1 ||
		len(c.Initial.Preparations) != 1 || c.Selected.GeneratedSHA256 == "" || !jointSmokeBool(e.Replayed, replay) || !jointSmokeInt(e.Calls, 0) ||
		e.Separation.UniqueInputs != 8 || e.Separation.DuplicateRows != 0 || e.Separation.ConstructionInputs != 0 || e.Separation.OtherInputs != 8 {
		return fmt.Errorf("V2 construction source, caller scope or reuse counters differ")
	}
	if replay && (selected == "" || selected != c.Selected.GeneratedSHA256) {
		return fmt.Errorf("V2 selected program changed")
	}
	if c.Selected.GeneratedSHA256 != scalarPreflightDigest([]byte(c.Selected.Source)) ||
		c.Selected.SelectedSourceSHA256 != scalarPreflightDigest([]byte(c.Selected.GoooSource)) {
		return fmt.Errorf("V2 selected source or projection digest differs")
	}
	return nil
}
